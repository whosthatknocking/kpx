package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/cli"
	"github.com/whosthatknocking/kpx/internal/config"
	"github.com/whosthatknocking/kpx/internal/vault"
)

func openVaultForRead(path string) (*vault.Vault, error) {
	password, err := cli.ReadSecret(cli.SecretOptions{
		Label:     "Master password",
		NoInput:   opts.NoInput,
		FromStdin: opts.MasterPasswordStdin,
	})
	if err != nil {
		return nil, err
	}

	return vault.Open(path, password)
}

func openVaultForWrite(path string) (*vault.Vault, error) {
	return openVaultForRead(path)
}

func resolveDatabasePath(args []string, trailingRequired int) (string, []string, error) {
	switch len(args) {
	case trailingRequired + 1:
		return args[0], args[1:], nil
	case trailingRequired:
		cfg, err := config.Load()
		if err != nil {
			return "", nil, err
		}
		if cfg.DefaultDatabase == "" {
			return "", nil, cli.NewExitError(cli.ExitGeneric, "database path not provided and no default database configured; set default_database in ~/.kpx/config.yml")
		}
		return cfg.DefaultDatabase, args, nil
	default:
		return "", nil, cli.NewExitError(cli.ExitGeneric, "invalid arguments")
	}
}

func entryPassword(label string, value string, fromStdin bool) (string, error) {
	if !fromStdin {
		return value, nil
	}

	return cli.ReadSecret(cli.SecretOptions{
		Label:     label,
		NoInput:   opts.NoInput,
		FromStdin: true,
	})
}

func writeSuccess(cmd *cobra.Command, format string, args ...any) {
	if opts.Quiet {
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), format, args...)
}
