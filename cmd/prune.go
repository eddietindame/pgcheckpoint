package cmd

import (
	"fmt"
	"os"
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
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		count, err := checkpoint.PruneCheckpoints()

		if err != nil {
			return fmt.Errorf("Error pruning checkpoints: %w\n", err)
		}

		fmt.Printf("Checkpoints pruned: %d\n", count)
		return nil
	},
}
