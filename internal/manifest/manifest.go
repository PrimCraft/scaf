// Package manifest defines types for plugin manifests and lockfiles.
package manifest

// Manifest is the input configuration file (plugins.yaml).
type Manifest struct {
	Velocity VelocityConfig           `yaml:"velocity,omitempty"`
	Paper    PaperConfig              `yaml:"paper,omitempty"`
	Plugins  map[string]*PluginConfig `yaml:"plugins,omitempty"`
}

// VelocityConfig configures the Velocity proxy.
type VelocityConfig struct {
	Version string `yaml:"version,omitempty"`
}

// PaperConfig configures Paper server.
type PaperConfig struct {
	Version string `yaml:"version,omitempty"`
}

// PluginConfig is the configuration for a single plugin.
type PluginConfig struct {
	Source       string   `yaml:"source,omitempty"`
	Project      string   `yaml:"project,omitempty"`
	Version      string   `yaml:"version,omitempty"`
	Platform     string   `yaml:"platform,omitempty"`
	Loader       string   `yaml:"loader,omitempty"`
	GameVersions []string `yaml:"game_versions,omitempty"`
	Bucket       string   `yaml:"bucket,omitempty"`
	Key          string   `yaml:"key,omitempty"`
	URL          string   `yaml:"url,omitempty"`
}

// ToResolverConfig converts to resolver.PluginConfig.
func (p *PluginConfig) ToResolverConfig() map[string]interface{} {
	return map[string]interface{}{
		"source":        p.Source,
		"project":       p.Project,
		"version":       p.Version,
		"platform":      p.Platform,
		"loader":        p.Loader,
		"game_versions": p.GameVersions,
		"bucket":        p.Bucket,
		"key":           p.Key,
		"url":           p.URL,
	}
}
