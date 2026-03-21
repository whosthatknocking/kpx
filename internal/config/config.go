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

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName, fileName), nil
}

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
