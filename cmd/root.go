package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/buildinfo"
)

var opts globalOptions

var rootCmd = &cobra.Command{
	Use:           "kpx",
	Short:         "Work with KeePassXC-compatible KDBX4 databases",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       buildinfo.String(),
}

func Execute() error {
	if wantsJSONVersion(os.Args[1:]) {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(map[string]string{"version": buildinfo.String()})
	}
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress success output")
	rootCmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "Emit JSON output when supported")
	rootCmd.PersistentFlags().BoolVar(&opts.NoInput, "no-input", false, "Disable interactive prompting")
	rootCmd.PersistentFlags().BoolVar(&opts.MasterPasswordStdin, "master-password-stdin", false, "Read the database master password from stdin")
	rootCmd.SetVersionTemplate(fmt.Sprintf("kpx %s\n", rootCmd.Version))
}

func wantsJSONVersion(args []string) bool {
	var wantsJSON bool
	var wantsVersion bool

	for _, arg := range args {
		switch arg {
		case "--json":
			wantsJSON = true
		case "--version":
			wantsVersion = true
		}
	}

	return wantsJSON && wantsVersion
}
