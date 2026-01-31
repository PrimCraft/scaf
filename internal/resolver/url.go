package resolver

import (
	"context"
	"fmt"
	"time"
)

// URLResolver resolves plugins from direct URLs.
type URLResolver struct{}

// NewURLResolver creates a new URL resolver.
func NewURLResolver() *URLResolver {
	return &URLResolver{}
}

func (u *URLResolver) Name() string { return "url" }

func (u *URLResolver) Resolve(ctx context.Context, cfg PluginConfig) (*Result, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("url source requires 'url' field")
	}

	version := cfg.Version
	if version == "" {
		version = "unknown"
	}

	return &Result{
		Source:     "url",
		Version:    version,
		URL:        cfg.URL,
		ResolvedAt: time.Now().UTC(),
	}, nil
}
