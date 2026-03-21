package store

import (
	"errors"
	"os"
	"path/filepath"
)

// WriteFileAtomic replaces path with data via a temp file in the same directory.
// It preserves existing file permissions when the target already exists.
func WriteFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	mode := os.FileMode(0o600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".kpx-*")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()
	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}

	if err := tmpFile.Chmod(mode); err != nil {
		cleanup()
		return err
	}
	if _, err := tmpFile.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	dirHandle, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer dirHandle.Close()

	return dirHandle.Sync()
}
