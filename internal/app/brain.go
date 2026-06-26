package app

import (
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

	cmd.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Commit and push local BQA Brain cache changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return brain.Sync(false)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show BQA Brain connection status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return brain.Status()
		},
	})

	return cmd
}
