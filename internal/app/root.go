package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "bqa",
		Short: "BQA-OS — Big Data QA Operating System",
		Long:  "BQA-OS is a local-first agent operating system for Big Data QA, ETL testing, session ingestion, memory, skills, workflows, and master-agent orchestration.",
	}

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(discoverCmd())
	rootCmd.AddCommand(ingestCmd())
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(codexCmd())
	rootCmd.AddCommand(doctorCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
