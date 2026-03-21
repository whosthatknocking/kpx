package cmd

import (
	"fmt"
	"io"
	"sort"

	"github.com/whosthatknocking/kpx/internal/vault"
)

func printEntry(w io.Writer, entry vault.EntryRecord, reveal bool) {
	fmt.Fprintf(w, "Path: %s\n", entry.Path)
	fmt.Fprintf(w, "Title: %s\n", entry.Title)
	fmt.Fprintf(w, "UserName: %s\n", entry.UserName)
	if reveal {
		fmt.Fprintf(w, "Password: %s\n", entry.Password)
	} else {
		fmt.Fprintln(w, "Password: [redacted]")
	}
	fmt.Fprintf(w, "URL: %s\n", entry.URL)
	fmt.Fprintf(w, "Notes: %s\n", entry.Notes)

	keys := make([]string, 0, len(entry.CustomFields))
	for key := range entry.CustomFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(w, "%s: %s\n", key, entry.CustomFields[key])
	}
}
