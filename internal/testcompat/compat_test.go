package testcompat

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/whosthatknocking/kpx/internal/config"
	"github.com/whosthatknocking/kpx/internal/vault"
)

type fixtureManifest struct {
	Fixtures []fixtureSpec `json:"fixtures"`
}

type fixtureSpec struct {
	Name         string          `json:"name"`
	Source       string          `json:"source"`
	Path         string          `json:"path,omitempty"`
	URL          string          `json:"url,omitempty"`
	SHA256       string          `json:"sha256,omitempty"`
	Password     string          `json:"password"`
	DatabaseName string          `json:"database_name"`
	Groups       []string        `json:"groups"`
	Entries      []fixtureEntry  `json:"entries"`
	Searches     []fixtureSearch `json:"searches,omitempty"`
}

type fixtureEntry struct {
	Path         string            `json:"path"`
	UserName     string            `json:"username"`
	Password     string            `json:"password"`
	URL          string            `json:"url"`
	Notes        string            `json:"notes"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

type fixtureSearch struct {
	Query  string   `json:"query"`
	Exact  bool     `json:"exact"`
	Expect []string `json:"expect"`
}

func TestCompatibilityFixtures(t *testing.T) {
	manifest := loadManifest(t)
	if len(manifest.Fixtures) == 0 {
		t.Fatal("fixture manifest is empty")
	}

	includeRemoteFixtures := includeRemoteFixtures()

	for _, fixture := range manifest.Fixtures {
		fixture := fixture
		t.Run(fixture.Name, func(t *testing.T) {
			if fixture.Source == "url" && !includeRemoteFixtures {
				t.Skip("set KPX_REMOTE_FIXTURES=1 to fetch official remote compatibility fixtures")
			}

			for _, tc := range []struct {
				name         string
				saveMethod   string
				enableBackup bool
			}{
				{name: "temporary-file", saveMethod: "temporary_file"},
				{name: "direct-write", saveMethod: "direct_write"},
				{name: "temporary-file-with-backups", saveMethod: "temporary_file", enableBackup: true},
			} {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					homeDir := t.TempDir()
					t.Setenv("HOME", homeDir)

					dbPath := prepareFixture(t, fixture)

					cfg := config.File{
						SaveMethod: tc.saveMethod,
					}
					if tc.enableBackup {
						cfg.BackupDirectory = filepath.Join(homeDir, "backups")
						cfg.BackupFilenameFormat = "{db_stem}.{timestamp}.{db_ext}"
					}
					if err := config.Save(cfg); err != nil {
						t.Fatalf("config.Save() failed: %v", err)
					}

					v := openAndAssertFixture(t, dbPath, fixture)

					for i := 0; i < 3; i++ {
						if err := v.Save(); err != nil {
							t.Fatalf("Save cycle %d failed: %v", i+1, err)
						}
						v = openAndAssertFixture(t, dbPath, fixture)
					}

					if tc.enableBackup {
						backupPath := newestBackup(t, cfg.BackupDirectory)
						backupVault := openAndAssertFixture(t, backupPath, fixture)
						if err := backupVault.Save(); err != nil {
							t.Fatalf("backup Save failed: %v", err)
						}
						openAndAssertFixture(t, backupPath, fixture)
					}
				})
			}
		})
	}
}

func openAndAssertFixture(t *testing.T, dbPath string, fixture fixtureSpec) *vault.Vault {
	t.Helper()

	v, err := vault.Open(dbPath, fixture.Password)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	gotGroups := v.ListGroups()
	if !slices.Equal(gotGroups, fixture.Groups) {
		t.Fatalf("groups = %#v, want %#v", gotGroups, fixture.Groups)
	}

	for _, entrySpec := range fixture.Entries {
		entry, err := v.GetEntry(entrySpec.Path)
		if err != nil {
			t.Fatalf("GetEntry(%q) failed: %v", entrySpec.Path, err)
		}

		if entry.Path != entrySpec.Path {
			t.Fatalf("entry path = %q, want %q", entry.Path, entrySpec.Path)
		}
		if entry.UserName != entrySpec.UserName {
			t.Fatalf("entry username for %s = %q, want %q", entrySpec.Path, entry.UserName, entrySpec.UserName)
		}
		if entry.Password != entrySpec.Password {
			t.Fatalf("entry password for %s = %q, want %q", entrySpec.Path, entry.Password, entrySpec.Password)
		}
		if entry.URL != entrySpec.URL {
			t.Fatalf("entry URL for %s = %q, want %q", entrySpec.Path, entry.URL, entrySpec.URL)
		}
		if entry.Notes != entrySpec.Notes {
			t.Fatalf("entry notes for %s = %q, want %q", entrySpec.Path, entry.Notes, entrySpec.Notes)
		}
		if !mapsEqual(entry.CustomFields, entrySpec.CustomFields) {
			t.Fatalf("entry custom fields for %s = %#v, want %#v", entrySpec.Path, entry.CustomFields, entrySpec.CustomFields)
		}
	}

	for _, searchSpec := range fixture.Searches {
		results := v.FindEntries(searchSpec.Query, searchSpec.Exact)
		gotPaths := make([]string, 0, len(results))
		for _, result := range results {
			gotPaths = append(gotPaths, result.Path)
		}
		if !slices.Equal(gotPaths, searchSpec.Expect) {
			t.Fatalf("search(%q, exact=%v) = %#v, want %#v", searchSpec.Query, searchSpec.Exact, gotPaths, searchSpec.Expect)
		}
	}

	return v
}

func loadManifest(t *testing.T) fixtureManifest {
	t.Helper()

	path := filepath.Join("testdata", "fixtures", "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", path, err)
	}

	var manifest fixtureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("json.Unmarshal(%s) failed: %v", path, err)
	}
	return manifest
}

func prepareFixture(t *testing.T, spec fixtureSpec) string {
	t.Helper()

	target := filepath.Join(t.TempDir(), spec.Name+".kdbx")

	switch spec.Source {
	case "generated":
		if err := generateFixture(target, spec); err != nil {
			t.Fatalf("generateFixture(%s) failed: %v", spec.Name, err)
		}
	case "file":
		source := filepath.Join("testdata", "fixtures", spec.Path)
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatalf("ReadFile(%s) failed: %v", source, err)
		}
		if err := os.WriteFile(target, data, 0o600); err != nil {
			t.Fatalf("WriteFile(%s) failed: %v", target, err)
		}
	case "url":
		data, err := fetchFixture(spec)
		if err != nil {
			t.Fatalf("fetchFixture(%s) failed: %v", spec.Name, err)
		}
		if err := os.WriteFile(target, data, 0o600); err != nil {
			t.Fatalf("WriteFile(%s) failed: %v", target, err)
		}
	default:
		t.Fatalf("unsupported fixture source %q", spec.Source)
	}

	return target
}

func fetchFixture(spec fixtureSpec) ([]byte, error) {
	if spec.URL == "" {
		return nil, fmt.Errorf("fixture URL is required for source=url")
	}

	client := &http.Client{Timeout: 20 * time.Second}

	request, err := http.NewRequest(http.MethodGet, spec.URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "kpx-testcompat/1")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if spec.SHA256 != "" {
		sum := sha256.Sum256(data)
		got := hex.EncodeToString(sum[:])
		if got != spec.SHA256 {
			return nil, fmt.Errorf("sha256 mismatch: got %s want %s", got, spec.SHA256)
		}
	}

	return data, nil
}

func includeRemoteFixtures() bool {
	value := os.Getenv("KPX_REMOTE_FIXTURES")
	return value == "1" || value == "true" || value == "TRUE"
}

func generateFixture(path string, spec fixtureSpec) error {
	if err := vault.Create(path, vault.CreateOptions{
		MasterPassword: spec.Password,
		DatabaseName:   spec.DatabaseName,
	}); err != nil {
		return err
	}

	v, err := vault.Open(path, spec.Password)
	if err != nil {
		return err
	}

	for _, groupPath := range spec.Groups {
		if err := v.AddGroup(groupPath); err != nil {
			return err
		}
	}

	for _, entry := range spec.Entries {
		if err := v.AddEntry(entry.Path, vault.EntryInput{
			UserName:     entry.UserName,
			Password:     entry.Password,
			URL:          entry.URL,
			Notes:        entry.Notes,
			CustomFields: entry.CustomFields,
		}); err != nil {
			return err
		}
	}

	return v.Save()
}

func mapsEqual(left map[string]string, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func newestBackup(t *testing.T, dir string) string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir(%s) failed: %v", dir, err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one backup in %s", dir)
	}

	var newest string
	var newestInfo os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("entry.Info(%s) failed: %v", entry.Name(), err)
		}
		if newest == "" || info.ModTime().After(newestInfo.ModTime()) {
			newest = filepath.Join(dir, entry.Name())
			newestInfo = info
		}
	}
	if newest == "" {
		t.Fatalf("expected backup file in %s", dir)
	}
	return newest
}
