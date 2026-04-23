package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/whosthatknocking/kpx/internal/buildinfo"
)

func TestCLIFlow(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath, "--name", "Test Vault")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Created "+dbPath)

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Created group /Personal")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "ls", dbPath)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal")

	result = runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--username",
		"alice",
		"--password",
		"super-secret",
		"--url",
		"https://github.com",
		"--notes",
		"Personal account",
		"--field",
		"Environment=prod",
	)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Created entry /Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "ls", dbPath, "/Personal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", dbPath, "/Personal/GitHub")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Title: GitHub")
	result.requireStdoutContains(t, "UserName: alice")
	result.requireStdoutContains(t, "Password: [redacted]")
	result.requireStdoutContains(t, "Environment: prod")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", dbPath, "/Personal/GitHub", "--reveal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Password: super-secret")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "password", dbPath, "/Personal/GitHub")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "super-secret")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "find", dbPath, "git")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "find", dbPath, "personal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal/GitHub")

	result = runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"edit",
		dbPath,
		"/Personal/GitHub",
		"--url",
		"https://github.com/login",
		"--set-field",
		"Team=Platform",
		"--clear-field",
		"Environment",
	)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Updated entry /Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", dbPath, "/Personal/GitHub", "--reveal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "URL: https://github.com/login")
	result.requireStdoutContains(t, "Team: Platform")
	result.requireStdoutNotContains(t, "Environment: prod")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--no-input", "entry", "rm", dbPath, "/Personal/GitHub", "--force")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Deleted entry /Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "find", dbPath, "git")
	result.requireSuccess(t)
	result.requireStdoutNotContains(t, "/Personal/GitHub")
}

func TestDefaultDatabaseConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)

	configPath := configPathForTest(tempDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("default_database: "+dbPath+"\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", "/Personal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Created group /Personal")

	result = runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		"/Personal/GitHub",
		"--username",
		"alice",
		"--password",
		"super-secret",
	)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Created entry /Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", "/Personal/GitHub", "--reveal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Password: super-secret")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "password", "/Personal/GitHub")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "super-secret")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "find", "git")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal/GitHub")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "find", "personal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal/GitHub")

	if err := os.WriteFile(configPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "ls")
	if result.exitCode == 0 {
		t.Fatalf("expected group ls without explicit database or config default to fail\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	result.requireStderrContains(t, "database path not provided and no default database configured")
}

func TestDBCreateRefusesToOverwriteExistingDatabase(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath)
	if result.exitCode == 0 {
		t.Fatalf("expected db create on an existing database to fail\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	result.requireStderrContains(t, "database already exists")
}

func TestGroupAddRefusesToCreateExistingGroup(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal")
	if result.exitCode == 0 {
		t.Fatalf("expected group add on an existing group to fail\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	result.requireStderrContains(t, "group already exists")
}

func TestEntryShowUsesRevealConfigUnlessFlagOverrides(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--password",
		"super-secret",
	).requireSuccess(t)

	configPath := configPathForTest(tempDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("default_database: "+dbPath+"\nreveal: true\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", "/Personal/GitHub")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Password: super-secret")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", dbPath, "/Personal/GitHub")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Password: super-secret")

	if err := os.WriteFile(configPath, []byte("default_database: "+dbPath+"\nreveal: false\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", "/Personal/GitHub")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Password: [redacted]")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "show", "/Personal/GitHub", "--reveal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Password: super-secret")
}

func TestMasterPasswordCacheUsesConfiguredSeconds(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)

	configPath := configPathForTest(tempDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("master_password_cache_seconds: 60\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "ls", dbPath)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal")

	result = runKPX(t, tempDir, "", "--no-input", "group", "ls", dbPath)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal")
}

func TestMasterPasswordCacheWorksAcrossCommands(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	exportPath := filepath.Join(tempDir, "paper.txt")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath, "--name", "Test Vault").requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--username",
		"alice",
		"--password",
		"super-secret",
	).requireSuccess(t)

	configPath := configPathForTest(tempDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("master_password_cache_seconds: 60\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "validate", dbPath).requireSuccess(t)

	for _, args := range [][]string{
		{"--no-input", "db", "validate", dbPath},
		{"--no-input", "group", "ls", dbPath},
		{"--no-input", "group", "add", dbPath, "/Work"},
		{"--no-input", "entry", "ls", dbPath, "/Personal"},
		{"--no-input", "entry", "show", dbPath, "/Personal/GitHub"},
		{"--no-input", "find", dbPath, "GitHub"},
		{"--no-input", "entry", "add", dbPath, "/Personal/GitLab", "--username", "alice", "--password", "another-secret"},
		{"--no-input", "entry", "edit", dbPath, "/Personal/GitHub", "--notes", "Cached update"},
		{"--no-input", "entry", "rm", dbPath, "/Personal/GitLab", "--force"},
		{"--no-input", "export", "paper", dbPath, "--output", exportPath},
	} {
		result := runKPX(t, tempDir, "", args...)
		result.requireSuccess(t)
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	if !strings.Contains(string(data), "Password: super-secret") {
		t.Fatalf("paper export did not contain expected secret:\n%s", string(data))
	}
}

func TestSaveCreatesBackupWithDefaultFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("os.ReadDir() failed: %v", err)
	}

	foundBackup := false
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "vault.") && strings.HasSuffix(name, ".kdbx") && name != "vault.kdbx" {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Fatalf("expected backup file in %s, found entries: %v", tempDir, entryNames(entries))
	}
}

func TestSaveCreatesBackupInConfiguredDirectoryWithConfiguredName(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)

	configPath := configPathForTest(tempDir)
	backupDir := filepath.Join(tempDir, "backups")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	configData := "backup_directory: " + backupDir + "\nbackup_filename_format: \"{db_stem}-snapshot.{db_ext}\"\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)

	backupPath := filepath.Join(backupDir, "vault-snapshot.kdbx")
	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("os.Stat(%s) failed: %v", backupPath, err)
	}
}

func TestSaveUsesConfiguredDirectWriteMethod(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)

	configPath := configPathForTest(tempDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	configData := "save_method: direct_write\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Created group /Personal")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "ls", dbPath)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "/Personal")
}

func TestPaperExportWritesFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	outputPath := filepath.Join(tempDir, "paper.txt")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath, "--name", "Test Vault").requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--username",
		"alice",
		"--password",
		"super-secret",
		"--url",
		"https://github.com",
		"--notes",
		"Personal account",
		"--field",
		"Recovery Code=ABCD-EFGH",
	).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "export", "paper", dbPath, "--output", outputPath)
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Wrote paper export")

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	text := string(data)

	for _, want := range []string{
		"kpx Paper Backup",
		"Tool Version: " + buildinfo.BaseVersion(),
		"Database: Test Vault",
		"Source File: " + dbPath,
		"Path: /Personal/GitHub",
		"Password: super-secret",
		"Recovery Code: ABCD-EFGH",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("paper export did not contain %q\n%s", want, text)
		}
	}
}

func TestPaperExportRequiresExplicitDestination(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "export", "paper", dbPath)
	if result.exitCode == 0 {
		t.Fatalf("expected export without --output or --stdout to fail\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	result.requireStderrContains(t, "paper export requires --output or explicit --stdout")
}

func TestGroupListJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "group", "ls", dbPath)
	result.requireSuccess(t)
	var got struct {
		Groups []string `json:"groups"`
	}
	result.requireJSONEquals(t, &got)
	if len(got.Groups) != 1 || got.Groups[0] != "/Personal" {
		t.Fatalf("groups = %#v, want %#v", got.Groups, []string{"/Personal"})
	}
}

func TestEntryShowJSONRedactsPasswordUnlessReveal(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--username",
		"alice",
		"--password",
		"super-secret",
		"--field",
		"Environment=prod",
	).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "show", dbPath, "/Personal/GitHub")
	result.requireSuccess(t)
	var redacted struct {
		Entry struct {
			Password     string            `json:"password"`
			CustomFields map[string]string `json:"custom_fields"`
		} `json:"entry"`
	}
	result.requireJSONEquals(t, &redacted)
	if redacted.Entry.Password != "[redacted]" {
		t.Fatalf("password = %q, want %q", redacted.Entry.Password, "[redacted]")
	}
	if redacted.Entry.CustomFields["Environment"] != "prod" {
		t.Fatalf("custom_fields[Environment] = %q, want %q", redacted.Entry.CustomFields["Environment"], "prod")
	}

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "show", dbPath, "/Personal/GitHub", "--reveal")
	result.requireSuccess(t)
	var revealed struct {
		Entry struct {
			Password     string            `json:"password"`
			CustomFields map[string]string `json:"custom_fields"`
		} `json:"entry"`
	}
	result.requireJSONEquals(t, &revealed)
	if revealed.Entry.Password != "super-secret" {
		t.Fatalf("password = %q, want %q", revealed.Entry.Password, "super-secret")
	}
	if revealed.Entry.CustomFields["Environment"] != "prod" {
		t.Fatalf("custom_fields[Environment] = %q, want %q", revealed.Entry.CustomFields["Environment"], "prod")
	}
}

func TestVersionJSON(t *testing.T) {
	t.Parallel()

	result := runKPX(t, t.TempDir(), "", "--json", "version")
	result.requireSuccess(t)
	var cmdVersion struct {
		Version string `json:"version"`
	}
	result.requireJSONEquals(t, &cmdVersion)
	if cmdVersion.Version == "" {
		t.Fatal("version was empty")
	}

	result = runKPX(t, t.TempDir(), "", "--version", "--json")
	result.requireSuccess(t)
	var flagVersion struct {
		Version string `json:"version"`
	}
	result.requireJSONEquals(t, &flagVersion)
	if flagVersion.Version == "" {
		t.Fatal("version was empty")
	}
}

func TestHelpDoesNotExposeCompletionCommand(t *testing.T) {
	t.Parallel()

	result := runKPX(t, t.TempDir(), "", "--help")
	result.requireSuccess(t)
	result.requireStdoutNotContains(t, "  completion  ")
}

func TestEntryPasswordJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--password",
		"super-secret",
	).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "password", dbPath, "/Personal/GitHub")
	result.requireSuccess(t)
	var got struct {
		Path     string `json:"path"`
		Password string `json:"password"`
	}
	result.requireJSONEquals(t, &got)
	if got.Path != "/Personal/GitHub" {
		t.Fatalf("path = %q, want %q", got.Path, "/Personal/GitHub")
	}
	if got.Password != "super-secret" {
		t.Fatalf("password = %q, want %q", got.Password, "super-secret")
	}
}

func TestJSONStatusContracts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	exportPath := filepath.Join(tempDir, "paper.txt")

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "db", "create", dbPath, "--name", "Test Vault")
	result.requireSuccess(t)
	result.requireStatusJSON(t, "created", "database", dbPath, "", "")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "db", "validate", dbPath)
	result.requireSuccess(t)
	result.requireStatusJSON(t, "validated", "database", dbPath, "", "")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "group", "add", dbPath, "/Personal")
	result.requireSuccess(t)
	result.requireStatusJSON(t, "created", "group", "/Personal", "", "")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "add", dbPath, "/Personal/GitHub", "--password", "super-secret")
	result.requireSuccess(t)
	result.requireStatusJSON(t, "created", "entry", "/Personal/GitHub", "", "")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "edit", dbPath, "/Personal/GitHub", "--notes", "Updated")
	result.requireSuccess(t)
	result.requireStatusJSON(t, "updated", "entry", "/Personal/GitHub", "", "")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "export", "paper", dbPath, "--output", exportPath)
	result.requireSuccess(t)
	result.requireStatusJSON(t, "exported", "database", "", exportPath, "paper")

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "rm", dbPath, "/Personal/GitHub", "--force")
	result.requireSuccess(t)
	result.requireStatusJSON(t, "deleted", "entry", "/Personal/GitHub", "", "")
}

func TestEntryListAndFindJSONContracts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "entry", "add", dbPath, "/Personal/GitHub", "--password", "super-secret").requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "ls", dbPath, "/Personal")
	result.requireSuccess(t)
	var list struct {
		Group   string   `json:"group"`
		Entries []string `json:"entries"`
	}
	result.requireJSONEquals(t, &list)
	if list.Group != "/Personal" {
		t.Fatalf("group = %q, want %q", list.Group, "/Personal")
	}
	if len(list.Entries) != 1 || list.Entries[0] != "/Personal/GitHub" {
		t.Fatalf("entries = %#v, want %#v", list.Entries, []string{"/Personal/GitHub"})
	}

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "find", dbPath, "github")
	result.requireSuccess(t)
	var find struct {
		Query   string   `json:"query"`
		Exact   bool     `json:"exact"`
		Results []string `json:"results"`
	}
	result.requireJSONEquals(t, &find)
	if find.Query != "github" || find.Exact {
		t.Fatalf("find = %#v, want query github exact false", find)
	}
	if len(find.Results) != 1 || find.Results[0] != "/Personal/GitHub" {
		t.Fatalf("results = %#v, want %#v", find.Results, []string{"/Personal/GitHub"})
	}
}

func TestEntryRemoveRequiresForceWithNoInput(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)
	runKPX(
		t,
		tempDir,
		"hunter2\n",
		"--master-password-stdin",
		"entry",
		"add",
		dbPath,
		"/Personal/GitHub",
		"--password",
		"super-secret",
	).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--no-input", "entry", "rm", dbPath, "/Personal/GitHub")
	if result.exitCode == 0 {
		t.Fatalf("expected delete without --force to fail under --no-input\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	result.requireStderrContains(t, "delete requires --force when --no-input is set")
}

func TestVersionCommand(t *testing.T) {
	t.Parallel()

	result := runKPX(t, t.TempDir(), "", "version")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "kpx "+buildinfo.String())
}

func TestVersionFlag(t *testing.T) {
	t.Parallel()

	result := runKPX(t, t.TempDir(), "", "--version")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "kpx "+buildinfo.String())
}

func TestBaseVersionIsEmbedded(t *testing.T) {
	t.Parallel()

	if got := buildinfo.BaseVersion(); got == "" {
		t.Fatal("BaseVersion() was empty")
	}
}

func TestHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i := range args {
		if args[i] == "--" {
			os.Args = append([]string{args[0]}, args[i+1:]...)
			main()
			return
		}
	}

	os.Exit(2)
}

type commandResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func runKPX(t *testing.T, dir string, stdin string, args ...string) commandResult {
	t.Helper()

	cmdArgs := append([]string{"-test.run=TestHelperProcess", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Dir = dir
	cmd.Env = append(
		os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"HOME="+dir,
		"XDG_CONFIG_HOME="+filepath.Join(dir, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(dir, ".cache"),
	)
	cmd.Stdin = strings.NewReader(stdin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := commandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
	}

	if err == nil {
		return result
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("command failed to run: %v", err)
	}
	result.exitCode = exitErr.ExitCode()
	return result
}

func configPathForTest(dir string) string {
	return filepath.Join(dir, ".config", "kpx", "config.yml")
}

func (r commandResult) requireSuccess(t *testing.T) {
	t.Helper()
	if r.exitCode != 0 {
		t.Fatalf("expected success, got exit code %d\nstdout:\n%s\nstderr:\n%s", r.exitCode, r.stdout, r.stderr)
	}
}

func (r commandResult) requireStdoutContains(t *testing.T, want string) {
	t.Helper()
	if !strings.Contains(r.stdout, want) {
		t.Fatalf("stdout did not contain %q\nstdout:\n%s\nstderr:\n%s", want, r.stdout, r.stderr)
	}
}

func (r commandResult) requireStdoutNotContains(t *testing.T, unwanted string) {
	t.Helper()
	if strings.Contains(r.stdout, unwanted) {
		t.Fatalf("stdout unexpectedly contained %q\nstdout:\n%s\nstderr:\n%s", unwanted, r.stdout, r.stderr)
	}
}

func (r commandResult) requireStderrContains(t *testing.T, want string) {
	t.Helper()
	if !strings.Contains(r.stderr, want) {
		t.Fatalf("stderr did not contain %q\nstdout:\n%s\nstderr:\n%s", want, r.stdout, r.stderr)
	}
}

func (r commandResult) requireJSONEquals(t *testing.T, target any) {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(r.stdout))
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("stdout was not valid JSON: %v\nstdout:\n%s\nstderr:\n%s", err, r.stdout, r.stderr)
	}
}

func (r commandResult) requireStatusJSON(t *testing.T, status string, kind string, path string, output string, format string) {
	t.Helper()
	var got struct {
		Status string `json:"status"`
		Kind   string `json:"kind"`
		Path   string `json:"path,omitempty"`
		Output string `json:"output,omitempty"`
		Format string `json:"format,omitempty"`
	}
	r.requireJSONEquals(t, &got)
	if got.Status != status || got.Kind != kind || got.Path != path || got.Output != output || got.Format != format {
		t.Fatalf("status json = %#v, want status=%q kind=%q path=%q output=%q format=%q", got, status, kind, path, output, format)
	}
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}
