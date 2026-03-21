package cmd

import (
	"fmt"

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
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress success output")
	rootCmd.PersistentFlags().BoolVar(&opts.NoInput, "no-input", false, "Disable interactive prompting")
	rootCmd.PersistentFlags().BoolVar(&opts.MasterPasswordStdin, "master-password-stdin", false, "Read the database master password from stdin")
	rootCmd.SetVersionTemplate(fmt.Sprintf("kpx %s\n", rootCmd.Version))
}
