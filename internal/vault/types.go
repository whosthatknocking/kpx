package vault

import (
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/whosthatknocking/kpx/internal/store"
)

type CreateOptions struct {
	MasterPassword string
	DatabaseName   string
}

// EntryInput contains the supported fields for a new entry.
type EntryInput struct {
	UserName     string
	Password     string
	URL          string
	Notes        string
	CustomFields map[string]string
}

// EntryPatch describes partial updates for an existing entry.
type EntryPatch struct {
	Title             *string
	UserName          *string
	Password          *string
	URL               *string
	Notes             *string
	SetCustomFields   map[string]string
	ClearCustomFields []string
}

// EntryRecord is the app-facing representation used by CLI output and tests.
type EntryRecord struct {
	Path                  string
	Title                 string
	UserName              string
	Password              string
	URL                   string
	Notes                 string
	CustomFields          map[string]string
	ProtectedCustomFields map[string]bool
}

// Vault wraps the KDBX library behind path-based operations used by the CLI.
type Vault struct {
	path      string
	db        *gokeepasslib.Database
	writeLock *store.FileLock
}
