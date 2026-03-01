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
		files, err := checkpoint.ListCheckpointFilenames()

		if err != nil {
			return err
		}

		for _, file := range files {
			fmt.Println(file)
		}

		return nil
	},
}
