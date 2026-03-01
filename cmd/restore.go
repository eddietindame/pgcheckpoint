package cmd

import (
	"fmt"
	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
	"github.com/eddietindame/pgcheckpoint/internal/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore [checkpoint]",
	Short: "Restore database to a checkpoint.",
	Long: `Restore the configured PostgreSQL database to a checkpoint by executing
the checkpoint SQL file using psql. If a checkpoint filename is provided
as an argument (e.g. checkpoint_2.sql), that checkpoint is restored.
Otherwise the latest checkpoint for the active profile is used.

This will overwrite the current state of the database with the contents
of the checkpoint file.`,
	Args: cobra.MaximumNArgs(1),
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

		var target string
		if len(args) > 0 {
			target = args[0]
		}

		out, restoredCheckpoint, err := checkpoint.RestoreCheckpoint(url, getCheckpointDir(), profile, target)

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
