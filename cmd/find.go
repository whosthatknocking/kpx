package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	var exact bool

	findCmd := &cobra.Command{
		Use:   "find <database> <query>",
		Short: "Search entries by title",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultForRead(args[0])
			if err != nil {
				return err
			}

			results := v.FindEntries(args[1], exact)
			for _, result := range results {
				fmt.Fprintln(cmd.OutOrStdout(), result.Path)
			}
			return nil
		},
	}

	findCmd.Flags().BoolVar(&exact, "exact", false, "Match the title exactly instead of substring matching")
	rootCmd.AddCommand(findCmd)
}
