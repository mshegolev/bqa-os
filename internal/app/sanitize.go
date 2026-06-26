package app

import (
	"fmt"

	"github.com/mshegolev/bqa-os/internal/sanitize"
	"github.com/spf13/cobra"
)

func sanitizeCmd() *cobra.Command {
	var write bool
	cmd := &cobra.Command{
		Use:   "sanitize [path]",
		Short: "Scan and redact sensitive values in text files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			result, err := sanitize.Path(path, write)
			if err != nil {
				return err
			}
			fmt.Printf("Sanitize scanned=%d changed=%d redactions=%d write=%v\n", result.FilesScanned, result.FilesChanged, result.Redactions, write)
			if !write && result.Redactions > 0 {
				fmt.Println("Dry run only. Re-run with --write to apply redactions.")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&write, "write", false, "apply redactions in place")
	return cmd
}
