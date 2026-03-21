package testcompat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/whosthatknocking/kpx/internal/vault"
)

type fixtureManifest struct {
	Fixtures []fixtureSpec `json:"fixtures"`
}

type fixtureSpec struct {
	Name         string          `json:"name"`
	Source       string          `json:"source"`
	Path         string          `json:"path,omitempty"`
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
	t.Parallel()

	manifest := loadManifest(t)
	if len(manifest.Fixtures) == 0 {
		t.Fatal("fixture manifest is empty")
	}

	for _, fixture := range manifest.Fixtures {
		fixture := fixture
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()

			dbPath := prepareFixture(t, fixture)

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

			if err := v.Save(); err != nil {
				t.Fatalf("Save failed: %v", err)
			}

			reopened, err := vault.Open(dbPath, fixture.Password)
			if err != nil {
				t.Fatalf("reopen after save failed: %v", err)
			}

			for _, entrySpec := range fixture.Entries {
				entry, err := reopened.GetEntry(entrySpec.Path)
				if err != nil {
					t.Fatalf("GetEntry after save for %q failed: %v", entrySpec.Path, err)
				}
				if entry.Password != entrySpec.Password {
					t.Fatalf("entry password after save for %s = %q, want %q", entrySpec.Path, entry.Password, entrySpec.Password)
				}
			}
		})
	}
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
	default:
		t.Fatalf("unsupported fixture source %q", spec.Source)
	}

	return target
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
