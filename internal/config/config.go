package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/whosthatknocking/kpx/internal/xdg"
	"gopkg.in/yaml.v3"
)

const (
	dirName       = "kpx"
	legacyDirName = ".kpx"
	fileName      = "config.yml"
)

// DefaultPath returns the user-scoped config path used by kpx.
func DefaultPath() (string, error) {
	root, err := xdg.ConfigHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, dirName, fileName), nil
}

func legacyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, legacyDirName, fileName), nil
}

// Load reads the optional user config file. Missing config is not an error.
func Load() (File, error) {
	path, err := DefaultPath()
	if err != nil {
		return File{}, err
	}

	cfg, ok, err := loadFile(path)
	if err != nil {
		return File{}, err
	}
	if ok {
		return cfg, nil
	}

	legacy, err := legacyPath()
	if err != nil {
		return File{}, err
	}

	migrated, err := migrateLegacy(path, legacy)
	if err != nil {
		return File{}, err
	}
	if !migrated {
		return File{}, nil
	}

	cfg, _, err = loadFile(path)
	return cfg, err
}

func migrateLegacy(targetPath string, sourcePath string) (bool, error) {
	if _, err := os.Stat(targetPath); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	sourceInfo, err := os.Stat(sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
		return false, err
	}

	if err := os.Rename(sourcePath, targetPath); err == nil {
		return true, nil
	} else {
		if _, statErr := os.Stat(targetPath); statErr == nil {
			return true, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if _, statErr := os.Stat(targetPath); statErr == nil {
				return true, nil
			}
			return false, nil
		}
		return false, err
	}
	if err := os.WriteFile(targetPath, data, sourceInfo.Mode().Perm()); err != nil {
		return false, err
	}
	if err := os.Remove(sourcePath); err != nil {
		return false, err
	}
	return true, nil
}

func loadFile(path string) (File, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{}, false, nil
	}
	if err != nil {
		return File{}, false, err
	}

	var cfg File
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return File{}, false, err
	}
	return cfg, true, nil
}

// Save writes the user config file with restrictive permissions.
func Save(cfg File) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
