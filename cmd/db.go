package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/cli"
	"github.com/whosthatknocking/kpx/internal/vault"
)

func init() {
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
				FromStdin:     opts.MasterPasswordStdin,
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

			if opts.JSON {
				return writeStatus(cmd.OutOrStdout(), statusView{
					Status: "created",
					Kind:   "database",
					Path:   path,
				})
			}
			if !opts.Quiet {
				fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
			}
			return nil
		},
	}
	createCmd.Flags().StringVar(&createName, "name", "", "Database name stored in metadata")

	validateCmd := &cobra.Command{
		Use:   "validate [database]",
		Short: "Validate that a database can be opened with the supplied password",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _, err := resolveDatabasePath(args, 0)
			if err != nil {
				return err
			}

			v, err := openVaultForRead(path)
			if err != nil {
				return err
			}

			if opts.JSON {
				return writeStatus(cmd.OutOrStdout(), statusView{
					Status: "validated",
					Kind:   "database",
					Path:   v.Path(),
				})
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
