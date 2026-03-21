package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultBackupFilenameFormat = "{db_filename}.{timestamp}.{db_ext}"

// BackupFile copies the current database to a timestamped backup before save.
func BackupFile(path string, opts BackupOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	destinationDir := opts.DestinationDir
	if destinationDir == "" {
		destinationDir = filepath.Dir(path)
	}
	if err := os.MkdirAll(destinationDir, 0o700); err != nil {
		return err
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	filename := renderBackupFilename(path, opts.FilenameFormat, now)
	destinationPath := filepath.Join(destinationDir, filename)

	tmpFile, err := os.CreateTemp(destinationDir, ".kpx-backup-*")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()
	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}

	if err := tmpFile.Chmod(info.Mode().Perm()); err != nil {
		cleanup()
		return err
	}
	if _, err := io.Copy(tmpFile, source); err != nil {
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

	if err := os.Rename(tmpName, destinationPath); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return syncDir(destinationDir)
}

func renderBackupFilename(path string, format string, now time.Time) string {
	if format == "" {
		format = defaultBackupFilenameFormat
	}

	baseName := filepath.Base(path)
	dbExtWithDot := filepath.Ext(baseName)
	dbExt := strings.TrimPrefix(dbExtWithDot, ".")
	dbFilename := strings.TrimSuffix(baseName, dbExtWithDot)
	dbStem := dbFilename
	timestamp := now.UTC().Format("20060102T150405.000000000Z")

	replacer := strings.NewReplacer(
		"{db_filename}", dbFilename,
		"{db_stem}", dbStem,
		"{db_ext}", dbExt,
		"{timestamp}", timestamp,
	)
	rendered := replacer.Replace(format)
	rendered = filepath.Base(rendered)
	if strings.TrimSpace(rendered) == "" || rendered == "." || rendered == string(filepath.Separator) {
		return fmt.Sprintf("%s.%s.%s", dbStem, timestamp, dbExt)
	}
	return rendered
}
