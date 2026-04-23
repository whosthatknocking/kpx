package xdg

import (
	"fmt"
	"os"
	"path/filepath"
)

func ConfigHome() (string, error) {
	return dir("XDG_CONFIG_HOME", ".config")
}

func CacheHome() (string, error) {
	return dir("XDG_CACHE_HOME", ".cache")
}

func dir(envVar string, fallback string) (string, error) {
	if value := os.Getenv(envVar); value != "" {
		if !filepath.IsAbs(value) {
			return "", fmt.Errorf("%s must be an absolute path", envVar)
		}
		return value, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, fallback), nil
}
