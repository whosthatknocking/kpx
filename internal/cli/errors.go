package cli

import "errors"

const (
	ExitGeneric    = 1
	ExitAuth       = 2
	ExitFormat     = 3
	ExitNotFound   = 4
	ExitAmbiguous  = 5
	ExitSaveFailed = 6
)

func (e *ExitError) Error() string {
	return e.Message
}

// NewExitError creates an error that main translates into a non-zero exit code.
func NewExitError(code int, message string) error {
	return &ExitError{Code: code, Message: message}
}

// AsExitError reports whether err wraps an ExitError.
func AsExitError(err error) (*ExitError, bool) {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr, true
	}
	return nil, false
}
