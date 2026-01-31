package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const paperMCAPIBase = "https://api.papermc.io/v2"

// PaperMCResolver resolves Velocity/Paper/etc from PaperMC API.
type PaperMCResolver struct {
	client *http.Client
}

// NewPaperMCResolver creates a new PaperMC resolver.
func NewPaperMCResolver(client *http.Client) *PaperMCResolver {
	return &PaperMCResolver{client: client}
}

func (p *PaperMCResolver) Name() string { return "papermc" }

func (p *PaperMCResolver) Resolve(ctx context.Context, cfg PluginConfig) (*Result, error) {
	project := cfg.Project
	if project == "" {
		project = "velocity"
	}

	// Get all versions for the project
	versions, err := p.fetchVersions(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("fetching versions: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %s", project)
	}

	// Select version based on constraint
	selectedVersion, err := SelectBestVersion(versions, cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("selecting version: %w", err)
	}
	if selectedVersion == "" {
		return nil, fmt.Errorf("no version of %s matches constraint %q", project, cfg.Version)
	}

	// Get latest build for this version
	build, downloadName, err := p.fetchLatestBuild(ctx, project, selectedVersion)
	if err != nil {
		return nil, fmt.Errorf("fetching build: %w", err)
	}

	downloadURL := fmt.Sprintf("%s/projects/%s/versions/%s/builds/%d/downloads/%s",
		paperMCAPIBase, project, selectedVersion, build, downloadName)

	return &Result{
		Source:     "papermc",
		Project:    project,
		Version:    selectedVersion,
		Build:      build,
		URL:        downloadURL,
		ResolvedAt: time.Now().UTC(),
	}, nil
}

type paperMCProjectResponse struct {
	Versions []string `json:"versions"`
}

type paperMCBuildsResponse struct {
	Builds []paperMCBuild `json:"builds"`
}

type paperMCBuild struct {
	Build     int `json:"build"`
	Downloads struct {
		Application struct {
			Name   string `json:"name"`
			SHA256 string `json:"sha256"`
		} `json:"application"`
	} `json:"downloads"`
}

func (p *PaperMCResolver) fetchVersions(ctx context.Context, project string) ([]string, error) {
	url := fmt.Sprintf("%s/projects/%s", paperMCAPIBase, project)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PaperMC API returned %d", resp.StatusCode)
	}

	var data paperMCProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Reverse to get newest first (API returns oldest first)
	versions := data.Versions
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	return versions, nil
}

func (p *PaperMCResolver) fetchLatestBuild(ctx context.Context, project, version string) (int, string, error) {
	url := fmt.Sprintf("%s/projects/%s/versions/%s/builds", paperMCAPIBase, project, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("PaperMC API returned %d", resp.StatusCode)
	}

	var data paperMCBuildsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, "", err
	}

	if len(data.Builds) == 0 {
		return 0, "", fmt.Errorf("no builds found for %s %s", project, version)
	}

	// Latest build is last in the array
	latest := data.Builds[len(data.Builds)-1]
	return latest.Build, latest.Downloads.Application.Name, nil
}
