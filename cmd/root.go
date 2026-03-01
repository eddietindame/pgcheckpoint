/*
Copyright © 2026 Eddie Tindame <eddie.tindame@googlemail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func checkDependencies() error {
	deps := []string{"pg_dump", "psql"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("%s not found in PATH", dep)
		}
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgcheckpoint",
	Short: "A tool for creating and restoring from Postgres database checkpoints.",
	Long:  ``,
	RunE:  createCmd.RunE,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var port int
var filename string

func init() {
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pgcheckpoint.yaml)")

	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 5432, "Postgres port for database connection.")

	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", "checkpoint_1.sql", "Filename for checkpoint")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
