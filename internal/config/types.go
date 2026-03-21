package config

type File struct {
	DefaultDatabase            string `yaml:"default_database,omitempty"`
	Reveal                     bool   `yaml:"reveal,omitempty"`
	MasterPasswordCacheSeconds int    `yaml:"master_password_cache_seconds,omitempty"`
}
