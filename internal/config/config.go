package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	dirName  = ".kpx"
	fileName = "config.yml"
)

// DefaultPath returns the user-scoped config path used by kpx.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName, fileName), nil
}

// Load reads the optional user config file. Missing config is not an error.
func Load() (File, error) {
	path, err := DefaultPath()
	if err != nil {
		return File{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{}, nil
	}
	if err != nil {
		return File{}, err
	}

	var cfg File
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return File{}, err
	}
	return cfg, nil
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
