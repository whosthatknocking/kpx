package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	groupCmd := &cobra.Command{
		Use:   "group",
		Short: "Group commands",
	}

	groupLsCmd := &cobra.Command{
		Use:   "ls <database>",
		Short: "List groups as paths",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForRead(args[0])
			if err != nil {
				return err
			}

			for _, path := range v.ListGroups() {
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}

	groupAddCmd := &cobra.Command{
		Use:   "add <database> <group-path>",
		Short: "Create a group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForWrite(args[0])
			if err != nil {
				return err
			}

			if err := v.AddGroup(args[1]); err != nil {
				return err
			}
			if err := v.Save(); err != nil {
				return err
			}

			writeSuccess(cmd, "Created group %s\n", args[1])
			return nil
		},
	}

	groupCmd.AddCommand(groupLsCmd, groupAddCmd)
	rootCmd.AddCommand(groupCmd)
}
