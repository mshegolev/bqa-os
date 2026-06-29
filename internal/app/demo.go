package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coredemoarchive "github.com/mshegolev/bqa-os/internal/core/demoarchive"
	"github.com/spf13/cobra"
)

func demoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Generate synthetic BQA-OS demo assets",
	}
	cmd.AddCommand(demoArchiveCmd())
	return cmd
}

func demoArchiveCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Generate a synthetic demo archive for static site upload",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			writer := fsadapter.DemoArchiveWriter{}
			uc := coredemoarchive.UseCase{Writer: writer}

			result, err := uc.Run(ctx, outputPath)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Demo archive written: %s\n", result.OutputPath)
			fmt.Fprintf(out, "Files included: %d\n", result.FilesCreated)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", ".bqa/output/demo-archive.zip", "demo archive output zip path")
	return cmd
}
