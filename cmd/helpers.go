package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/cache"
	"github.com/whosthatknocking/kpx/internal/cli"
	"github.com/whosthatknocking/kpx/internal/config"
	"github.com/whosthatknocking/kpx/internal/vault"
)

func openVaultForRead(path string) (*vault.Vault, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if cfg.MasterPasswordCacheSeconds > 0 {
		password, ok, err := cache.Read(path, time.Now())
		if err != nil {
			return nil, err
		}
		if ok {
			v, err := vault.Open(path, password)
			if err == nil {
				return v, nil
			}
			if exitErr, ok := cli.AsExitError(err); !ok || exitErr.Code != cli.ExitAuth {
				return nil, err
			}
			if err := cache.Delete(path); err != nil {
				return nil, err
			}
		}
	}

	password, err := cli.ReadSecret(cli.SecretOptions{
		Label:     "Master password",
		NoInput:   opts.NoInput,
		FromStdin: opts.MasterPasswordStdin,
	})
	if err != nil {
		return nil, err
	}

	v, err := vault.Open(path, password)
	if err != nil {
		return nil, err
	}

	if cfg.MasterPasswordCacheSeconds > 0 {
		_ = cache.Write(path, password, time.Duration(cfg.MasterPasswordCacheSeconds)*time.Second, time.Now())
	}

	return v, nil
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
