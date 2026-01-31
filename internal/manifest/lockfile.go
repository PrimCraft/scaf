package manifest

import "time"

// Lockfile is the resolved plugin versions (plugins.lock.yaml).
type Lockfile struct {
	ResolvedAt time.Time                `yaml:"resolved_at"`
	Velocity   *ResolvedComponent       `yaml:"velocity,omitempty"`
	Paper      *ResolvedComponent       `yaml:"paper,omitempty"`
	Plugins    map[string]*ResolvedPlugin `yaml:"plugins,omitempty"`
}

// ResolvedComponent is a resolved server/proxy component.
type ResolvedComponent struct {
	Version string `yaml:"version"`
	Build   int    `yaml:"build,omitempty"`
	URL     string `yaml:"url"`
}

// ResolvedPlugin is a resolved plugin.
type ResolvedPlugin struct {
	Source     string    `yaml:"source"`
	Project    string    `yaml:"project,omitempty"`
	Version    string    `yaml:"version"`
	Platform   string    `yaml:"platform,omitempty"`
	Loader     string    `yaml:"loader,omitempty"`
	URL        string    `yaml:"url,omitempty"`
	S3URI      string    `yaml:"s3_uri,omitempty"`
	SHA256     string    `yaml:"sha256,omitempty"`
	SHA512     string    `yaml:"sha512,omitempty"`
	ResolvedAt time.Time `yaml:"resolved_at"`
}

// NewLockfile creates a new empty lockfile.
func NewLockfile() *Lockfile {
	return &Lockfile{
		ResolvedAt: time.Now().UTC(),
		Plugins:    make(map[string]*ResolvedPlugin),
	}
}
