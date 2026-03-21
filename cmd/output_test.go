package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/whosthatknocking/kpx/internal/vault"
)

func TestEntryJSONViewRedactsProtectedCustomFieldsUnlessReveal(t *testing.T) {
	t.Parallel()

	entry := vault.EntryRecord{
		Path:     "/Personal/GitHub",
		Title:    "GitHub",
		UserName: "alice",
		Password: "super-secret",
		CustomFields: map[string]string{
			"TOTP Seed":   "JBSWY3DPEHPK3PXP",
			"Environment": "prod",
		},
		ProtectedCustomFields: map[string]bool{
			"TOTP Seed": true,
		},
	}

	view := entryJSONView(entry, false)
	if got := view.Password; got != "[redacted]" {
		t.Fatalf("Password = %q, want %q", got, "[redacted]")
	}
	if got := view.CustomFields["TOTP Seed"]; got != "[redacted]" {
		t.Fatalf("protected custom field = %q, want %q", got, "[redacted]")
	}
	if got := view.CustomFields["Environment"]; got != "prod" {
		t.Fatalf("unprotected custom field = %q, want %q", got, "prod")
	}

	revealed := entryJSONView(entry, true)
	if got := revealed.Password; got != "super-secret" {
		t.Fatalf("revealed Password = %q, want %q", got, "super-secret")
	}
	if got := revealed.CustomFields["TOTP Seed"]; got != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("revealed protected custom field = %q, want original value", got)
	}
}

func TestPrintEntryRedactsProtectedCustomFieldsUnlessReveal(t *testing.T) {
	t.Parallel()

	entry := vault.EntryRecord{
		Path:     "/Personal/GitHub",
		Title:    "GitHub",
		UserName: "alice",
		Password: "super-secret",
		CustomFields: map[string]string{
			"TOTP Seed":   "JBSWY3DPEHPK3PXP",
			"Environment": "prod",
		},
		ProtectedCustomFields: map[string]bool{
			"TOTP Seed": true,
		},
	}

	var redacted bytes.Buffer
	printEntry(&redacted, entry, false)
	redactedText := redacted.String()
	if !strings.Contains(redactedText, "Password: [redacted]") {
		t.Fatalf("redacted output missing redacted password:\n%s", redactedText)
	}
	if !strings.Contains(redactedText, "TOTP Seed: [redacted]") {
		t.Fatalf("redacted output missing redacted protected custom field:\n%s", redactedText)
	}
	if !strings.Contains(redactedText, "Environment: prod") {
		t.Fatalf("redacted output missing unprotected custom field:\n%s", redactedText)
	}

	var revealed bytes.Buffer
	printEntry(&revealed, entry, true)
	revealedText := revealed.String()
	if !strings.Contains(revealedText, "Password: super-secret") {
		t.Fatalf("revealed output missing password:\n%s", revealedText)
	}
	if !strings.Contains(revealedText, "TOTP Seed: JBSWY3DPEHPK3PXP") {
		t.Fatalf("revealed output missing protected custom field:\n%s", revealedText)
	}
}
