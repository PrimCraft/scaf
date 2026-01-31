package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const hangarAPIBase = "https://hangar.papermc.io/api/v1"

// HangarResolver resolves plugins from Hangar (PaperMC plugin repository).
type HangarResolver struct {
	client *http.Client
}

// NewHangarResolver creates a new Hangar resolver.
func NewHangarResolver(client *http.Client) *HangarResolver {
	return &HangarResolver{client: client}
}

func (h *HangarResolver) Name() string { return "hangar" }

func (h *HangarResolver) Resolve(ctx context.Context, cfg PluginConfig) (*Result, error) {
	platform := cfg.Platform
	if platform == "" {
		platform = "VELOCITY"
	}

	// Fetch all versions
	versions, err := h.fetchVersions(ctx, cfg.Project)
	if err != nil {
		return nil, fmt.Errorf("fetching versions: %w", err)
	}

	// Filter versions that have downloads for our platform
	var available []hangarVersion
	for _, v := range versions {
		if _, ok := v.Downloads[platform]; ok {
			available = append(available, v)
		}
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no versions found for %s on %s", cfg.Project, platform)
	}

	// Extract version strings
	versionStrings := make([]string, len(available))
	for i, v := range available {
		versionStrings[i] = v.Name
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
	var selected *hangarVersion
	for _, v := range available {
		if v.Name == selectedVersion {
			selected = &v
			break
		}
	}

	download := selected.Downloads[platform]

	return &Result{
		Source:     "hangar",
		Project:    cfg.Project,
		Version:    selected.Name,
		Platform:   platform,
		URL:        download.DownloadURL,
		SHA256:     download.FileInfo.SHA256Hash,
		ResolvedAt: time.Now().UTC(),
	}, nil
}

type hangarVersionsResponse struct {
	Pagination struct {
		Count  int `json:"count"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination"`
	Result []hangarVersion `json:"result"`
}

type hangarVersion struct {
	Name      string                       `json:"name"`
	Downloads map[string]hangarDownload    `json:"downloads"`
}

type hangarDownload struct {
	FileInfo struct {
		Name       string `json:"name"`
		SizeBytes  int64  `json:"sizeBytes"`
		SHA256Hash string `json:"sha256Hash"`
	} `json:"fileInfo"`
	DownloadURL string `json:"downloadUrl"`
}

func (h *HangarResolver) fetchVersions(ctx context.Context, project string) ([]hangarVersion, error) {
	url := fmt.Sprintf("%s/projects/%s/versions?limit=100", hangarAPIBase, project)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hangar API returned %d", resp.StatusCode)
	}

	var data hangarVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Result, nil
}
