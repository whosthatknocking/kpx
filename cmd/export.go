package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/whosthatknocking/kpx/internal/buildinfo"
	"github.com/whosthatknocking/kpx/internal/cli"
	iexport "github.com/whosthatknocking/kpx/internal/export"
	"github.com/whosthatknocking/kpx/internal/store"
	"github.com/whosthatknocking/kpx/internal/vault"
	"golang.org/x/term"
)

func init() {
	var exportOutput string
	var exportStdout bool
	var exportForce bool
	var exportGroup string
	var exportEntry string

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export database content",
	}

	exportPaperCmd := &cobra.Command{
		Use:   "paper [database]",
		Short: "Export a printable plaintext recovery document",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _, err := resolveDatabasePath(args, 0)
			if err != nil {
				return err
			}

			if exportOutput == "" && !exportStdout {
				return cli.NewExitError(cli.ExitGeneric, "paper export requires --output or explicit --stdout")
			}
			if exportGroup != "" && exportEntry != "" {
				return cli.NewExitError(cli.ExitGeneric, "use either --group or --entry, not both")
			}
			if !exportForce && opts.NoInput {
				return cli.NewExitError(cli.ExitGeneric, "paper export requires --force when --no-input is set")
			}
			if !exportForce && !opts.NoInput {
				ok, err := cli.Confirm(cmd.ErrOrStderr(), "Export plaintext secrets for paper backup?")
				if err != nil {
					return err
				}
				if !ok {
					return cli.NewExitError(cli.ExitGeneric, "aborted")
				}
			}
			if exportStdout && term.IsTerminal(int(os.Stdout.Fd())) {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning: writing plaintext secrets to terminal stdout")
			}

			v, err := openVaultForRead(path)
			if err != nil {
				return err
			}

			entries, err := exportEntries(v, exportGroup, exportEntry)
			if err != nil {
				return err
			}

			doc := iexport.Document{
				GeneratedAt: time.Now().UTC(),
				ToolVersion: buildinfo.BaseVersion(),
				Database:    v.DatabaseName(),
				SourceFile:  v.Path(),
				Entries:     entries,
			}

			output := iexport.RenderPaper(doc)

			switch {
			case exportOutput != "":
				if err := store.WriteFileAtomic(exportOutput, []byte(output)); err != nil {
					return cli.NewExitError(cli.ExitSaveFailed, fmt.Sprintf("failed to write export %s: %v", exportOutput, err))
				}
				if opts.JSON {
					return writeStatus(cmd.OutOrStdout(), statusView{
						Status: "exported",
						Kind:   "database",
						Output: exportOutput,
						Format: "paper",
					})
				}
				writeSuccess(cmd, "Wrote paper export to %s\n", exportOutput)
				return nil
			case exportStdout:
				fmt.Fprint(cmd.OutOrStdout(), output)
				return nil
			default:
				return cli.NewExitError(cli.ExitGeneric, "paper export requires --output or explicit --stdout")
			}
		},
	}

	exportPaperCmd.Flags().StringVar(&exportOutput, "output", "", "Write the paper export to a file")
	exportPaperCmd.Flags().BoolVar(&exportStdout, "stdout", false, "Write the paper export to stdout")
	exportPaperCmd.Flags().BoolVar(&exportForce, "force", false, "Skip the plaintext export confirmation")
	exportPaperCmd.Flags().StringVar(&exportGroup, "group", "", "Export only entries under the given group path")
	exportPaperCmd.Flags().StringVar(&exportEntry, "entry", "", "Export only the given entry path")

	exportCmd.AddCommand(exportPaperCmd)
	rootCmd.AddCommand(exportCmd)
}

func exportEntries(v *vault.Vault, groupPath string, entryPath string) ([]iexport.Entry, error) {
	switch {
	case entryPath != "":
		record, err := v.GetEntry(entryPath)
		if err != nil {
			return nil, err
		}
		return []iexport.Entry{toExportEntry(record)}, nil
	case groupPath != "":
		records, err := v.ListEntriesRecursive(groupPath)
		if err != nil {
			return nil, err
		}
		return toExportEntries(records), nil
	default:
		records, err := v.ListEntriesRecursive("/")
		if err != nil {
			return nil, err
		}
		return toExportEntries(records), nil
	}
}

func toExportEntries(records []vault.EntryRecord) []iexport.Entry {
	entries := make([]iexport.Entry, 0, len(records))
	for _, record := range records {
		entries = append(entries, toExportEntry(record))
	}
	return entries
}

func toExportEntry(record vault.EntryRecord) iexport.Entry {
	return iexport.Entry{
		Path:         record.Path,
		Title:        record.Title,
		UserName:     record.UserName,
		Password:     record.Password,
		URL:          record.URL,
		Notes:        record.Notes,
		CustomFields: record.CustomFields,
	}
}
