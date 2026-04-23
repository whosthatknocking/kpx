package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPathUsesXDGConfigHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	xdgConfigHome := filepath.Join(t.TempDir(), "xdg-config")
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() failed: %v", err)
	}

	want := filepath.Join(xdgConfigHome, "kpx", "config.yml")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestDefaultPathFallsBackToHomeConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() failed: %v", err)
	}

	want := filepath.Join(homeDir, ".config", "kpx", "config.yml")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestLoadMigratesLegacyPathToXDGPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))

	legacyConfigPath := filepath.Join(homeDir, ".kpx", "config.yml")
	if err := os.MkdirAll(filepath.Dir(legacyConfigPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(legacyConfigPath, []byte("reveal: true\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if !cfg.Reveal {
		t.Fatal("Load() did not read reveal setting from legacy config")
	}

	newConfigPath := filepath.Join(homeDir, ".config", "kpx", "config.yml")
	data, err := os.ReadFile(newConfigPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s) failed: %v", newConfigPath, err)
	}
	if string(data) != "reveal: true\n" {
		t.Fatalf("migrated config contents = %q, want %q", string(data), "reveal: true\n")
	}
	if _, err := os.Stat(legacyConfigPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy config to be removed after migration, stat err = %v", err)
	}
}

func TestLoadPrefersXDGPathOverLegacyPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))

	newConfigPath := filepath.Join(homeDir, ".config", "kpx", "config.yml")
	if err := os.MkdirAll(filepath.Dir(newConfigPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(newConfigPath, []byte("reveal: false\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	legacyConfigPath := filepath.Join(homeDir, ".kpx", "config.yml")
	if err := os.MkdirAll(filepath.Dir(legacyConfigPath), 0o700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(legacyConfigPath, []byte("reveal: true\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Reveal {
		t.Fatal("Load() unexpectedly read legacy config when XDG config exists")
	}

	if _, err := os.Stat(legacyConfigPath); err != nil {
		t.Fatalf("expected legacy config to remain untouched when XDG config exists, stat err = %v", err)
	}
}
