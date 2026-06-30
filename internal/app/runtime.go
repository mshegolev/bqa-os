package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func runtimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Inspect and manage AI coding runtime adapters",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "detect",
		Short: "Detect supported AI coding CLIs in PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses, err := newRuntimeUseCase().Detect(context.Background())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, status := range statuses {
				if status.Found {
					fmt.Fprintf(out, "%-8s %s\n", status.Name, status.BinaryPath)
					continue
				}
				fmt.Fprintf(out, "%-8s missing\n", status.Name)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "install-commands",
		Short: "Install project-local BQA Master command helpers",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := newRuntimeUseCase().InstallCommands(context.Background())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, path := range res.Commands {
				fmt.Fprintf(out, "BQA runtime command written: %s\n", path)
			}
			fmt.Fprintf(out, "BQA master context generated: %s\n", res.ContextPath)
			fmt.Fprintln(out, "Claude Code can use /bqa-master in this project.")
			fmt.Fprintln(out, "Codex and OpenCode command helpers are available under .bqa/runtime-commands/.")
			return nil
		},
	})

	return cmd
}
