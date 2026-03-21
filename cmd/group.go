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
		Use:   "ls [database]",
		Short: "List groups as paths",
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

			for _, path := range v.ListGroups() {
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}

	groupAddCmd := &cobra.Command{
		Use:   "add [database] <group-path>",
		Short: "Create a group",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, remaining, err := resolveDatabasePath(args, 1)
			if err != nil {
				return err
			}

			v, err := openVaultForWrite(path)
			if err != nil {
				return err
			}
			defer v.Close()

			if err := v.AddGroup(remaining[0]); err != nil {
				return err
			}
			if err := v.Save(); err != nil {
				return err
			}

			writeSuccess(cmd, "Created group %s\n", remaining[0])
			return nil
		},
	}

	groupCmd.AddCommand(groupLsCmd, groupAddCmd)
	rootCmd.AddCommand(groupCmd)
}
