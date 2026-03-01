/*
Copyright © 2026 Eddie Tindame <eddie.tindame@googlemail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var (
	cfgFile    string
	port       int
	filename   string
	dbUser     string
	dbPassword string
	dbHost     string
	dbName     string
	dbSSLMode  string
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.pgcheckpoint.yaml)")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 5432, "Postgres port for database connection.")
	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", "checkpoint_1.sql", "Filename for checkpoint")
	rootCmd.PersistentFlags().StringVar(&dbUser, "db-user", "user", "Database user")
	rootCmd.PersistentFlags().StringVar(&dbPassword, "db-password", "password",
		"Database password")
	rootCmd.PersistentFlags().StringVar(&dbHost, "db-host", "localhost", "Database host")
	rootCmd.PersistentFlags().StringVar(&dbName, "db-name", "db", "Database name")
	rootCmd.PersistentFlags().StringVar(&dbSSLMode, "db-sslmode", "disable", "SSL mode")

	// Bind flags to viper keys so config file values become flag defaults
	viper.BindPFlag("db_port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("db_user", rootCmd.PersistentFlags().Lookup("db-user"))
	viper.BindPFlag("db_password", rootCmd.PersistentFlags().Lookup("db-password"))
	viper.BindPFlag("db_host", rootCmd.PersistentFlags().Lookup("db-host"))
	viper.BindPFlag("db_name", rootCmd.PersistentFlags().Lookup("db-name"))
	viper.BindPFlag("db_sslmode", rootCmd.PersistentFlags().Lookup("db-sslmode"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".pgcheckpoint" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".pgcheckpoint")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
