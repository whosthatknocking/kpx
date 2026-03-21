package export

import "time"

type Document struct {
	GeneratedAt time.Time
	ToolVersion string
	Database    string
	SourceFile  string
	Entries     []Entry
}

type Entry struct {
	Path         string
	Title        string
	UserName     string
	Password     string
	URL          string
	Notes        string
	CustomFields map[string]string
}
