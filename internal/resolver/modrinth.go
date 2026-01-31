package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const modrinthAPIBase = "https://api.modrinth.com/v2"

// ModrinthResolver resolves plugins from Modrinth.
type ModrinthResolver struct {
	client *http.Client
}

// NewModrinthResolver creates a new Modrinth resolver.
func NewModrinthResolver(client *http.Client) *ModrinthResolver {
	return &ModrinthResolver{client: client}
}

func (m *ModrinthResolver) Name() string { return "modrinth" }

func (m *ModrinthResolver) Resolve(ctx context.Context, cfg PluginConfig) (*Result, error) {
	loader := cfg.Loader
	if loader == "" {
		loader = "velocity"
	}

	// Fetch versions
	versions, err := m.fetchVersions(ctx, cfg.Project, loader, cfg.GameVersions)
	if err != nil {
		return nil, fmt.Errorf("fetching versions: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %s on %s", cfg.Project, loader)
	}

	// Extract version strings
	versionStrings := make([]string, len(versions))
	for i, v := range versions {
		versionStrings[i] = v.VersionNumber
	}

	// Select best version based on constraint
	selectedVersion, err := SelectBestVersion(versionStrings, cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("selecting version: %w", err)
	}
	if selectedVersion == "" {
		return nil, fmt.Errorf("no version of %s matches constraint %q", cfg.Project, cfg.Version)
	}

	// Find the selected version data
	var selected *modrinthVersion
	for _, v := range versions {
		if v.VersionNumber == selectedVersion {
			selected = &v
			break
		}
	}

	// Get primary file or first file
	var file modrinthFile
	for _, f := range selected.Files {
		if f.Primary {
			file = f
			break
		}
	}
	if file.URL == "" && len(selected.Files) > 0 {
		file = selected.Files[0]
	}

	return &Result{
		Source:     "modrinth",
		Project:    cfg.Project,
		Version:    selected.VersionNumber,
		Loader:     loader,
		URL:        file.URL,
		SHA512:     file.Hashes.SHA512,
		SHA256:     file.Hashes.SHA256,
		ResolvedAt: time.Now().UTC(),
	}, nil
}

type modrinthVersion struct {
	VersionNumber string         `json:"version_number"`
	Files         []modrinthFile `json:"files"`
	Loaders       []string       `json:"loaders"`
	GameVersions  []string       `json:"game_versions"`
}

type modrinthFile struct {
	URL     string `json:"url"`
	Primary bool   `json:"primary"`
	Hashes  struct {
		SHA512 string `json:"sha512"`
		SHA256 string `json:"sha256"`
	} `json:"hashes"`
}

func (m *ModrinthResolver) fetchVersions(ctx context.Context, project, loader string, gameVersions []string) ([]modrinthVersion, error) {
	apiURL := fmt.Sprintf("%s/project/%s/version", modrinthAPIBase, project)

	// Build query params
	params := url.Values{}
	if loader != "" {
		loaders, _ := json.Marshal([]string{loader})
		params.Set("loaders", string(loaders))
	}
	if len(gameVersions) > 0 {
		gv, _ := json.Marshal(gameVersions)
		params.Set("game_versions", string(gv))
	}

	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mcpkg/1.0 (github.com/PrimCraft/mcpkg)")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth API returned %d", resp.StatusCode)
	}

	var versions []modrinthVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	return versions, nil
}
