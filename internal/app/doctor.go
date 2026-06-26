package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check BQA-OS workspace health",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("BQA doctor: TODO checks for .bqa workspace, registry, memory, agents, skills, workflows")
			return nil
		},
	}
}
