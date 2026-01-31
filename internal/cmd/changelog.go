package cmd

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/PrimCraft/scaf/internal/manifest"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog <old-lockfile> <new-lockfile>",
	Short: "Generate changelog between two lock files",
	Long:  `Compare two lock files and output the differences in markdown format.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runChangelog,
}

func runChangelog(cmd *cobra.Command, args []string) error {
	oldPath := args[0]
	newPath := args[1]

	// Load old lockfile
	oldData, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("reading old lock file: %w", err)
	}

	var oldLock manifest.Lockfile
	if err := yaml.Unmarshal(oldData, &oldLock); err != nil {
		return fmt.Errorf("parsing old lock file: %w", err)
	}

	// Load new lockfile
	newData, err := os.ReadFile(newPath)
	if err != nil {
		return fmt.Errorf("reading new lock file: %w", err)
	}

	var newLock manifest.Lockfile
	if err := yaml.Unmarshal(newData, &newLock); err != nil {
		return fmt.Errorf("parsing new lock file: %w", err)
	}

	// Compare and output changes
	hasChanges := false

	// Check Velocity
	if oldLock.Velocity != nil || newLock.Velocity != nil {
		oldVer := "N/A"
		oldBuild := 0
		newVer := "N/A"
		newBuild := 0

		if oldLock.Velocity != nil {
			oldVer = oldLock.Velocity.Version
			oldBuild = oldLock.Velocity.Build
		}
		if newLock.Velocity != nil {
			newVer = newLock.Velocity.Version
			newBuild = newLock.Velocity.Build
		}

		if oldVer != newVer || oldBuild != newBuild {
			oldStr := formatVersion(oldVer, oldBuild)
			newStr := formatVersion(newVer, newBuild)
			fmt.Printf("- **Velocity**: %s -> %s\n", oldStr, newStr)
			hasChanges = true
		}
	}

	// Check Paper
	if oldLock.Paper != nil || newLock.Paper != nil {
		oldVer := "N/A"
		oldBuild := 0
		newVer := "N/A"
		newBuild := 0

		if oldLock.Paper != nil {
			oldVer = oldLock.Paper.Version
			oldBuild = oldLock.Paper.Build
		}
		if newLock.Paper != nil {
			newVer = newLock.Paper.Version
			newBuild = newLock.Paper.Build
		}

		if oldVer != newVer || oldBuild != newBuild {
			oldStr := formatVersion(oldVer, oldBuild)
			newStr := formatVersion(newVer, newBuild)
			fmt.Printf("- **Paper**: %s -> %s\n", oldStr, newStr)
			hasChanges = true
		}
	}

	// Check for updated plugins
	for name, newPlugin := range newLock.Plugins {
		oldPlugin, existed := oldLock.Plugins[name]

		if !existed {
			fmt.Printf("- **%s**: added (%s)\n", name, newPlugin.Version)
			hasChanges = true
		} else if oldPlugin.Version != newPlugin.Version {
			fmt.Printf("- **%s**: %s -> %s\n", name, oldPlugin.Version, newPlugin.Version)
			hasChanges = true
		}
	}

	// Check for removed plugins
	for name, oldPlugin := range oldLock.Plugins {
		if _, exists := newLock.Plugins[name]; !exists {
			fmt.Printf("- **%s**: removed (was %s)\n", name, oldPlugin.Version)
			hasChanges = true
		}
	}

	if !hasChanges {
		fmt.Println("No changes detected")
	}

	return nil
}

func formatVersion(version string, build int) string {
	if build > 0 {
		return fmt.Sprintf("%s (build %d)", version, build)
	}
	return version
}
