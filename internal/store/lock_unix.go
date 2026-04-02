//go:build unix

package store

import (
	"os"
	"path/filepath"
	"syscall"
)

// FileLock coordinates cooperating kpx processes through an adjacent lock file.
type FileLock struct {
	file     *os.File
	lockPath string
}

func LockExclusive(path string) (*FileLock, error) {
	return lockPath(path, syscall.LOCK_EX)
}

func LockShared(path string) (*FileLock, error) {
	return lockPath(path, syscall.LOCK_SH)
}

func lockPath(path string, mode int) (*FileLock, error) {
	lockPath := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), mode); err != nil {
		_ = file.Close()
		return nil, err
	}

	return &FileLock{file: file, lockPath: lockPath}, nil
}

func (l *FileLock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	l.lockPath = ""

	if err != nil {
		return err
	}
	return closeErr
}
