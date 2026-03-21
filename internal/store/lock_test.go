//go:build unix

package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExclusiveLockBlocksSecondWriter(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "vault.kdbx")

	first, err := LockExclusive(path)
	if err != nil {
		t.Fatalf("LockExclusive() failed: %v", err)
	}
	defer first.Close()

	acquired := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		second, err := LockExclusive(path)
		if err == nil {
			close(acquired)
			err = second.Close()
		}
		done <- err
	}()

	select {
	case <-acquired:
		t.Fatal("second writer lock acquired before first lock was released")
	case <-time.After(100 * time.Millisecond):
	}

	if err := first.Close(); err != nil {
		t.Fatalf("first.Close() failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("second lock failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("second writer lock did not acquire after release")
	}
}

func TestSharedLockBlocksWriter(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "vault.kdbx")

	reader, err := LockShared(path)
	if err != nil {
		t.Fatalf("LockShared() failed: %v", err)
	}
	defer reader.Close()

	acquired := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		writer, err := LockExclusive(path)
		if err == nil {
			close(acquired)
			err = writer.Close()
		}
		done <- err
	}()

	select {
	case <-acquired:
		t.Fatal("writer lock acquired while shared lock was still held")
	case <-time.After(100 * time.Millisecond):
	}

	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("writer lock failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("writer lock did not acquire after shared lock was released")
	}
}

func TestLockFileIsRemovedAfterClose(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "vault.kdbx")
	lockPath := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".lock")

	lock, err := LockExclusive(path)
	if err != nil {
		t.Fatalf("LockExclusive() failed: %v", err)
	}

	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("expected lock file to exist while lock is held: %v", err)
	}

	if err := lock.Close(); err != nil {
		t.Fatalf("lock.Close() failed: %v", err)
	}

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("expected lock file to be removed after close, stat err = %v", err)
	}
}
