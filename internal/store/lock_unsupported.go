//go:build !unix

package store

import "errors"

var errUnsupportedLockPlatform = errors.New("advisory file locking is only supported on Unix-like platforms")

// FileLock is a placeholder type for unsupported platforms so callers compile
// cleanly and receive a clear runtime error instead of undefined symbols.
type FileLock struct{}

func LockExclusive(path string) (*FileLock, error) {
	return nil, errUnsupportedLockPlatform
}

func LockShared(path string) (*FileLock, error) {
	return nil, errUnsupportedLockPlatform
}

func (l *FileLock) Close() error {
	return nil
}
