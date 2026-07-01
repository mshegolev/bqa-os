package app

import (
	"fmt"
	"os"
	"path/filepath"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/core/workspace"
	"github.com/spf13/cobra"
)

func workspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage the BQA workspace registry (projects for task work)",
		Long: "The workspace registry records local repo/worktree roots in .bqa/workspace.yaml.\n" +
			"Later commands (bqa task start) create task worktrees against these projects.\n" +
			"Registering a project never copies or modifies the target repository.",
	}

	newUseCase := func(baseDir string) workspace.UseCase {
		return workspace.UseCase{
			Store:     fsadapter.WorkspaceStore{},
			Inspector: fsadapter.PathInspector{},
			BaseDir:   baseDir,
		}
	}

	var initName, initBaseDir string
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a BQA workspace registry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := initName
			if name == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				name = filepath.Base(wd)
			}
			res, err := newUseCase(initBaseDir).Init(cmd.Context(), name)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized workspace %q at %s\n", res.Name, res.Path)
			return nil
		},
	}
	initCmd.Flags().StringVar(&initName, "name", "", "workspace name (default: current directory name)")
	initCmd.Flags().StringVar(&initBaseDir, "base-dir", ".bqa", "BQA base directory")
	cmd.AddCommand(initCmd)

	var addRepo, addETL, addBranchRole, addBaseDir string
	addCmd := &cobra.Command{
		Use:   "add <project-id> <path>",
		Short: "Register a local repo/worktree root (does not copy it)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := newUseCase(addBaseDir).Add(cmd.Context(), workspace.Project{
				ID: args[0], Path: args[1], Repo: addRepo, ETL: addETL, BranchRole: addBranchRole,
			})
			if err != nil {
				return err
			}
			if res.Warning != "" {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning: "+res.Warning)
			}
			p := res.Project
			fmt.Fprintf(cmd.OutOrStdout(), "Added project %s (repo=%s etl=%s branch_role=%s): %s\n", p.ID, p.Repo, p.ETL, p.BranchRole, p.Path)
			return nil
		},
	}
	addCmd.Flags().StringVar(&addRepo, "repo", "", "repository name (required)")
	addCmd.Flags().StringVar(&addETL, "etl", "", "ETL pipeline this project targets")
	addCmd.Flags().StringVar(&addBranchRole, "branch-role", workspace.DefaultBranchRole, "branch role for this project")
	addCmd.Flags().StringVar(&addBaseDir, "base-dir", ".bqa", "BQA base directory")
	_ = addCmd.MarkFlagRequired("repo")
	cmd.AddCommand(addCmd)

	var listBaseDir string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List registered projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := newUseCase(listBaseDir).List(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			ws := res.Workspace
			fmt.Fprintf(out, "Workspace: %s\n", ws.Name)
			if len(ws.Projects) == 0 {
				fmt.Fprintln(out, "No projects registered.")
			} else {
				fmt.Fprintf(out, "Projects: %d\n", len(ws.Projects))
				for _, p := range ws.Projects {
					fmt.Fprintf(out, "  %s  repo=%s  etl=%s  branch_role=%s  %s\n", p.ID, p.Repo, p.ETL, p.BranchRole, p.Path)
				}
			}
			fmt.Fprintf(out, "Tasks: %d\n", len(ws.Tasks))
			return nil
		},
	}
	listCmd.Flags().StringVar(&listBaseDir, "base-dir", ".bqa", "BQA base directory")
	cmd.AddCommand(listCmd)

	return cmd
}
