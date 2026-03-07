package cmd

import (
	"fmt"

	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
	"github.com/eddietindame/pgcheckpoint/internal/ui"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(renameCmd)
}

var renameCmd = &cobra.Command{
	Use:   "rename <checkpoint> <new-name>",
	Short: "Rename a checkpoint.",
	Long: `Change the name portion of an existing checkpoint file for the active profile.

The checkpoint can be specified by full filename (e.g. checkpoint_3_old-name.sql)
or by its short name (e.g. old-name).

Pass an empty string as the new name to remove the name entirely.

Examples:
  pgcheckpoint rename checkpoint_3_old-name.sql new-name
  pgcheckpoint rename old-name new-name
  pgcheckpoint rename checkpoint_3.sql "my name"
  pgcheckpoint rename checkpoint_3_old-name.sql ""`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := getNamingMode()
		if err != nil {
			return err
		}

		newFilename, err := checkpoint.RenameCheckpoint(getCheckpointDir(), profile, args[0], args[1], mode)
		if err != nil {
			return fmt.Errorf("error renaming checkpoint: %w", err)
		}

		ui.Success("Renamed", args[0]+" -> "+newFilename)
		return nil
	},
}
