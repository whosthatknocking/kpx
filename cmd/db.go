package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/cli"
	"github.com/whosthatknocking/kpx/internal/vault"
)

func init() {
	var createPasswordStdin bool
	var createName string

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
	}

	createCmd := &cobra.Command{
		Use:   "create <database>",
		Short: "Create a new KDBX4 database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			password, err := cli.ReadNewPassword(cli.SecretOptions{
				Label:         "Master password",
				NoInput:       opts.NoInput,
				FromStdin:     createPasswordStdin,
				ConfirmPrompt: "Confirm master password",
			})
			if err != nil {
				return err
			}

			name := createName
			if name == "" {
				name = filepath.Base(path)
			}

			if err := vault.Create(path, vault.CreateOptions{
				MasterPassword: password,
				DatabaseName:   name,
			}); err != nil {
				return err
			}

			if !opts.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
			}
			return nil
		},
	}
	createCmd.Flags().BoolVar(&createPasswordStdin, "password-stdin", false, "Read the master password from stdin")
	createCmd.Flags().StringVar(&createName, "name", "", "Database name stored in metadata")

	validateCmd := &cobra.Command{
		Use:   "validate <database>",
		Short: "Validate that a database can be opened with the supplied password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			password, err := cli.ReadSecret(cli.SecretOptions{
				Label:     "Master password",
				NoInput:   opts.NoInput,
				FromStdin: opts.MasterPasswordStdin,
			})
			if err != nil {
				return err
			}

			v, err := vault.Open(args[0], password)
			if err != nil {
				return err
			}

			if !opts.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Validated %s\n", v.Path())
			}
			return nil
		},
	}

	dbCmd.AddCommand(createCmd, validateCmd)
	rootCmd.AddCommand(dbCmd)
}
