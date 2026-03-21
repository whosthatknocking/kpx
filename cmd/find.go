package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	var exact bool

	findCmd := &cobra.Command{
		Use:   "find [database] <query>",
		Short: "Search entries by title or path",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, remaining, err := resolveDatabasePath(args, 1)
			if err != nil {
				return err
			}

			v, err := openVaultForRead(path)
			if err != nil {
				return err
			}

			results := v.FindEntries(remaining[0], exact)
			if opts.JSON {
				paths := make([]string, 0, len(results))
				for _, result := range results {
					paths = append(paths, result.Path)
				}
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"query":   remaining[0],
					"exact":   exact,
					"results": paths,
				})
			}

			for _, result := range results {
				fmt.Fprintln(cmd.OutOrStdout(), result.Path)
			}
			return nil
		},
	}

	findCmd.Flags().BoolVar(&exact, "exact", false, "Match the full title or path exactly instead of substring matching")
	rootCmd.AddCommand(findCmd)
}
