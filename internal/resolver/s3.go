package resolver

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// S3Resolver resolves plugins from S3 buckets.
type S3Resolver struct{}

// NewS3Resolver creates a new S3 resolver.
func NewS3Resolver() *S3Resolver {
	return &S3Resolver{}
}

func (s *S3Resolver) Name() string { return "s3" }

func (s *S3Resolver) Resolve(ctx context.Context, cfg PluginConfig) (*Result, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 source requires 'bucket' field")
	}
	if cfg.Key == "" {
		return nil, fmt.Errorf("s3 source requires 'key' field")
	}

	version := cfg.Version
	if version == "" {
		version = "latest"
	}

	// Substitute ${version} placeholder in key
	key := cfg.Key
	if strings.Contains(key, "${version}") {
		if version == "latest" {
			return nil, fmt.Errorf("s3 source with ${version} placeholder requires explicit version, not 'latest'")
		}
		key = strings.ReplaceAll(key, "${version}", version)
	}

	s3URI := fmt.Sprintf("s3://%s/%s", cfg.Bucket, key)

	return &Result{
		Source:     "s3",
		Version:    version,
		S3URI:      s3URI,
		ResolvedAt: time.Now().UTC(),
	}, nil
}
