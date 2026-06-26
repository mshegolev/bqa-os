package app

import (
	"fmt"

	"github.com/mshegolev/bqa-os/internal/version"
	"github.com/spf13/cobra"
)

func selfUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "self-update",
		Short: "Show how to update BQA-OS client",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Current version: %s\n", version.Version)
			fmt.Println("GitHub Release based self-update is not available yet.")
			fmt.Println("For now, update with:")
			fmt.Println("curl -fsSL https://raw.githubusercontent.com/mshegolev/bqa-os/main/install.sh | bash")
			fmt.Println("hash -r")
			fmt.Println("bqa version")
		},
	}
}
