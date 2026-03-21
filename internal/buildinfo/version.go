package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

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
	base = Version
	if base == "" {
		base = "dev"
	}

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
	if base == "dev" && commit != "" {
		base = "dev+" + commit
	}

	return base, commit, date, modified
}

func shortRevision(value string) string {
	if len(value) > 7 {
		return value[:7]
	}
	return strings.TrimSpace(value)
}
