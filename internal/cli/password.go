package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadSecret reads a secret from stdin or an interactive tty prompt.
func ReadSecret(opts SecretOptions) (string, error) {
	if opts.FromStdin {
		return readLine(os.Stdin)
	}
	if opts.NoInput {
		return "", NewExitError(ExitGeneric, "interactive input disabled; use the appropriate stdin flag")
	}

	return readPasswordPrompt(opts.Label)
}

// ReadNewPassword reads and optionally confirms a newly chosen secret.
func ReadNewPassword(opts SecretOptions) (string, error) {
	password, err := ReadSecret(opts)
	if err != nil {
		return "", err
	}

	if opts.FromStdin || opts.NoInput || opts.ConfirmPrompt == "" {
		return password, nil
	}

	confirm, err := readPasswordPrompt(opts.ConfirmPrompt)
	if err != nil {
		return "", err
	}
	if password != confirm {
		return "", NewExitError(ExitGeneric, "passwords did not match")
	}
	return password, nil
}

func readPasswordPrompt(label string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", NewExitError(ExitGeneric, "no interactive tty available; use the appropriate stdin flag")
	}
	defer tty.Close()

	if label != "" {
		fmt.Fprintf(tty, "%s: ", label)
	}
	bytes, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func readLine(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// Confirm prompts for a yes/no answer on the controlling tty.
func Confirm(w io.Writer, prompt string) (bool, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false, NewExitError(ExitGeneric, "no interactive tty available; rerun with --force or --no-input")
	}
	defer tty.Close()

	fmt.Fprintf(w, "%s [y/N]: ", prompt)
	answer, err := readLine(tty)
	if err != nil {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

// StringPtr is a small helper for optional string fields in command patches.
func StringPtr(value string) *string {
	return &value
}
