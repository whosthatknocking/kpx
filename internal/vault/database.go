package vault

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
	"github.com/whosthatknocking/kpx/internal/cli"
	"github.com/whosthatknocking/kpx/internal/config"
	"github.com/whosthatknocking/kpx/internal/store"
)

// Create initializes a new KDBX4 database at path.
func Create(path string, opts CreateOptions) error {
	if opts.MasterPassword == "" {
		return cli.NewExitError(cli.ExitGeneric, "master password cannot be empty")
	}

	writeLock, err := store.LockExclusive(path)
	if err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to lock %s: %v", path, err))
	}
	defer writeLock.Close()

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion4())
	db.Credentials = gokeepasslib.NewPasswordCredentials(opts.MasterPassword)
	db.Content.Meta.DatabaseName = opts.DatabaseName

	root := gokeepasslib.NewGroup()
	root.Name = opts.DatabaseName
	db.Content.Root.Groups = []gokeepasslib.Group{root}

	v := &Vault{path: path, db: db, writeLock: writeLock}
	return v.Save()
}

// Open loads, decrypts, and unlocks a database for use by the CLI.
func Open(path string, password string) (*Vault, error) {
	readLock, err := store.LockShared(path)
	if err != nil {
		return nil, err
	}
	defer readLock.Close()

	return openUnlocked(path, password, nil)
}

// OpenForWrite acquires an exclusive lock before opening a database for mutation.
func OpenForWrite(path string, password string) (*Vault, error) {
	writeLock, err := store.LockExclusive(path)
	if err != nil {
		return nil, cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to lock %s: %v", path, err))
	}

	v, err := openUnlocked(path, password, writeLock)
	if err != nil {
		_ = writeLock.Close()
		return nil, err
	}
	return v, nil
}

func openUnlocked(path string, password string, writeLock *store.FileLock) (*Vault, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cli.NewExitError(cli.ExitNotFound, fmt.Sprintf("database not found: %s", path))
		}
		return nil, err
	}
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	if err := gokeepasslib.NewDecoder(file).Decode(db); err != nil {
		message := fmt.Sprintf("failed to open %s", path)
		if strings.Contains(strings.ToLower(err.Error()), "credentials") {
			return nil, cli.NewExitError(cli.ExitAuth, message)
		}
		return nil, cli.NewExitError(cli.ExitFormat, fmt.Sprintf("%s: %v", message, err))
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, err
	}
	if db.Content == nil || db.Content.Root == nil {
		return nil, cli.NewExitError(cli.ExitFormat, "database content is missing a root node")
	}
	if len(db.Content.Root.Groups) == 0 {
		root := gokeepasslib.NewGroup()
		root.Name = filepath.Base(path)
		db.Content.Root.Groups = []gokeepasslib.Group{root}
	}

	return &Vault{path: path, db: db, writeLock: writeLock}, nil
}

// Path returns the backing file path for the opened database.
func (v *Vault) Path() string {
	return v.path
}

// Save backs up the current file, writes the database using the configured
// save method, and restores unlocked protected values in memory.
func (v *Vault) Save() (err error) {
	if v.writeLock == nil {
		v.writeLock, err = store.LockExclusive(v.path)
		if err != nil {
			return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to lock %s: %v", v.path, err))
		}
		defer func() {
			closeErr := v.Close()
			if closeErr != nil && err == nil {
				err = cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to release lock for %s: %v", v.path, closeErr))
			}
		}()
	}

	if err := v.db.LockProtectedEntries(); err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to lock protected entries: %v", err))
	}
	defer func() {
		unlockErr := v.db.UnlockProtectedEntries()
		if unlockErr != nil && err == nil {
			err = cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to restore unlocked entry state: %v", unlockErr))
		}
	}()

	var buf bytes.Buffer
	if err := gokeepasslib.NewEncoder(&buf).Encode(v.db); err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to encode database: %v", err))
	}
	cfg, err := config.Load()
	if err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to load config: %v", err))
	}
	if err := store.BackupFile(v.path, store.BackupOptions{
		DestinationDir: cfg.BackupDirectory,
		FilenameFormat: cfg.BackupFilenameFormat,
	}); err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to back up %s: %v", v.path, err))
	}
	if err := store.WriteFile(v.path, buf.Bytes(), store.SaveOptions{
		Method: cfg.SaveMethod,
	}); err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to save %s: %v", v.path, err))
	}
	return nil
}

// Close releases any retained write lock held for a mutating command.
func (v *Vault) Close() error {
	if v == nil || v.writeLock == nil {
		return nil
	}
	lock := v.writeLock
	v.writeLock = nil
	return lock.Close()
}

func (v *Vault) rootGroup() *gokeepasslib.Group {
	return &v.db.Content.Root.Groups[0]
}

func isNotFound(err error) bool {
	exitErr, ok := cli.AsExitError(err)
	return ok && exitErr.Code == cli.ExitNotFound
}
