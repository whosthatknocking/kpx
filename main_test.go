package main

import (
	"bytes"
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

	configPath := filepath.Join(tempDir, ".kpx", "config.yml")
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

	configPath := filepath.Join(tempDir, ".kpx", "config.yml")
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

	configPath := filepath.Join(tempDir, ".kpx", "config.yml")
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

	configPath := filepath.Join(tempDir, ".kpx", "config.yml")
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
		{"--no-input", "export", "paper", dbPath, "--output", exportPath, "--force"},
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

	configPath := filepath.Join(tempDir, ".kpx", "config.yml")
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

	configPath := filepath.Join(tempDir, ".kpx", "config.yml")
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

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "export", "paper", dbPath, "--output", outputPath, "--force")
	result.requireSuccess(t)
	result.requireStdoutContains(t, "Wrote paper export")

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	text := string(data)

	for _, want := range []string{
		"kpx Paper Backup",
		"Tool Version: 0.1.8",
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

func TestPaperExportRequiresForceWithNoInput(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	outputPath := filepath.Join(tempDir, "paper.txt")
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--no-input", "export", "paper", dbPath, "--output", outputPath)
	if result.exitCode == 0 {
		t.Fatalf("expected export under --no-input without --force to fail\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
	}
	result.requireStderrContains(t, "paper export requires --force when --no-input is set")
}

func TestGroupListJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")

	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "db", "create", dbPath).requireSuccess(t)
	runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "group", "add", dbPath, "/Personal").requireSuccess(t)

	result := runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "group", "ls", dbPath)
	result.requireSuccess(t)
	result.requireStdoutContains(t, `"groups": [`)
	result.requireStdoutContains(t, `"/Personal"`)
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
	result.requireStdoutContains(t, `"password": "[redacted]"`)
	result.requireStdoutContains(t, `"Environment": "prod"`)

	result = runKPX(t, tempDir, "hunter2\n", "--master-password-stdin", "--json", "entry", "show", dbPath, "/Personal/GitHub", "--reveal")
	result.requireSuccess(t)
	result.requireStdoutContains(t, `"password": "super-secret"`)
	result.requireStdoutContains(t, `"Environment": "prod"`)
}

func TestVersionJSON(t *testing.T) {
	t.Parallel()

	result := runKPX(t, t.TempDir(), "", "--json", "version")
	result.requireSuccess(t)
	result.requireStdoutContains(t, `"version": "`)

	result = runKPX(t, t.TempDir(), "", "--version", "--json")
	result.requireSuccess(t)
	result.requireStdoutContains(t, `"version": "`)
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
	result.requireStdoutContains(t, `"password": "super-secret"`)
	result.requireStdoutContains(t, `"path": "/Personal/GitHub"`)
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

	if got := buildinfo.BaseVersion(); got != "0.1.8" {
		t.Fatalf("BaseVersion() = %q, want %q", got, "0.1.8")
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
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HOME="+dir)
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

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}
