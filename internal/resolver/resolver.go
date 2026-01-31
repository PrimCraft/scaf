// Package resolver provides plugin version resolution from various sources.
package resolver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Result is the resolved plugin information.
type Result struct {
	Source     string    `yaml:"source"`
	Project    string    `yaml:"project,omitempty"`
	Version    string    `yaml:"version"`
	Build      int       `yaml:"build,omitempty"`
	Platform   string    `yaml:"platform,omitempty"`
	Loader     string    `yaml:"loader,omitempty"`
	URL        string    `yaml:"url,omitempty"`
	S3URI      string    `yaml:"s3_uri,omitempty"`
	SHA256     string    `yaml:"sha256,omitempty"`
	SHA512     string    `yaml:"sha512,omitempty"`
	ResolvedAt time.Time `yaml:"resolved_at"`
}

// PluginConfig is the input configuration from the manifest.
type PluginConfig struct {
	Source       string   `yaml:"source"`
	Project      string   `yaml:"project,omitempty"`
	Version      string   `yaml:"version"`
	Platform     string   `yaml:"platform,omitempty"`
	Loader       string   `yaml:"loader,omitempty"`
	GameVersions []string `yaml:"game_versions,omitempty"`
	Bucket       string   `yaml:"bucket,omitempty"`
	Key          string   `yaml:"key,omitempty"`
	URL          string   `yaml:"url,omitempty"`
}

// Resolver resolves a plugin from a specific source.
type Resolver interface {
	// Name returns the source name (e.g., "hangar", "modrinth").
	Name() string
	// Resolve fetches the best matching version for the given config.
	Resolve(ctx context.Context, cfg PluginConfig) (*Result, error)
}

// Registry holds all available resolvers.
type Registry struct {
	resolvers map[string]Resolver
	client    *http.Client
}

// NewRegistry creates a new resolver registry with default resolvers.
func NewRegistry() *Registry {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	r := &Registry{
		resolvers: make(map[string]Resolver),
		client:    client,
	}

	// Register built-in resolvers
	r.Register(NewHangarResolver(client))
	r.Register(NewModrinthResolver(client))
	r.Register(NewPaperMCResolver(client))
	r.Register(NewS3Resolver())
	r.Register(NewURLResolver())

	return r
}

// Register adds a resolver to the registry.
func (r *Registry) Register(res Resolver) {
	r.resolvers[res.Name()] = res
}

// Get returns a resolver by name.
func (r *Registry) Get(name string) (Resolver, bool) {
	res, ok := r.resolvers[name]
	return res, ok
}

// Resolve resolves a plugin using the appropriate resolver.
func (r *Registry) Resolve(ctx context.Context, source string, cfg PluginConfig) (*Result, error) {
	resolver, ok := r.Get(source)
	if !ok {
		return nil, fmt.Errorf("unknown source: %s", source)
	}
	return resolver.Resolve(ctx, cfg)
}

// Sources returns all registered source names.
func (r *Registry) Sources() []string {
	names := make([]string, 0, len(r.resolvers))
	for name := range r.resolvers {
		names = append(names, name)
	}
	return names
}
