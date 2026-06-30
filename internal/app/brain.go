package app

import (
	"fmt"

	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/spf13/cobra"
)

func brainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brain",
		Short: "Manage BQA Brain repository",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "connect <repo-url>",
		Short: "Connect to a BQA Brain git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return brain.Connect(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pull",
		Short: "Clone or update the local BQA Brain cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			return brain.Pull()
		},
	})

	var runSanitize bool
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Commit and push local BQA Brain cache changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return brain.Sync(runSanitize)
		},
	}
	syncCmd.Flags().BoolVar(&runSanitize, "sanitize", false, "sanitize brain cache before sync")
	cmd.AddCommand(syncCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show BQA Brain connection status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return brain.Status()
		},
	})

	var installFrom string
	var installTarget string
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install a generated BQA Brain package into a target client project",
		Long: "Copies the safe artifacts of a generated brain package (registry, agents, skills,\n" +
			"workflows, prompts, knowledge) into <target>/.bqa/. The source must be a valid brain\n" +
			"export and the target must be an existing directory. Unrelated files in the target are\n" +
			"never modified, and raw sessions or secrets are never copied.",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := brain.Install(installFrom, installTarget)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Source: %s\n", result.Source)
			fmt.Fprintf(out, "Target: %s\n", result.Target)
			fmt.Fprintf(out, "Installed into: %s\n", result.BqaDir)
			fmt.Fprintf(out, "Directories: %d, files: %d\n", len(result.Directories), len(result.Files))
			for _, file := range result.Files {
				fmt.Fprintf(out, "  %s\n", file)
			}
			return nil
		},
	}
	installCmd.Flags().StringVar(&installFrom, "from", "", "source brain package directory (required)")
	installCmd.Flags().StringVar(&installTarget, "target", "", "target client project directory (required)")
	cmd.AddCommand(installCmd)

	return cmd
}
