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

	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion4())
	db.Credentials = gokeepasslib.NewPasswordCredentials(opts.MasterPassword)
	db.Content.Meta.DatabaseName = opts.DatabaseName

	root := gokeepasslib.NewGroup()
	root.Name = opts.DatabaseName
	db.Content.Root.Groups = []gokeepasslib.Group{root}

	v := &Vault{path: path, db: db}
	return v.Save()
}

// Open loads, decrypts, and unlocks a database for use by the CLI.
func Open(path string, password string) (*Vault, error) {
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

	return &Vault{path: path, db: db}, nil
}

// Path returns the backing file path for the opened database.
func (v *Vault) Path() string {
	return v.path
}

// Save writes the database atomically and restores unlocked protected values in memory.
func (v *Vault) Save() (err error) {
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
	if err := store.WriteFileAtomic(v.path, buf.Bytes()); err != nil {
		return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to save %s: %v", v.path, err))
	}
	return nil
}

func (v *Vault) rootGroup() *gokeepasslib.Group {
	return &v.db.Content.Root.Groups[0]
}

func isNotFound(err error) bool {
	exitErr, ok := cli.AsExitError(err)
	return ok && exitErr.Code == cli.ExitNotFound
}
