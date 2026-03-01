package cmd

import (
	"fmt"
	"os"
	"pgcheckpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new checkpoint.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkDependencies(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Database url: %s\n", checkpoint.GetPgUrl(port))

		out, err := checkpoint.CreateCheckpoint(filename, port)

		if err != nil {
			return err
		}

		fmt.Println(out)
		return nil
	},
}
