package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/whosthatknocking/kpx/internal/vault"
)

type entryView struct {
	Path         string            `json:"path"`
	Title        string            `json:"title"`
	UserName     string            `json:"username"`
	Password     string            `json:"password"`
	URL          string            `json:"url"`
	Notes        string            `json:"notes"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

type versionView struct {
	Version string `json:"version"`
}

type groupsListView struct {
	Groups []string `json:"groups"`
}

type entriesListView struct {
	Group   string   `json:"group"`
	Entries []string `json:"entries"`
}

type entryEnvelopeView struct {
	Entry entryView `json:"entry"`
}

type entryPasswordView struct {
	Path     string `json:"path"`
	Password string `json:"password"`
}

type findResultsView struct {
	Query   string   `json:"query"`
	Exact   bool     `json:"exact"`
	Results []string `json:"results"`
}

type statusView struct {
	Status string `json:"status"`
	Kind   string `json:"kind,omitempty"`
	Path   string `json:"path,omitempty"`
	Output string `json:"output,omitempty"`
	Format string `json:"format,omitempty"`
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func entryJSONView(entry vault.EntryRecord, reveal bool) entryView {
	password := "[redacted]"
	if reveal {
		password = entry.Password
	}

	customFields := make(map[string]string, len(entry.CustomFields))
	for key, value := range entry.CustomFields {
		if !reveal && entry.ProtectedCustomFields[key] {
			customFields[key] = "[redacted]"
			continue
		}
		customFields[key] = value
	}

	return entryView{
		Path:         entry.Path,
		Title:        entry.Title,
		UserName:     entry.UserName,
		Password:     password,
		URL:          entry.URL,
		Notes:        entry.Notes,
		CustomFields: customFields,
	}
}

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
		value := entry.CustomFields[key]
		if !reveal && entry.ProtectedCustomFields[key] {
			value = "[redacted]"
		}
		fmt.Fprintf(w, "%s: %s\n", key, value)
	}
}

func writeStatus(cmdOutput io.Writer, view statusView) error {
	return writeJSON(cmdOutput, view)
}
