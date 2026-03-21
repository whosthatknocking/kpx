package cmd

type globalOptions struct {
	Quiet               bool
	JSON                bool
	NoInput             bool
	MasterPasswordStdin bool
}

type entryAddOptions struct {
	UserName           string
	URL                string
	Notes              string
	Password           string
	EntryPasswordStdin bool
	Fields             []string
}

type entryEditOptions struct {
	Title              string
	UserName           string
	URL                string
	Notes              string
	Password           string
	EntryPasswordStdin bool
	SetFields          []string
	ClearFields        []string
}
