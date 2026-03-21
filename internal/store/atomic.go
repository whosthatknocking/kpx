package store

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	SaveMethodAtomic = "temporary_file"
	SaveMethodDirect = "direct_write"
)

// WriteFile saves data using the configured persistence strategy.
func WriteFile(path string, data []byte, opts SaveOptions) error {
	switch opts.Method {
	case "", SaveMethodAtomic:
		return WriteFileAtomic(path, data)
	case SaveMethodDirect:
		return WriteFileDirect(path, data)
	default:
		return errors.New("unsupported save method: " + opts.Method)
	}
}

// WriteFileAtomic replaces path with data via a temp file in the same directory.
// It preserves existing file permissions when the target already exists.
func WriteFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	mode, err := fileModeForPath(path)
	if err != nil {
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

	return syncDir(dir)
}

// WriteFileDirect writes directly to the target path and fsyncs the file.
// It preserves existing file permissions when the target already exists.
func WriteFileDirect(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	mode, err := fileModeForPath(path)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	return syncDir(dir)
}

func fileModeForPath(path string) (os.FileMode, error) {
	mode := os.FileMode(0o600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	return mode, nil
}

func syncDir(path string) error {
	dirHandle, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer dirHandle.Close()

	return dirHandle.Sync()
}
