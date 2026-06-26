package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build agents, skills, workflows, rules, guardrails, and registry from memory",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("TODO: build BQA artifacts from extracted knowledge")
			return nil
		},
	}
}
