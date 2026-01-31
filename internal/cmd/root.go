// Package cmd implements the CLI commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set by goreleaser
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "scaf",
	Short: "Scaffolding for Minecraft servers",
	Long: `scaf resolves and downloads Minecraft plugins from various sources.

Supported sources:
  - hangar    PaperMC Hangar plugin repository
  - modrinth  Modrinth mod/plugin repository
  - papermc   PaperMC API (Velocity, Paper, etc.)
  - s3        AWS S3 buckets
  - url       Direct URLs

Example manifest (plugins.yaml):
  velocity:
    version: "~=3.4"

  plugins:
    tab:
      source: hangar
      project: NEZNAMY/TAB
      version: ">=5.0.0"
      platform: VELOCITY`,
	Version: Version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(changelogCmd)
}
