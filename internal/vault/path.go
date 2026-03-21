package vault

import (
	"path"
	"strings"
)

func normalizeGroupPath(input string) string {
	if input == "" {
		return "/"
	}

	cleaned := path.Clean("/" + strings.TrimSpace(input))
	if cleaned == "." {
		return "/"
	}
	return cleaned
}

func splitGroupPath(input string) []string {
	normalized := normalizeGroupPath(input)
	if normalized == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(normalized, "/"), "/")
}

func splitEntryPath(input string) (groupPath string, title string, ok bool) {
	normalized := normalizeGroupPath(input)
	if normalized == "/" {
		return "", "", false
	}

	parts := strings.Split(strings.TrimPrefix(normalized, "/"), "/")
	if len(parts) == 0 {
		return "", "", false
	}

	title = parts[len(parts)-1]
	if title == "" {
		return "", "", false
	}

	if len(parts) == 1 {
		return "/", title, true
	}
	return "/" + strings.Join(parts[:len(parts)-1], "/"), title, true
}

func joinGroupPath(parent, name string) string {
	parent = normalizeGroupPath(parent)
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}
