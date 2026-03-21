package buildinfo

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"
)

//go:embed VERSION.txt
var embeddedVersion string

var (
	Commit = ""
	Date   = ""
)

// String returns the base release version plus any available build metadata.
func String() string {
	base, commit, date, modified := current()

	if commit == "" && date == "" && !modified {
		return base
	}
	if commit != "" && date != "" {
		suffix := fmt.Sprintf("commit %s, built %s", commit, date)
		if modified {
			suffix += ", modified"
		}
		return fmt.Sprintf("%s (%s)", base, suffix)
	}
	if commit != "" {
		if modified {
			return fmt.Sprintf("%s (commit %s, modified)", base, commit)
		}
		return fmt.Sprintf("%s (commit %s)", base, commit)
	}
	if modified {
		return fmt.Sprintf("%s (modified)", base)
	}
	return fmt.Sprintf("%s (built %s)", base, date)
}

func current() (base string, commit string, date string, modified bool) {
	base = BaseVersion()

	commit = Commit
	date = Date

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if commit == "" {
					commit = shortRevision(setting.Value)
				}
			case "vcs.time":
				if date == "" {
					date = setting.Value
				}
			case "vcs.modified":
				if setting.Value == "true" {
					modified = true
				}
			}
		}
	}

	if Commit != "" {
		commit = shortRevision(Commit)
	}
	if base == "" {
		base = "dev"
	}
	if base == "dev" && commit != "" {
		base = "dev+" + commit
	}

	return base, commit, date, modified
}

// BaseVersion returns the embedded release version from VERSION.txt.
func BaseVersion() string {
	return strings.TrimSpace(embeddedVersion)
}

func shortRevision(value string) string {
	if len(value) > 7 {
		return value[:7]
	}
	return strings.TrimSpace(value)
}
