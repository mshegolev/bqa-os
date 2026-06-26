package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize .bqa workspace in current repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			dirs := []string{
				".bqa/input/sessions/raw",
				".bqa/input/sessions/normalized",
				".bqa/output",
				".bqa/registry",
				".bqa/memory",
				".bqa/agents",
				".bqa/skills",
				".bqa/workflows",
				".bqa/rules",
				".bqa/guardrails",
				".bqa/prompts",
			}
			for _, dir := range dirs {
				if err := os.MkdirAll(filepath.Clean(dir), 0o755); err != nil {
					return err
				}
			}
			fmt.Println("BQA workspace initialized in .bqa/")
			return nil
		},
	}
}
