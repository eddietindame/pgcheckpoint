package cmd

import (
	"fmt"
	"pgcheckpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove all but the latest checkpoint.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkDependencies(); err != nil {
			return fmt.Errorf("error: %v\n", err)
		}

		count, err := checkpoint.PruneCheckpoints(profile)

		if err != nil {
			return fmt.Errorf("Error pruning checkpoints: %w", err)
		}

		fmt.Println("Checkpoints pruned:", count)
		return nil
	},
}
