package cli

import (
	"fmt"
	"strings"
)

func ParseFieldAssignments(values []string) (map[string]string, error) {
	fields := make(map[string]string, len(values))
	for _, item := range values {
		key, value, ok := strings.Cut(item, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, NewExitError(ExitGeneric, fmt.Sprintf("invalid field assignment %q; expected KEY=VALUE", item))
		}
		fields[key] = value
	}
	return fields, nil
}
