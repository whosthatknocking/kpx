package cmd

type globalOptions struct {
	Quiet               bool
	JSON                bool
	NoInput             bool
	MasterPasswordStdin bool
}

type entryAddOptions struct {
	UserName      string
	URL           string
	Notes         string
	Password      string
	PasswordStdin bool
	Fields        []string
}

type entryEditOptions struct {
	Title         string
	UserName      string
	URL           string
	Notes         string
	Password      string
	PasswordStdin bool
	SetFields     []string
	ClearFields   []string
}
