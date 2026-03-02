package cmd

import (
	"fmt"
	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove all but the latest checkpoint.",
	Long: `Remove all checkpoint files for the active profile except the most recent
one. This frees up disk space while keeping the latest checkpoint available
for restore.

Use --naming-mode to match the naming convention of your checkpoints
(sequential, timestamp, compact, or unix). Displays the number of
checkpoints that were removed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkDependencies(); err != nil {
			return fmt.Errorf("error: %v\n", err)
		}

		mode, err := getNamingMode()
		if err != nil {
			return err
		}

		count, err := checkpoint.PruneCheckpoints(getCheckpointDir(), profile, mode)

		if err != nil {
			return fmt.Errorf("error pruning checkpoints: %w", err)
		}

		fmt.Println("Checkpoints pruned:", count)
		return nil
	},
}
