package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coredoctor "github.com/mshegolev/bqa-os/internal/core/doctor"
	"github.com/spf13/cobra"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "doctor",
		Short:        "Check BQA-OS workspace health",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := fsadapter.BQAWorkspaceStore{}
			uc := coredoctor.UseCase{Inspector: store, RegistryReader: store}
			result, err := uc.Run(context.Background())
			if err != nil {
				return err
			}

			renderDoctorResult(cmd.OutOrStdout(), result)
			if !result.Healthy {
				return errors.New("doctor found workspace issues")
			}
			return nil
		},
	}
}

func renderDoctorResult(out io.Writer, result coredoctor.Result) {
	for _, check := range result.Checks {
		fmt.Fprintf(out, "%s %s %s - %s\n", check.Status, check.Name, check.Path, check.Message)
	}
}
