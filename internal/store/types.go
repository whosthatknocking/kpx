package store

import "time"

type BackupOptions struct {
	DestinationDir string
	FilenameFormat string
	Now            time.Time
}

type SaveOptions struct {
	Method string
}
