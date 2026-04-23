package cache

import (
	"os"
	"path/filepath"
	"strings"
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

	cachePath, err := path()
	if err != nil {
		t.Fatalf("path() failed: %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("expected expired cache file to be removed, stat err = %v", err)
	}
}

func TestPathUsesXDGCacheHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	xdgCacheHome := filepath.Join(t.TempDir(), "xdg-cache")
	t.Setenv("XDG_CACHE_HOME", xdgCacheHome)

	got, err := path()
	if err != nil {
		t.Fatalf("path() failed: %v", err)
	}

	want := filepath.Join(xdgCacheHome, "kpx", "master-password-cache.yml")
	if got != want {
		t.Fatalf("path() = %q, want %q", got, want)
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

	cachePath, err := path()
	if err != nil {
		t.Fatalf("path() failed: %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("expected cache file to be removed after delete, stat err = %v", err)
	}
}

func TestReadOnlyRemovesExpiredEntryForRequestedDatabase(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dbPathA := filepath.Join(t.TempDir(), "vault-a.kdbx")
	dbPathB := filepath.Join(t.TempDir(), "vault-b.kdbx")
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	if err := Write(dbPathA, "expired-password", 10*time.Second, now); err != nil {
		t.Fatalf("Write(dbPathA) failed: %v", err)
	}
	if err := Write(dbPathB, "valid-password", 60*time.Second, now); err != nil {
		t.Fatalf("Write(dbPathB) failed: %v", err)
	}

	if _, ok, err := Read(dbPathA, now.Add(11*time.Second)); err != nil {
		t.Fatalf("Read(dbPathA) failed: %v", err)
	} else if ok {
		t.Fatalf("Read(dbPathA) unexpectedly found expired cache entry")
	}

	password, ok, err := Read(dbPathB, now.Add(11*time.Second))
	if err != nil {
		t.Fatalf("Read(dbPathB) failed: %v", err)
	}
	if !ok {
		t.Fatalf("Read(dbPathB) did not find valid cache entry")
	}
	if password != "valid-password" {
		t.Fatalf("Read(dbPathB) password = %q, want %q", password, "valid-password")
	}
}

func TestReadRemovesOtherExpiredEntriesWhileKeepingValidOnes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dbPathA := filepath.Join(t.TempDir(), "vault-a.kdbx")
	dbPathB := filepath.Join(t.TempDir(), "vault-b.kdbx")
	dbPathC := filepath.Join(t.TempDir(), "vault-c.kdbx")
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	checkTime := now.Add(11 * time.Second)

	if err := Write(dbPathA, "expired-a", 10*time.Second, now); err != nil {
		t.Fatalf("Write(dbPathA) failed: %v", err)
	}
	if err := Write(dbPathB, "valid-b", 60*time.Second, now); err != nil {
		t.Fatalf("Write(dbPathB) failed: %v", err)
	}
	if err := Write(dbPathC, "expired-c", 5*time.Second, now); err != nil {
		t.Fatalf("Write(dbPathC) failed: %v", err)
	}

	password, ok, err := Read(dbPathB, checkTime)
	if err != nil {
		t.Fatalf("Read(dbPathB) failed: %v", err)
	}
	if !ok {
		t.Fatalf("Read(dbPathB) did not find valid cache entry")
	}
	if password != "valid-b" {
		t.Fatalf("Read(dbPathB) password = %q, want %q", password, "valid-b")
	}

	if _, ok, err := Read(dbPathA, checkTime); err != nil {
		t.Fatalf("Read(dbPathA) failed: %v", err)
	} else if ok {
		t.Fatalf("Read(dbPathA) unexpectedly found expired cache entry")
	}
	if _, ok, err := Read(dbPathC, checkTime); err != nil {
		t.Fatalf("Read(dbPathC) failed: %v", err)
	} else if ok {
		t.Fatalf("Read(dbPathC) unexpectedly found expired cache entry")
	}

	cachePath, err := path()
	if err != nil {
		t.Fatalf("path() failed: %v", err)
	}
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "valid-b") {
		t.Fatalf("expected valid cache entry to remain in cache file:\n%s", text)
	}
	if strings.Contains(text, "expired-a") || strings.Contains(text, "expired-c") {
		t.Fatalf("expected expired cache entries to be removed from cache file:\n%s", text)
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
