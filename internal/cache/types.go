package cache

import "time"

type File struct {
	Entries map[string]Entry `yaml:"entries,omitempty"`
}

type Entry struct {
	Password  string    `yaml:"password,omitempty"`
	ExpiresAt time.Time `yaml:"expires_at,omitempty"`
}
