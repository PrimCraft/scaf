package resolver

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ParseVersion attempts to parse a version string into a semver.Version.
// It handles common Minecraft version formats.
func ParseVersion(version string) (*semver.Version, error) {
	// Remove leading 'v' if present
	clean := strings.TrimPrefix(version, "v")

	// Handle versions like "5.0.4-pre1" -> "5.0.4-pre1" (semver handles this)
	// Handle versions like "5.0.4-SNAPSHOT" -> "5.0.4-SNAPSHOT"

	v, err := semver.NewVersion(clean)
	if err != nil {
		// Try extracting just X.Y.Z
		re := regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)?)`)
		if match := re.FindString(clean); match != "" {
			return semver.NewVersion(match)
		}
		return nil, err
	}
	return v, nil
}

// ParseConstraint parses a version constraint string.
// If the string looks like a plain version (e.g., "5.0.0"), it's treated as exact match.
//
// To include prereleases, use the -0 suffix convention:
//   - "~3.4"     matches >=3.4.0, <3.5.0 (stable only)
//   - "~3.4.0-0" matches >=3.4.0-0, <3.5.0 (includes prereleases like SNAPSHOT)
//   - ">=5.0.0-0" matches any version >=5.0.0 including prereleases
func ParseConstraint(constraint string) (*semver.Constraints, error) {
	if constraint == "" || constraint == "latest" {
		return nil, nil
	}

	// If it looks like a plain version, treat as exact match
	if matched, _ := regexp.MatchString(`^\d+\.\d+`, constraint); matched {
		if !strings.ContainsAny(constraint, "=<>~^,| -") {
			constraint = "=" + constraint
		}
	}

	return semver.NewConstraint(constraint)
}

// FilterVersions filters and sorts versions by constraint, returning newest first.
func FilterVersions(versions []string, constraint string) ([]string, error) {
	if constraint == "" || constraint == "latest" {
		return versions, nil
	}

	c, err := ParseConstraint(constraint)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return versions, nil
	}

	type versionPair struct {
		original string
		parsed   *semver.Version
	}

	var matched []versionPair
	for _, v := range versions {
		sv, err := ParseVersion(v)
		if err != nil {
			continue // Skip unparseable versions
		}
		if c.Check(sv) {
			matched = append(matched, versionPair{v, sv})
		}
	}

	// Sort by semver descending (newest first)
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].parsed.GreaterThan(matched[j].parsed)
	})

	result := make([]string, len(matched))
	for i, v := range matched {
		result[i] = v.original
	}
	return result, nil
}

// SelectBestVersion selects the best version from a list based on constraint.
func SelectBestVersion(versions []string, constraint string) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}

	if constraint == "" || constraint == "latest" {
		return versions[0], nil
	}

	filtered, err := FilterVersions(versions, constraint)
	if err != nil {
		return "", err
	}
	if len(filtered) == 0 {
		return "", nil
	}
	return filtered[0], nil
}
