package cmd

import (
	"fmt"
	"pgcheckpoint/internal/checkpoint"
	"pgcheckpoint/internal/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		url := db.GetPgUrl(
			viper.GetString("db_user"),
			viper.GetString("db_password"),
			viper.GetString("db_host"),
			viper.GetInt("db_port"),
			viper.GetString("db_name"),
			viper.GetString("db_sslmode"),
		)
		fmt.Println("Database url:", url)

		out, restoredCheckpoint, err := checkpoint.RestoreCheckpoint(url, profile)

		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}

		if len(out) > 0 {
			fmt.Printf("\n%s\n", out)
		}
		fmt.Println("Checkpoint restored:", restoredCheckpoint)
		return nil
	},
}
