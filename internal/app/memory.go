package app

import (
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/core/memgov"
	"github.com/spf13/cobra"
)

func memoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Govern candidate QA memory (learn/review/promote/reject)",
		Long: "The memory governance loop extracts candidate memory from normalized sessions,\n" +
			"lets a human review it, and promotes or rejects candidates by id. Nothing enters\n" +
			"stable memory (approved_patterns.yaml) without an explicit promote.",
	}

	useCase := func(sessions, memoryDir string) memgov.UseCase {
		return memgov.UseCase{
			Reader:    fsadapter.KnowledgeStore{SessionBaseDir: sessions},
			Store:     fsadapter.GovernanceStore{},
			MemoryDir: memoryDir,
		}
	}

	var learnSessions, learnMemoryDir string
	learnCmd := &cobra.Command{
		Use:   "learn",
		Short: "Extract candidate memory from normalized sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase(learnSessions, learnMemoryDir).Learn(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Sessions processed: %d\n", res.SessionsProcessed)
			fmt.Fprintf(out, "Lessons added: %d\n", res.LessonsAdded)
			fmt.Fprintf(out, "Skill candidates added: %d\n", res.SkillsAdded)
			fmt.Fprintf(out, "Memory dir: %s\n", learnMemoryDir)
			return nil
		},
	}
	learnCmd.Flags().StringVar(&learnSessions, "sessions", ".bqa/input/sessions", "session input directory")
	learnCmd.Flags().StringVar(&learnMemoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(learnCmd)

	var reviewMemoryDir string
	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "List pending candidates awaiting a decision",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase("", reviewMemoryDir).Review(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(res.Pending) == 0 {
				fmt.Fprintln(out, "No pending candidates.")
				return nil
			}
			fmt.Fprintf(out, "Pending candidates: %d\n", len(res.Pending))
			for _, it := range res.Pending {
				fmt.Fprintf(out, "  %s [%s] %s\n", it.ID, it.Domain, it.Name)
			}
			return nil
		},
	}
	reviewCmd.Flags().StringVar(&reviewMemoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(reviewCmd)

	var promoteMemoryDir string
	promoteCmd := &cobra.Command{
		Use:   "promote <id>",
		Short: "Approve a pending candidate into approved_patterns.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase("", promoteMemoryDir).Promote(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s): %s\n", res.Action, res.Item.ID, res.Item.Domain, res.Item.Name)
			return nil
		},
	}
	promoteCmd.Flags().StringVar(&promoteMemoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(promoteCmd)

	var rejectMemoryDir string
	rejectCmd := &cobra.Command{
		Use:   "reject <id>",
		Short: "Reject a pending candidate into rejected_patterns.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase("", rejectMemoryDir).Reject(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s): %s\n", res.Action, res.Item.ID, res.Item.Domain, res.Item.Name)
			return nil
		},
	}
	rejectCmd.Flags().StringVar(&rejectMemoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(rejectCmd)

	return cmd
}
