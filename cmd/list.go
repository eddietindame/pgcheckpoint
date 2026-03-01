package cmd

import (
	"fmt"
	"pgcheckpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List checkpoints.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := checkpoint.ListCheckpointFilenames(profile)

		if err != nil {
			return fmt.Errorf("Error listing checkpoints: %w", err)
		}

		if len(files) == 0 {
			fmt.Println("No checkpoints.")
		}

		for _, file := range files {
			fmt.Println(file)
		}

		return nil
	},
}
