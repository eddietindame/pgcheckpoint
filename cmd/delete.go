package cmd

import (
	"fmt"
	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete <checkpoint>",
	Short: "Delete a specific checkpoint.",
	Long: `Delete a specific checkpoint file by name for the active profile.

Unlike prune, which removes all but the latest checkpoint, delete removes
a single named checkpoint file (e.g. checkpoint_2.sql).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := getNamingMode()
		if err != nil {
			return err
		}

		filename, err := checkpoint.DeleteCheckpoint(getCheckpointDir(), profile, args[0], mode)
		if err != nil {
			return fmt.Errorf("error deleting checkpoint: %w", err)
		}

		fmt.Println("Checkpoint deleted:", filename)
		return nil
	},
}
