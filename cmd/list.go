package cmd

import (
	"fmt"

	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
	"github.com/eddietindame/pgcheckpoint/internal/ui"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List checkpoints.",
	Long: `List all checkpoint files stored under the active profile's checkpoints
directory. Each checkpoint is displayed by filename, sorted by creation order.

If no checkpoints exist for the current profile, a message is displayed
indicating that none were found.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := checkpoint.ListCheckpointFilenames(getCheckpointDir(), profile)

		if err != nil {
			return fmt.Errorf("Error listing checkpoints: %w", err)
		}

		if len(files) == 0 {
			ui.Warn("No checkpoints.")
		}

		for _, file := range files {
			ui.ListItem(file)
		}

		return nil
	},
}
