package cli

// ExitError carries a user-facing message and process exit code.
type ExitError struct {
	Code    int
	Message string
}

// SecretOptions controls how the CLI reads a secret value.
type SecretOptions struct {
	Label         string
	NoInput       bool
	FromStdin     bool
	ConfirmPrompt string
}
