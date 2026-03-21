package vault

import "testing"

func TestNormalizeGroupPath(t *testing.T) {
	tests := map[string]string{
		"":               "/",
		"/":              "/",
		"Personal":       "/Personal",
		"/Personal/Work": "/Personal/Work",
		"//Personal///A": "/Personal/A",
	}

	for input, want := range tests {
		if got := normalizeGroupPath(input); got != want {
			t.Fatalf("normalizeGroupPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSplitEntryPath(t *testing.T) {
	group, title, ok := splitEntryPath("/Personal/GitHub")
	if !ok {
		t.Fatal("expected valid entry path")
	}
	if group != "/Personal" {
		t.Fatalf("group = %q, want %q", group, "/Personal")
	}
	if title != "GitHub" {
		t.Fatalf("title = %q, want %q", title, "GitHub")
	}
}

func TestJoinGroupPath(t *testing.T) {
	if got := joinGroupPath("/", "Personal"); got != "/Personal" {
		t.Fatalf("join root = %q", got)
	}
	if got := joinGroupPath("/Personal", "Email"); got != "/Personal/Email" {
		t.Fatalf("join nested = %q", got)
	}
}
