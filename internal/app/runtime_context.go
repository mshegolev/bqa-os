package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/adapters/runtimebin"
	coreruntime "github.com/mshegolev/bqa-os/internal/core/runtime"
	"github.com/spf13/cobra"
)

func newRuntimeUseCase() coreruntime.UseCase {
	return coreruntime.UseCase{
		Writer:   fsadapter.RuntimeStore{TargetDir: "."},
		Detector: runtimebin.Detector{},
	}
}

func runtimeContextCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Prepare BQA Master Agent context for %s", runtimeLabel(name)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			res, err := newRuntimeUseCase().Prepare(ctx, name)
			if err != nil {
				return err
			}
			if name == "codex" {
				if err := augmentCodexContext(ctx, res.ContextPath); err != nil {
					return err
				}
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "BQA context generated: %s\n", res.ContextPath)
			if res.Detected {
				fmt.Fprintf(out, "Detected %s CLI: %s\n", res.Runtime, res.BinaryPath)
				fmt.Fprintf(out, "Next: start %s and paste or reference %s as the initial project instruction.\n", res.Command, res.ContextPath)
				return nil
			}
			fmt.Fprintf(out, "%s CLI was not found in PATH. Install it first, then run this command again.\n", res.Runtime)
			return nil
		},
	}
}

// augmentCodexContext appends the project-specific QA knowledge section to the
// already-written master context for the codex runtime. It reads the knowledge
// artifacts produced by `bqa build` via the existing KnowledgeStore reader and
// degrades gracefully (a "run bqa build" hint) when none are present.
func augmentCodexContext(ctx context.Context, contextPath string) error {
	base, err := os.ReadFile(filepath.Clean(contextPath))
	if err != nil {
		return err
	}

	reader := fsadapter.KnowledgeStore{}
	section, _ := codexKnowledgeSection(ctx, reader)

	// Write back to the exact path we just read, so the read and write targets
	// can never diverge.
	writer := fsadapter.RuntimeStore{TargetDir: "."}
	return writer.WriteRuntimeArtifact(ctx, contextPath, string(base)+section)
}

func runtimeLabel(name string) string {
	switch name {
	case "claude":
		return "Claude Code"
	case "codex":
		return "Codex CLI"
	case "opencode":
		return "OpenCode"
	default:
		return name
	}
}
