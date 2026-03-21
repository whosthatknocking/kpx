package buildinfo

import "fmt"

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

func String() string {
	base := Version
	if base == "" {
		base = "dev"
	}

	if Commit == "" && Date == "" {
		return base
	}
	if Commit != "" && Date != "" {
		return fmt.Sprintf("%s (commit %s, built %s)", base, Commit, Date)
	}
	if Commit != "" {
		return fmt.Sprintf("%s (commit %s)", base, Commit)
	}
	return fmt.Sprintf("%s (built %s)", base, Date)
}
