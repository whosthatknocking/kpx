package vault

import (
	"path/filepath"
	"testing"
)

func TestVaultRoundTripAcrossRepeatedSaves(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	password := "hunter2"

	if err := Create(dbPath, CreateOptions{
		MasterPassword: password,
		DatabaseName:   "RoundTrip",
	}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	v, err := Open(dbPath, password)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if err := v.AddGroup("/Personal"); err != nil {
		t.Fatalf("AddGroup failed: %v", err)
	}
	if err := v.AddEntry("/Personal/GitHub", EntryInput{
		UserName: "alice",
		Password: "first-secret",
		URL:      "https://github.com",
		Notes:    "initial",
		CustomFields: map[string]string{
			"Environment": "prod",
		},
	}); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}

	if err := v.EditEntry("/Personal/GitHub", EntryPatch{
		URL: cliStringPtr("https://github.com/login"),
		SetCustomFields: map[string]string{
			"Team": "Platform",
		},
		ClearCustomFields: []string{"Environment"},
	}); err != nil {
		t.Fatalf("EditEntry after save failed: %v", err)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("second Save failed: %v", err)
	}

	reopened, err := Open(dbPath, password)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}

	entry, err := reopened.GetEntry("/Personal/GitHub")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	if entry.UserName != "alice" {
		t.Fatalf("username = %q, want %q", entry.UserName, "alice")
	}
	if entry.Password != "first-secret" {
		t.Fatalf("password = %q, want %q", entry.Password, "first-secret")
	}
	if entry.URL != "https://github.com/login" {
		t.Fatalf("url = %q, want %q", entry.URL, "https://github.com/login")
	}
	if entry.CustomFields["Team"] != "Platform" {
		t.Fatalf("custom Team = %q, want %q", entry.CustomFields["Team"], "Platform")
	}
	if _, ok := entry.CustomFields["Environment"]; ok {
		t.Fatalf("expected Environment custom field to be removed")
	}
}

func TestEditEntryRenameRejectsDuplicateTitle(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vault.kdbx")
	password := "hunter2"

	if err := Create(dbPath, CreateOptions{
		MasterPassword: password,
		DatabaseName:   "RenameCheck",
	}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	v, err := Open(dbPath, password)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if err := v.AddGroup("/Personal"); err != nil {
		t.Fatalf("AddGroup failed: %v", err)
	}
	if err := v.AddEntry("/Personal/GitHub", EntryInput{}); err != nil {
		t.Fatalf("AddEntry GitHub failed: %v", err)
	}
	if err := v.AddEntry("/Personal/GitLab", EntryInput{}); err != nil {
		t.Fatalf("AddEntry GitLab failed: %v", err)
	}

	err = v.EditEntry("/Personal/GitLab", EntryPatch{
		Title: cliStringPtr("GitHub"),
	})
	if err == nil {
		t.Fatal("expected duplicate title rename to fail")
	}
}

func cliStringPtr(value string) *string {
	return &value
}
