package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	registryadapter "github.com/mshegolev/bqa-os/internal/adapters/registry"
	coreruntimeemit "github.com/mshegolev/bqa-os/internal/core/runtimeemit"
	"github.com/spf13/cobra"
)

func emitCmd() *cobra.Command {
	var target string
	var registryPath string
	var registryRoot string
	var format string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "emit",
		Short: "Emit native agents, skills, and workflows for Claude Code, Codex, and OpenCode from the BQA registry",
		Long: "Reads the BQA unified registry (team/brain/registry.json) and generates runtime-native\n" +
			"files into a target repository: .claude/ for Claude Code, .codex/ for Codex, and\n" +
			".opencode/ for OpenCode. Existing user files such as CLAUDE.md and root AGENTS.md are\n" +
			"never overwritten.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if target == "" {
				return fmt.Errorf("--target is required")
			}

			root := registryRoot
			if root == "" {
				resolved, err := deriveRegistryRoot(registryPath)
				if err != nil {
					return err
				}
				root = resolved
			}

			targets, err := parseTargets(format)
			if err != nil {
				return err
			}

			reader := registryadapter.Reader{RegistryPath: registryPath, Root: root}
			writer := fsadapter.RuntimeStore{TargetDir: target, DryRun: dryRun}
			uc := coreruntimeemit.UseCase{Reader: reader, Writer: writer, Targets: targets}

			result, err := uc.Run(context.Background())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			mode := "written"
			if dryRun {
				mode = "planned (dry-run)"
			}
			fmt.Fprintf(out, "Registry: %s\n", registryPath)
			fmt.Fprintf(out, "Target: %s\n", target)
			fmt.Fprintf(out, "Runtimes: %s\n", strings.Join(result.Targets, ", "))
			fmt.Fprintf(out, "Artifacts read: %d\n", result.ArtifactsRead)
			fmt.Fprintf(out, "Files %s: %d\n", mode, result.FilesWritten)
			for _, file := range result.Files {
				fmt.Fprintf(out, "  %s\n", file)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "target repository directory to write runtime files into (required)")
	cmd.Flags().StringVar(&registryPath, "registry", "team/brain/registry.json", "path to the BQA unified registry JSON")
	cmd.Flags().StringVar(&registryRoot, "registry-root", "", "root directory for resolving artifact sources (default: repo root inferred from --registry)")
	cmd.Flags().StringVar(&format, "format", "all", "comma-separated runtimes: claude, codex, opencode, or all")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "validate and list files without writing them")
	return cmd
}

// parseTargets expands the --format flag into a concrete runtime list.
func parseTargets(format string) ([]string, error) {
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" || format == "all" {
		return append([]string(nil), coreruntimeemit.SupportedTargets...), nil
	}
	var targets []string
	for _, part := range strings.Split(format, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			targets = append(targets, part)
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no runtimes selected in --format")
	}
	return targets, nil
}

// deriveRegistryRoot infers the source root from the registry path. The
// canonical layout is <root>/team/brain/registry.json, so the root is three
// directories up from the JSON file.
func deriveRegistryRoot(registryPath string) (string, error) {
	abs, err := filepath.Abs(registryPath)
	if err != nil {
		return "", err
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(abs)))
	if root == "" {
		return ".", nil
	}
	return root, nil
}
