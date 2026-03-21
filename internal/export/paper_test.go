package export

import (
	"strings"
	"testing"
	"time"
)

func TestRenderPaper(t *testing.T) {
	t.Parallel()

	doc := Document{
		GeneratedAt: time.Date(2026, 3, 21, 18, 42, 0, 0, time.UTC),
		ToolVersion: "0.1.6",
		Database:    "Personal Vault",
		SourceFile:  "/tmp/vault.kdbx",
		Entries: []Entry{
			{
				Path:     "/Personal/GitHub",
				Title:    "GitHub",
				UserName: "alice",
				Password: "super-secret",
				URL:      "https://github.com",
				Notes:    "Personal account\nKeep safe",
				CustomFields: map[string]string{
					"Recovery Code": "ABCD-EFGH",
				},
			},
		},
	}

	got := RenderPaper(doc)

	for _, want := range []string{
		"kpx Paper Backup",
		"Generated: 2026-03-21T18:42:00Z",
		"Tool Version: 0.1.6",
		"Database: Personal Vault",
		"Source File: /tmp/vault.kdbx",
		"Path: /Personal/GitHub",
		"Password: super-secret",
		"Notes:\n  Personal account\n  Keep safe",
		"Custom Fields:\nRecovery Code: ABCD-EFGH",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderPaper() did not contain %q\n%s", want, got)
		}
	}
}
