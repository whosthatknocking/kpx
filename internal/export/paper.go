package export

import (
	"fmt"
	"sort"
	"strings"
)

const separator = "========================================================================"

func RenderPaper(doc Document) string {
	var builder strings.Builder

	builder.WriteString("kpx Paper Backup\n")
	builder.WriteString(fmt.Sprintf("Generated: %s\n", doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z")))
	builder.WriteString(fmt.Sprintf("Tool Version: %s\n", doc.ToolVersion))
	if strings.TrimSpace(doc.Database) != "" {
		builder.WriteString(fmt.Sprintf("Database: %s\n", doc.Database))
	}
	builder.WriteString(fmt.Sprintf("Source File: %s\n", doc.SourceFile))

	for _, entry := range doc.Entries {
		builder.WriteString("\n")
		builder.WriteString(separator)
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("Path: %s\n", entry.Path))
		builder.WriteString(fmt.Sprintf("Title: %s\n", entry.Title))
		if strings.TrimSpace(entry.UserName) != "" {
			builder.WriteString(fmt.Sprintf("UserName: %s\n", entry.UserName))
		}
		builder.WriteString(fmt.Sprintf("Password: %s\n", entry.Password))
		if strings.TrimSpace(entry.URL) != "" {
			builder.WriteString(fmt.Sprintf("URL: %s\n", entry.URL))
		}
		if strings.TrimSpace(entry.Notes) != "" {
			builder.WriteString("Notes:\n")
			for _, line := range strings.Split(entry.Notes, "\n") {
				builder.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}

		keys := make([]string, 0, len(entry.CustomFields))
		for key := range entry.CustomFields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			builder.WriteString("\n")
			builder.WriteString("Custom Fields:\n")
			for _, key := range keys {
				value := entry.CustomFields[key]
				if strings.Contains(value, "\n") {
					builder.WriteString(fmt.Sprintf("%s:\n", key))
					for _, line := range strings.Split(value, "\n") {
						builder.WriteString(fmt.Sprintf("  %s\n", line))
					}
					continue
				}
				builder.WriteString(fmt.Sprintf("%s: %s\n", key, value))
			}
		}
	}

	return builder.String()
}
