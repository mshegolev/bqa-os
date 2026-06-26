package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [task]",
		Short: "Run BQA Master Agent task plan locally",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			task := strings.TrimSpace(strings.Join(args, " "))
			if task == "" {
				task = "Протестируй ETL в текущем проекте"
			}
			fmt.Printf("BQA Master task: %s\n", task)
			fmt.Println("TODO: load registry, select agents, create execution plan, produce report")
			return nil
		},
	}
}
