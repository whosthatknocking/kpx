package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteReadAndExpire(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dbPath := filepath.Join(t.TempDir(), "vault.kdbx")
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	if err := Write(dbPath, "hunter2", 10*time.Second, now); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	password, ok, err := Read(dbPath, now.Add(5*time.Second))
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
	if !ok {
		t.Fatalf("Read() did not find cached password")
	}
	if password != "hunter2" {
		t.Fatalf("Read() password = %q, want %q", password, "hunter2")
	}

	password, ok, err = Read(dbPath, now.Add(11*time.Second))
	if err != nil {
		t.Fatalf("Read() after expiry failed: %v", err)
	}
	if ok {
		t.Fatalf("Read() unexpectedly returned cached password %q after expiry", password)
	}
}

func TestDelete(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dbPath := filepath.Join(t.TempDir(), "vault.kdbx")
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	if err := Write(dbPath, "hunter2", 10*time.Second, now); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}
	if err := Delete(dbPath); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	_, ok, err := Read(dbPath, now)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
	if ok {
		t.Fatalf("Read() unexpectedly found deleted cache entry")
	}
}

func TestWriteCreatesSecureCacheFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dbPath := filepath.Join(t.TempDir(), "vault.kdbx")
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	if err := Write(dbPath, "hunter2", 10*time.Second, now); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	cachePath, err := path()
	if err != nil {
		t.Fatalf("path() failed: %v", err)
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("os.Stat() failed: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("cache mode = %#o, want %#o", got, 0o600)
	}
}
