package app

import (
	"fmt"

	"github.com/mshegolev/bqa-os/internal/core/doctor"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	var baseDir string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check BQA-OS workspace health",
		RunE: func(cmd *cobra.Command, args []string) error {
			report := doctor.Run(baseDir)
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "BQA doctor — workspace %q\n", baseDir)
			for _, c := range report.Checks {
				mark := "ok  "
				if !c.OK {
					mark = "FAIL"
				}
				fmt.Fprintf(out, "  [%s] %s\n", mark, c.Detail)
			}
			if !report.OK {
				return fmt.Errorf("workspace check failed: run `bqa init` to create missing directories")
			}
			fmt.Fprintln(out, "All checks passed.")
			return nil
		},
	}
	cmd.Flags().StringVar(&baseDir, "base-dir", ".bqa", "workspace base directory")
	return cmd
}
