package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesAndReplacesContent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "vault.kdbx")

	if err := WriteFileAtomic(path, []byte("first")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after first write failed: %v", err)
	}
	if string(data) != "first" {
		t.Fatalf("first content = %q, want %q", string(data), "first")
	}

	if err := WriteFileAtomic(path, []byte("second")); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after second write failed: %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("second content = %q, want %q", string(data), "second")
	}
}

func TestWriteFileAtomicPreservesPermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "vault.kdbx")

	if err := os.WriteFile(path, []byte("seed"), 0o640); err != nil {
		t.Fatalf("seed file write failed: %v", err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}

	if err := WriteFileAtomic(path, []byte("updated")); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("mode = %#o, want %#o", got, 0o640)
	}
}

func TestWriteFileDirectCreatesAndReplacesContent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "vault.kdbx")

	if err := WriteFileDirect(path, []byte("first")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after first write failed: %v", err)
	}
	if string(data) != "first" {
		t.Fatalf("first content = %q, want %q", string(data), "first")
	}

	if err := WriteFileDirect(path, []byte("second")); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after second write failed: %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("second content = %q, want %q", string(data), "second")
	}
}

func TestWriteFileDirectPreservesPermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "vault.kdbx")

	if err := os.WriteFile(path, []byte("seed"), 0o640); err != nil {
		t.Fatalf("seed file write failed: %v", err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}

	if err := WriteFileDirect(path, []byte("updated")); err != nil {
		t.Fatalf("WriteFileDirect failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("mode = %#o, want %#o", got, 0o640)
	}
}

func TestWriteFileRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "vault.kdbx")

	err := WriteFile(path, []byte("data"), SaveOptions{Method: "unknown"})
	if err == nil {
		t.Fatalf("WriteFile() unexpectedly succeeded with unsupported method")
	}
}
