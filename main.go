package main

import (
	"fmt"
	"os"

	"github.com/whosthatknocking/kpx/cmd"
	"github.com/whosthatknocking/kpx/internal/cli"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if exitErr, ok := cli.AsExitError(err); ok {
			if exitErr.Message != "" {
				fmt.Fprintln(os.Stderr, exitErr.Message)
			}
			os.Exit(exitErr.Code)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitGeneric)
	}
}
