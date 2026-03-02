/*
Copyright © 2026 Eddie Tindame <eddie.tindame@googlemail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
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

func getCheckpointDir() string {
	if dir := viper.GetString("checkpoint_dir"); dir != "" {
		return dir
	}
	return checkpoint.DefaultCheckpointDir()
}

func getNamingMode() string {
	return viper.GetString("naming_mode")
}

func viperKeyToFlagName(key string) string {
	return strings.ReplaceAll(key, "_", "-")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgcheckpoint",
	Short: "A tool for creating and restoring from Postgres database checkpoints.",
	Long: `pgcheckpoint is a CLI tool for creating and restoring PostgreSQL database
checkpoints using pg_dump and psql.

It supports global and project-level configuration files, config profiles,
and can be customised with flags or environment variables. Checkpoints can
be named sequentially (checkpoint_1.sql) or with timestamps
(checkpoint_2026-03-02_15-30-45.sql) via the --naming-mode flag.

When called without a subcommand, it defaults to creating a new checkpoint.

Requires pg_dump and psql to be available in your PATH.`,
	RunE: createCmd.RunE,
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
	cfgFile        string
	projectCfgFile string
	profile        string
	port           int
	filename       string
	dbUser         string
	dbPassword     string
	dbHost         string
	dbName         string
	dbSSLMode      string
	checkpointDir  string
	namingMode     string
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "global config file (default is $HOME/.pgcheckpoint.yaml)")
	rootCmd.PersistentFlags().StringVarP(&projectCfgFile, "project-config", "j", "", "config file (default is ./.pgcheckpoint.yaml)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "default", "config profile to use")

	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 5432, "Postgres port for database connection.")
	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", "checkpoint_1.sql", "Filename for checkpoint")
	rootCmd.PersistentFlags().StringVar(&dbUser, "db-user", "user", "Database user")
	rootCmd.PersistentFlags().StringVar(&dbPassword, "db-password", "password",
		"Database password")
	rootCmd.PersistentFlags().StringVar(&dbHost, "db-host", "localhost", "Database host")
	rootCmd.PersistentFlags().StringVar(&dbName, "db-name", "db", "Database name")
	rootCmd.PersistentFlags().StringVar(&dbSSLMode, "db-sslmode", "disable", "SSL mode")
	rootCmd.PersistentFlags().StringVar(&checkpointDir, "checkpoint-dir", "", "Checkpoint storage directory (default ~/.pgcheckpoint/checkpoints)")
	rootCmd.PersistentFlags().StringVar(&namingMode, "naming-mode", "sequential", "Checkpoint naming mode (sequential or timestamp)")

	// Bind flags to viper keys so config file values become flag defaults
	viper.BindPFlag("db_port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("db_user", rootCmd.PersistentFlags().Lookup("db-user"))
	viper.BindPFlag("db_password", rootCmd.PersistentFlags().Lookup("db-password"))
	viper.BindPFlag("db_host", rootCmd.PersistentFlags().Lookup("db-host"))
	viper.BindPFlag("db_name", rootCmd.PersistentFlags().Lookup("db-name"))
	viper.BindPFlag("db_sslmode", rootCmd.PersistentFlags().Lookup("db-sslmode"))
	viper.BindPFlag("checkpoint_dir", rootCmd.PersistentFlags().Lookup("checkpoint-dir"))
	viper.BindPFlag("naming_mode", rootCmd.PersistentFlags().Lookup("naming-mode"))
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Global config
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".pgcheckpoint")
		viper.AddConfigPath(home)                                           // ~/.pgcheckpoint.yml
		viper.AddConfigPath(filepath.Join(home, ".pgcheckpoint"))           // ~/.pgcheckpoint/.pgcheckpoint.yml
		viper.AddConfigPath(filepath.Join(home, ".config"))                 // ~/.config/.pgcheckpoint.yml
		viper.AddConfigPath(filepath.Join(home, ".config", "pgcheckpoint")) // ~/.config/pgcheckpoint/.pgcheckpoint.yml
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			cobra.CheckErr(fmt.Errorf("error reading global config: %w", err))
		}
	}

	// Project config (overrides global)
	if projectCfgFile != "" {
		viper.SetConfigFile(projectCfgFile)
	} else {
		viper.SetConfigName(".pgcheckpoint")
		viper.AddConfigPath(".")
	}

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			cobra.CheckErr(fmt.Errorf("error reading project config: %w", err))
		}
	}

	// Apply profile — only set values that weren't explicitly passed as flags
	sub := viper.Sub(profile)
	if sub != nil {
		for _, key := range sub.AllKeys() {
			flag := rootCmd.PersistentFlags().Lookup(viperKeyToFlagName(key))
			if flag == nil || !flag.Changed {
				viper.Set(key, sub.Get(key))
			}
		}
	} else if profile != "default" {
		cobra.CheckErr(fmt.Errorf("profile %q not found in config", profile))
	}

	viper.AutomaticEnv()
}
