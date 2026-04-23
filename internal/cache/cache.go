package cache

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/whosthatknocking/kpx/internal/store"
	"github.com/whosthatknocking/kpx/internal/xdg"
	"gopkg.in/yaml.v3"
)

const (
	dirName  = "kpx"
	fileName = "master-password-cache.yml"
)

func path() (string, error) {
	root, err := xdg.CacheHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, dirName, fileName), nil
}

func Read(databasePath string, now time.Time) (string, bool, error) {
	cachePath, err := path()
	if err != nil {
		return "", false, err
	}

	lock, err := store.LockExclusive(cachePath)
	if err != nil {
		return "", false, err
	}
	defer lock.Close()

	cfg, err := load(cachePath)
	if err != nil {
		return "", false, err
	}
	if pruneExpired(&cfg, now) {
		if err := saveOrDelete(cachePath, cfg); err != nil {
			return "", false, err
		}
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
		if err := saveOrDelete(cachePath, cfg); err != nil {
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

	cachePath, err := path()
	if err != nil {
		return err
	}

	lock, err := store.LockExclusive(cachePath)
	if err != nil {
		return err
	}
	defer lock.Close()

	cfg, err := load(cachePath)
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

	return save(cachePath, cfg)
}

func Delete(databasePath string) error {
	cachePath, err := path()
	if err != nil {
		return err
	}

	lock, err := store.LockExclusive(cachePath)
	if err != nil {
		return err
	}
	defer lock.Close()

	cfg, err := load(cachePath)
	if err != nil {
		return err
	}

	key, err := cacheKey(databasePath)
	if err != nil {
		return err
	}
	delete(cfg.Entries, key)
	return saveOrDelete(cachePath, cfg)
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

func load(cachePath string) (File, error) {
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

func save(cachePath string, cfg File) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return store.WriteFileAtomic(cachePath, data)
}

func saveOrDelete(cachePath string, cfg File) error {
	if len(cfg.Entries) == 0 {
		if err := os.Remove(cachePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	return save(cachePath, cfg)
}

func pruneExpired(cfg *File, now time.Time) bool {
	if len(cfg.Entries) == 0 {
		return false
	}

	changed := false
	for key, entry := range cfg.Entries {
		if entry.ExpiresAt.After(now) {
			continue
		}
		delete(cfg.Entries, key)
		changed = true
	}
	return changed
}
