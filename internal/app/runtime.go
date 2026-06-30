package app

import (
	"github.com/mshegolev/bqa-os/internal/runtime"
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
			return runtime.Detect()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "install-commands",
		Short: "Install project-local BQA Master command helpers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runtime.InstallCommands()
		},
	})

	return cmd
}
