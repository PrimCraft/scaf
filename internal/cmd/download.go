package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/PrimCraft/scaf/internal/manifest"
)

var (
	lockFile  string
	outputDir string
	parallel  int
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download plugins from lock file",
	Long:  `Download all plugins specified in the lock file to the output directory.`,
	RunE:  runDownload,
}

func init() {
	downloadCmd.Flags().StringVarP(&lockFile, "lockfile", "l", "plugins.lock.yaml", "Path to lock file")
	downloadCmd.Flags().StringVarP(&outputDir, "output", "o", "./plugins", "Output directory")
	downloadCmd.Flags().IntVarP(&parallel, "parallel", "p", 4, "Number of parallel downloads")
}

func runDownload(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Load lockfile
	data, err := os.ReadFile(lockFile)
	if err != nil {
		return fmt.Errorf("reading lock file: %w", err)
	}

	var lf manifest.Lockfile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return fmt.Errorf("parsing lock file: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}

	// Download Velocity if present
	if lf.Velocity != nil {
		fmt.Fprintf(os.Stderr, "Downloading Velocity %s...\n", lf.Velocity.Version)
		dest := filepath.Join(outputDir, "velocity.jar")
		if err := downloadHTTP(ctx, client, lf.Velocity.URL, dest); err != nil {
			return fmt.Errorf("downloading velocity: %w", err)
		}
		fmt.Fprintf(os.Stderr, "  -> %s\n", dest)
	}

	// Download Paper if present
	if lf.Paper != nil {
		fmt.Fprintf(os.Stderr, "Downloading Paper %s...\n", lf.Paper.Version)
		dest := filepath.Join(outputDir, "paper.jar")
		if err := downloadHTTP(ctx, client, lf.Paper.URL, dest); err != nil {
			return fmt.Errorf("downloading paper: %w", err)
		}
		fmt.Fprintf(os.Stderr, "  -> %s\n", dest)
	}

	// Download plugins
	for name, plugin := range lf.Plugins {
		fmt.Fprintf(os.Stderr, "Downloading %s %s...\n", name, plugin.Version)
		dest := filepath.Join(outputDir, name+".jar")

		var err error
		if plugin.S3URI != "" {
			err = downloadS3(ctx, plugin.S3URI, dest)
		} else if plugin.URL != "" {
			err = downloadHTTP(ctx, client, plugin.URL, dest)
		} else {
			err = fmt.Errorf("no download URL or S3 URI")
		}

		if err != nil {
			return fmt.Errorf("downloading %s: %w", name, err)
		}
		fmt.Fprintf(os.Stderr, "  -> %s\n", dest)
	}

	fmt.Fprintf(os.Stderr, "\nDownloaded to %s\n", outputDir)
	return nil
}

func downloadHTTP(ctx context.Context, client *http.Client, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func downloadS3(ctx context.Context, s3URI, dest string) error {
	// Parse s3://bucket/key
	var bucket, key string
	_, err := fmt.Sscanf(s3URI, "s3://%s", &bucket)
	if err != nil {
		return fmt.Errorf("invalid S3 URI: %s", s3URI)
	}

	// Split bucket and key
	for i, c := range bucket {
		if c == '/' {
			key = bucket[i+1:]
			bucket = bucket[:i]
			break
		}
	}

	// Load AWS config (uses default credential chain including OIDC)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, result.Body)
	return err
}
