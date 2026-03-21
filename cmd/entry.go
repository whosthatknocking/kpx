package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/cli"
	"github.com/whosthatknocking/kpx/internal/vault"
)

func init() {
	entryCmd := &cobra.Command{
		Use:   "entry",
		Short: "Entry commands",
	}

	entryLsCmd := &cobra.Command{
		Use:   "ls <database> <group-path>",
		Short: "List entries in a group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForRead(args[0])
			if err != nil {
				return err
			}

			entries, err := v.ListEntries(args[1])
			if err != nil {
				return err
			}

			for _, entry := range entries {
				fmt.Fprintln(cmd.OutOrStdout(), entry.Path)
			}
			return nil
		},
	}

	var showReveal bool
	entryShowCmd := &cobra.Command{
		Use:   "show <database> <entry-path>",
		Short: "Show entry details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForRead(args[0])
			if err != nil {
				return err
			}

			entry, err := v.GetEntry(args[1])
			if err != nil {
				return err
			}

			printEntry(cmd.OutOrStdout(), entry, showReveal)
			return nil
		},
	}
	entryShowCmd.Flags().BoolVar(&showReveal, "reveal", false, "Show the stored password")

	var addOpts entryAddOptions
	entryAddCmd := &cobra.Command{
		Use:   "add <database> <entry-path>",
		Short: "Add an entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForWrite(args[0])
			if err != nil {
				return err
			}

			password, err := entryPassword("Entry password", addOpts.Password, addOpts.PasswordStdin)
			if err != nil {
				return err
			}

			customFields, err := cli.ParseFieldAssignments(addOpts.Fields)
			if err != nil {
				return err
			}

			if err := v.AddEntry(args[1], vault.EntryInput{
				UserName:     addOpts.UserName,
				Password:     password,
				URL:          addOpts.URL,
				Notes:        addOpts.Notes,
				CustomFields: customFields,
			}); err != nil {
				return err
			}

			if err := v.Save(); err != nil {
				return err
			}

			writeSuccess(cmd, "Created entry %s\n", args[1])
			return nil
		},
	}
	entryAddCmd.Flags().StringVar(&addOpts.UserName, "username", "", "Username")
	entryAddCmd.Flags().StringVar(&addOpts.URL, "url", "", "URL")
	entryAddCmd.Flags().StringVar(&addOpts.Notes, "notes", "", "Notes")
	entryAddCmd.Flags().StringVar(&addOpts.Password, "password", "", "Entry password")
	entryAddCmd.Flags().BoolVar(&addOpts.PasswordStdin, "password-stdin", false, "Read the entry password from stdin")
	entryAddCmd.Flags().StringArrayVar(&addOpts.Fields, "field", nil, "Custom field assignment in KEY=VALUE form")

	var editOpts entryEditOptions
	entryEditCmd := &cobra.Command{
		Use:   "edit <database> <entry-path>",
		Short: "Edit entry fields",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForWrite(args[0])
			if err != nil {
				return err
			}

			setCustom, err := cli.ParseFieldAssignments(editOpts.SetFields)
			if err != nil {
				return err
			}

			patch := vault.EntryPatch{
				SetCustomFields:   setCustom,
				ClearCustomFields: editOpts.ClearFields,
			}

			if cmd.Flags().Changed("title") {
				patch.Title = cli.StringPtr(editOpts.Title)
			}
			if cmd.Flags().Changed("username") {
				patch.UserName = cli.StringPtr(editOpts.UserName)
			}
			if cmd.Flags().Changed("url") {
				patch.URL = cli.StringPtr(editOpts.URL)
			}
			if cmd.Flags().Changed("notes") {
				patch.Notes = cli.StringPtr(editOpts.Notes)
			}
			if cmd.Flags().Changed("password") || editOpts.PasswordStdin {
				password, err := entryPassword("Entry password", editOpts.Password, editOpts.PasswordStdin)
				if err != nil {
					return err
				}
				patch.Password = cli.StringPtr(password)
			}

			if err := v.EditEntry(args[1], patch); err != nil {
				return err
			}
			if err := v.Save(); err != nil {
				return err
			}

			writeSuccess(cmd, "Updated entry %s\n", strings.TrimSpace(args[1]))
			return nil
		},
	}
	entryEditCmd.Flags().StringVar(&editOpts.Title, "title", "", "Updated title")
	entryEditCmd.Flags().StringVar(&editOpts.UserName, "username", "", "Updated username")
	entryEditCmd.Flags().StringVar(&editOpts.URL, "url", "", "Updated URL")
	entryEditCmd.Flags().StringVar(&editOpts.Notes, "notes", "", "Updated notes")
	entryEditCmd.Flags().StringVar(&editOpts.Password, "password", "", "Updated password")
	entryEditCmd.Flags().BoolVar(&editOpts.PasswordStdin, "password-stdin", false, "Read the updated password from stdin")
	entryEditCmd.Flags().StringArrayVar(&editOpts.SetFields, "set-field", nil, "Set a custom field using KEY=VALUE")
	entryEditCmd.Flags().StringArrayVar(&editOpts.ClearFields, "clear-field", nil, "Remove a custom field by key")

	var rmForce bool
	entryRmCmd := &cobra.Command{
		Use:   "rm <database> <entry-path>",
		Short: "Delete an entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !rmForce && opts.NoInput {
				return cli.NewExitError(cli.ExitGeneric, "delete requires --force when --no-input is set")
			}
			if !rmForce && !opts.NoInput {
				ok, err := cli.Confirm(cmd.ErrOrStderr(), fmt.Sprintf("Delete %s?", args[1]))
				if err != nil {
					return err
				}
				if !ok {
					return cli.NewExitError(cli.ExitGeneric, "aborted")
				}
			}

			v, err := openVaultForWrite(args[0])
			if err != nil {
				return err
			}

			if err := v.DeleteEntry(args[1]); err != nil {
				return err
			}
			if err := v.Save(); err != nil {
				return err
			}

			writeSuccess(cmd, "Deleted entry %s\n", args[1])
			return nil
		},
	}
	entryRmCmd.Flags().BoolVar(&rmForce, "force", false, "Skip delete confirmation")

	entryCmd.AddCommand(entryLsCmd, entryShowCmd, entryAddCmd, entryEditCmd, entryRmCmd)
	rootCmd.AddCommand(entryCmd)
}
