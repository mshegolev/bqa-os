package app

import (
	"fmt"

	"github.com/mshegolev/bqa-os/internal/core/selfupdate"
	"github.com/mshegolev/bqa-os/internal/version"
	"github.com/spf13/cobra"
)

func selfUpdateCmd() *cobra.Command {
	var (
		checkOnly bool
		baseURL   string
	)
	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update BQA-OS client from GitHub Releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			updater := selfupdate.New()
			if baseURL != "" {
				updater.BaseURL = baseURL
			}

			fmt.Fprintf(out, "Current version: %s\n", version.Version)

			rel, err := updater.LatestRelease()
			if err != nil {
				return fmt.Errorf("could not resolve latest release: %w", err)
			}
			fmt.Fprintf(out, "Latest version:  %s\n", rel.TagName)

			if !selfupdate.IsNewer(version.Version, rel.TagName) {
				fmt.Fprintln(out, "Already up to date.")
				return nil
			}

			if version.Version == "dev" {
				fmt.Fprintln(out, "Running a dev build; a release is available.")
			} else {
				fmt.Fprintln(out, "An update is available.")
			}

			if checkOnly {
				fmt.Fprintf(out, "Run `bqa self-update` to install %s.\n", rel.TagName)
				return nil
			}

			fmt.Fprintf(out, "Downloading %s ...\n", rel.AssetName)
			path, err := updater.Apply(rel)
			if err != nil {
				return fmt.Errorf("update failed: %w", err)
			}
			fmt.Fprintf(out, "Updated %s to %s.\n", path, rel.TagName)
			fmt.Fprintln(out, "Run `bqa version` to confirm.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "only report current vs latest version without downloading")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "override the GitHub API base URL (testing/self-hosted)")
	_ = cmd.Flags().MarkHidden("base-url")
	return cmd
}
