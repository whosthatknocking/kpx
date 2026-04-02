package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/buildinfo"
)

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the kpx version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if opts.JSON {
				_ = writeJSON(cmd.OutOrStdout(), versionView{Version: buildinfo.String()})
				return
			}
			fmt.Fprintf(cmd.OutOrStdout(), "kpx %s\n", buildinfo.String())
		},
	}

	rootCmd.AddCommand(versionCmd)
}
