package cache

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	dirName  = ".kpx"
	fileName = "master-password-cache.yml"
)

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName, fileName), nil
}

func Read(databasePath string, now time.Time) (string, bool, error) {
	cfg, err := load()
	if err != nil {
		return "", false, err
	}

	key, err := cacheKey(databasePath)
	if err != nil {
		return "", false, err
	}

	entry, ok := cfg.Entries[key]
	if !ok {
		return "", false, nil
	}
	if !entry.ExpiresAt.After(now) {
		delete(cfg.Entries, key)
		if err := save(cfg); err != nil {
			return "", false, err
		}
		return "", false, nil
	}
	return entry.Password, true, nil
}

func Write(databasePath string, password string, ttl time.Duration, now time.Time) error {
	if ttl <= 0 {
		return Delete(databasePath)
	}

	cfg, err := load()
	if err != nil {
		return err
	}

	key, err := cacheKey(databasePath)
	if err != nil {
		return err
	}
	if cfg.Entries == nil {
		cfg.Entries = map[string]Entry{}
	}
	cfg.Entries[key] = Entry{
		Password:  password,
		ExpiresAt: now.Add(ttl),
	}

	return save(cfg)
}

func Delete(databasePath string) error {
	cfg, err := load()
	if err != nil {
		return err
	}

	key, err := cacheKey(databasePath)
	if err != nil {
		return err
	}

	if len(cfg.Entries) == 0 {
		return nil
	}
	delete(cfg.Entries, key)
	return save(cfg)
}

func cacheKey(databasePath string) (string, error) {
	if databasePath == "" {
		return "", nil
	}
	absolute, err := filepath.Abs(databasePath)
	if err != nil {
		return databasePath, nil
	}
	return absolute, nil
}

func load() (File, error) {
	cachePath, err := path()
	if err != nil {
		return File{}, err
	}

	data, err := os.ReadFile(cachePath)
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

func save(cfg File) error {
	cachePath, err := path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0o600)
}
