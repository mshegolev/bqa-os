package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "bqa",
		Short: "BQA-OS — Better QA Operating System",
		Long:  "BQA-OS connects QA knowledge, agents, skills, workflows, guardrails, and AI coding runtimes such as Codex, Claude Code, and OpenCode.",
	}

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(discoverCmd())
	rootCmd.AddCommand(ingestCmd())
	rootCmd.AddCommand(ingest2Cmd())
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(emitCmd())
	rootCmd.AddCommand(etlAgentPackCmd())
	rootCmd.AddCommand(demoCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(brainCmd())
	rootCmd.AddCommand(sanitizeCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(selfUpdateCmd())
	rootCmd.AddCommand(codexCmd())
	rootCmd.AddCommand(claudeCmd())
	rootCmd.AddCommand(opencodeCmd())
	rootCmd.AddCommand(runtimeCmd())
	rootCmd.AddCommand(teamCmd())
	rootCmd.AddCommand(doctorCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
