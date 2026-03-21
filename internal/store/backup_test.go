package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupFileUsesDefaultLocationAndFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	if err := os.WriteFile(dbPath, []byte("vault-data"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	now := time.Date(2026, 3, 21, 12, 34, 56, 0, time.UTC)
	if err := BackupFile(dbPath, BackupOptions{Now: now}); err != nil {
		t.Fatalf("BackupFile() failed: %v", err)
	}

	backupPath := filepath.Join(tempDir, "vault.20260321T123456Z.kdbx")
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	if string(data) != "vault-data" {
		t.Fatalf("backup content = %q, want %q", string(data), "vault-data")
	}
}

func TestBackupFileUsesCustomDestinationAndFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	if err := os.WriteFile(dbPath, []byte("vault-data"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	backupDir := filepath.Join(tempDir, "backups")
	now := time.Date(2026, 3, 21, 12, 34, 56, 0, time.UTC)
	if err := BackupFile(dbPath, BackupOptions{
		DestinationDir: backupDir,
		FilenameFormat: "{db_stem}-{timestamp}.{db_ext}",
		Now:            now,
	}); err != nil {
		t.Fatalf("BackupFile() failed: %v", err)
	}

	backupPath := filepath.Join(backupDir, "vault-20260321T123456Z.kdbx")
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	if string(data) != "vault-data" {
		t.Fatalf("backup content = %q, want %q", string(data), "vault-data")
	}
}
