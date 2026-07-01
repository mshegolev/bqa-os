package app

import (
	"fmt"

	"github.com/mshegolev/bqa-os/internal/adapters/brainstore"
	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/mshegolev/bqa-os/internal/core/memory"
	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/spf13/cobra"
)

// coreMemoryUseCase aliases the memory use case for brevity in wiring.
type coreMemoryUseCase = memory.UseCase

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
	_ = installCmd.MarkFlagRequired("from")
	_ = installCmd.MarkFlagRequired("target")
	cmd.AddCommand(installCmd)

	memoryUseCase := func() coreMemoryUseCase {
		return coreMemoryUseCase{
			Source:    fsadapter.MemorySource{},
			Auditor:   fsadapter.MemoryAuditor{},
			Installer: fsadapter.MemoryInstaller{},
			Brain:     brainstore.GitBrainStore{},
			Writers:   map[string]ports.ArchiveWriter{"zip": fsadapter.ZipArchive{}, "tar": fsadapter.TarArchive{}},
			Readers:   map[string]ports.ArchiveReader{"zip": fsadapter.ZipArchive{}, "tar": fsadapter.TarArchive{}},
		}
	}

	var expSource, expTarget, expOut string
	var expDryRun, expStrict bool
	var expExclude []string
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export project memory to a zip/tar bundle or the GitHub brain",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := memoryUseCase().Export(cmd.Context(), memory.ExportOptions{
				SourceRoot: expSource, Target: expTarget, OutPath: expOut,
				Exclude: expExclude, DryRun: expDryRun, Strict: expStrict,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Target: %s\n", res.Target)
			fmt.Fprintf(out, "Files: %d\n", len(res.Files))
			fmt.Fprintf(out, "Audit: scanned=%d candidates=%d\n", res.Audit.FilesScanned, res.Audit.Candidates)
			if res.DryRun {
				fmt.Fprintln(out, "Dry run — no archive written. Planned files:")
				for _, f := range res.Files {
					fmt.Fprintf(out, "  %s\n", f)
				}
				return nil
			}
			if res.OutPath != "" {
				fmt.Fprintf(out, "Wrote: %s\n", res.OutPath)
			}
			return nil
		},
	}
	exportCmd.Flags().StringVar(&expSource, "source", ".bqa", "source memory directory")
	exportCmd.Flags().StringVar(&expTarget, "target", "zip", "target: zip|tar|github")
	exportCmd.Flags().StringVar(&expOut, "out", "", "output archive path (required for zip/tar)")
	exportCmd.Flags().BoolVar(&expDryRun, "dry-run", false, "print planned files without writing")
	exportCmd.Flags().BoolVar(&expStrict, "strict", false, "abort if the audit finds redaction candidates")
	exportCmd.Flags().StringArrayVar(&expExclude, "exclude", nil, "glob patterns to exclude (repeatable)")
	cmd.AddCommand(exportCmd)

	var impFrom, impTarget string
	var impDryRun bool
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import a memory bundle (verifies manifest + checksums, then installs)",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := memoryUseCase().Import(cmd.Context(), memory.ImportOptions{FromPath: impFrom, Target: impTarget, DryRun: impDryRun})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Verified: %v\n", res.Verified)
			if res.DryRun {
				fmt.Fprintln(out, "Dry run — nothing installed.")
				return nil
			}
			fmt.Fprintf(out, "Installed into: %s (%d files)\n", res.Installed.Target, len(res.Installed.Files))
			return nil
		},
	}
	importCmd.Flags().StringVar(&impFrom, "from", "", "bundle path (.zip or .tar, required)")
	importCmd.Flags().StringVar(&impTarget, "target", "", "target project directory (required)")
	importCmd.Flags().BoolVar(&impDryRun, "dry-run", false, "verify only; do not install")
	_ = importCmd.MarkFlagRequired("from")
	_ = importCmd.MarkFlagRequired("target")
	cmd.AddCommand(importCmd)

	return cmd
}
