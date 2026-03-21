package vault

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
	"github.com/whosthatknocking/kpx/internal/cli"
)

// ListEntries returns the entries directly under the requested group path.
func (v *Vault) ListEntries(groupPath string) ([]EntryRecord, error) {
	group, err := v.groupByPath(groupPath)
	if err != nil {
		return nil, err
	}

	entries := make([]EntryRecord, 0, len(group.Entries))
	for i := range group.Entries {
		record := entryRecord(groupPath, &group.Entries[i])
		entries = append(entries, record)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

// ListEntriesRecursive returns all entries under the requested group path.
func (v *Vault) ListEntriesRecursive(groupPath string) ([]EntryRecord, error) {
	group, err := v.groupByPath(groupPath)
	if err != nil {
		return nil, err
	}

	normalized := normalizeGroupPath(groupPath)
	entries := make([]EntryRecord, 0)

	var walk func(parentPath string, group *gokeepasslib.Group)
	walk = func(parentPath string, group *gokeepasslib.Group) {
		for i := range group.Entries {
			entries = append(entries, entryRecord(parentPath, &group.Entries[i]))
		}
		for i := range group.Groups {
			child := &group.Groups[i]
			childPath := joinGroupPath(parentPath, child.Name)
			walk(childPath, child)
		}
	}

	walk(normalized, group)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

// GetEntry resolves a single entry by full path.
func (v *Vault) GetEntry(entryPath string) (EntryRecord, error) {
	groupPath, title, ok := splitEntryPath(entryPath)
	if !ok {
		return EntryRecord{}, cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("invalid entry path: %s", entryPath))
	}

	group, err := v.groupByPath(groupPath)
	if err != nil {
		return EntryRecord{}, err
	}

	entry, err := findUniqueEntry(group, title)
	if err != nil {
		return EntryRecord{}, err
	}

	return entryRecord(groupPath, entry), nil
}

// AddEntry creates an entry under an existing group path.
func (v *Vault) AddEntry(entryPath string, input EntryInput) error {
	groupPath, title, ok := splitEntryPath(entryPath)
	if !ok {
		return cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("invalid entry path: %s", entryPath))
	}

	group, err := v.groupByPath(groupPath)
	if err != nil {
		return err
	}
	if existing, err := findUniqueEntry(group, title); err == nil && existing != nil {
		return cli.NewExitError(cli.ExitAmbiguous, fmt.Sprintf("entry already exists: %s", entryPath))
	} else if err != nil && !isNotFound(err) {
		return err
	}

	entry := gokeepasslib.NewEntry()
	entry.Values = []gokeepasslib.ValueData{
		value(fieldTitle, title, false),
		value(fieldUserName, input.UserName, false),
		value(fieldPassword, input.Password, true),
		value(fieldURL, input.URL, false),
		value(fieldNotes, input.Notes, false),
	}

	keys := make([]string, 0, len(input.CustomFields))
	for key := range input.CustomFields {
		if isBuiltinField(key) {
			return cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("custom field %q conflicts with a built-in field", key))
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry.Values = append(entry.Values, value(key, input.CustomFields[key], false))
	}

	group.Entries = append(group.Entries, entry)
	return nil
}

// EditEntry updates built-in and custom fields for an existing entry.
func (v *Vault) EditEntry(entryPath string, patch EntryPatch) error {
	groupPath, title, ok := splitEntryPath(entryPath)
	if !ok {
		return cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("invalid entry path: %s", entryPath))
	}

	group, err := v.groupByPath(groupPath)
	if err != nil {
		return err
	}

	entry, err := findUniqueEntry(group, title)
	if err != nil {
		return err
	}

	if patch.Title != nil && *patch.Title != title {
		if existing, err := findUniqueEntry(group, *patch.Title); err == nil && existing != nil {
			return cli.NewExitError(cli.ExitAmbiguous, fmt.Sprintf("entry already exists with title %q", *patch.Title))
		} else if err != nil && !isNotFound(err) {
			return err
		}
	}

	if patch.Title != nil {
		setValue(entry, fieldTitle, *patch.Title, false)
	}
	if patch.UserName != nil {
		setValue(entry, fieldUserName, *patch.UserName, false)
	}
	if patch.Password != nil {
		setValue(entry, fieldPassword, *patch.Password, true)
	}
	if patch.URL != nil {
		setValue(entry, fieldURL, *patch.URL, false)
	}
	if patch.Notes != nil {
		setValue(entry, fieldNotes, *patch.Notes, false)
	}

	for key, value := range patch.SetCustomFields {
		if isBuiltinField(key) {
			return cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("custom field %q conflicts with a built-in field", key))
		}
		setValue(entry, key, value, false)
	}
	for _, key := range patch.ClearCustomFields {
		if isBuiltinField(key) {
			return cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("cannot clear built-in field %q with --clear-field", key))
		}
		removeValue(entry, key)
	}

	return nil
}

// DeleteEntry removes a single entry resolved by full path.
func (v *Vault) DeleteEntry(entryPath string) error {
	groupPath, title, ok := splitEntryPath(entryPath)
	if !ok {
		return cli.NewExitError(cli.ExitGeneric, fmt.Sprintf("invalid entry path: %s", entryPath))
	}

	group, err := v.groupByPath(groupPath)
	if err != nil {
		return err
	}

	index := -1
	for i := range group.Entries {
		if group.Entries[i].GetTitle() == title {
			if index >= 0 {
				return cli.NewExitError(cli.ExitAmbiguous, fmt.Sprintf("multiple entries matched %s", entryPath))
			}
			index = i
		}
	}
	if index < 0 {
		return cli.NewExitError(cli.ExitNotFound, fmt.Sprintf("entry not found: %s", entryPath))
	}

	group.Entries = append(group.Entries[:index], group.Entries[index+1:]...)
	return nil
}

// FindEntries searches entry titles with exact or case-insensitive substring matching.
func (v *Vault) FindEntries(query string, exact bool) []EntryRecord {
	lowered := strings.ToLower(query)
	results := make([]EntryRecord, 0)

	var walk func(groupPath string, group *gokeepasslib.Group)
	walk = func(groupPath string, group *gokeepasslib.Group) {
		for i := range group.Entries {
			entry := &group.Entries[i]
			title := entry.GetTitle()
			match := false
			if exact {
				match = strings.EqualFold(title, query)
			} else {
				match = strings.Contains(strings.ToLower(title), lowered)
			}
			if match {
				results = append(results, entryRecord(groupPath, entry))
			}
		}
		for i := range group.Groups {
			child := &group.Groups[i]
			walk(joinGroupPath(groupPath, child.Name), child)
		}
	}

	walk("/", v.rootGroup())
	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results
}

func findUniqueEntry(group *gokeepasslib.Group, title string) (*gokeepasslib.Entry, error) {
	var match *gokeepasslib.Entry
	for i := range group.Entries {
		entry := &group.Entries[i]
		if entry.GetTitle() != title {
			continue
		}
		if match != nil {
			return nil, cli.NewExitError(cli.ExitAmbiguous, fmt.Sprintf("multiple entries matched %q", title))
		}
		match = entry
	}
	if match == nil {
		return nil, cli.NewExitError(cli.ExitNotFound, fmt.Sprintf("entry not found: %s", title))
	}
	return match, nil
}

func entryRecord(groupPath string, entry *gokeepasslib.Entry) EntryRecord {
	record := EntryRecord{
		Path:                  joinGroupPath(groupPath, entry.GetTitle()),
		Title:                 entry.GetContent(fieldTitle),
		UserName:              entry.GetContent(fieldUserName),
		Password:              entry.GetContent(fieldPassword),
		URL:                   entry.GetContent(fieldURL),
		Notes:                 entry.GetContent(fieldNotes),
		CustomFields:          map[string]string{},
		ProtectedCustomFields: map[string]bool{},
	}

	for _, item := range entry.Values {
		if isBuiltinField(item.Key) {
			continue
		}
		record.CustomFields[item.Key] = item.Value.Content
		record.ProtectedCustomFields[item.Key] = item.Value.Protected.Bool
	}

	return record
}
