package config

type File struct {
	DefaultDatabase string `yaml:"default_database,omitempty"`
	Reveal          bool   `yaml:"reveal,omitempty"`
}
