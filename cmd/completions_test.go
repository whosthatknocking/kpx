package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateBashCompletionIncludesCompatibilityShim(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := GenerateBashCompletion(&out); err != nil {
		t.Fatalf("GenerateBashCompletion() failed: %v", err)
	}

	text := out.String()
	for _, want := range []string{
		"_get_comp_words_by_ref()",
		"if ! declare -F _get_comp_words_by_ref",
		"__start_kpx()",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("completion output did not contain %q", want)
		}
	}
}
