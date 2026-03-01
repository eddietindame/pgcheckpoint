package cmd

import (
	"fmt"
	"pgcheckpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database to latest checkpoint.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkDependencies(); err != nil {
			return fmt.Errorf("error: %v\n", err)
		}

		fmt.Println("Database url:", checkpoint.GetPgUrl(port))

		out, restoredCheckpoint, err := checkpoint.RestoreCheckpoint(port)

		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}

		fmt.Println(out)
		fmt.Println("Checkpoint restored:", restoredCheckpoint)
		return nil
	},
}
