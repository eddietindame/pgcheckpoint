package cmd

import (
	"fmt"

	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
	"github.com/eddietindame/pgcheckpoint/internal/db"
	"github.com/eddietindame/pgcheckpoint/internal/ui"

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
the checkpoint SQL file using psql. If a checkpoint is provided as an
argument, that checkpoint is restored. Otherwise the latest checkpoint
for the active profile is used.

The checkpoint can be specified by full filename (e.g. checkpoint_2.sql)
or by its short name (e.g. before-migration).

Use --naming-mode to match the naming convention of your checkpoints
(sequential, timestamp, compact, or unix) when restoring the latest.

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
		ui.Info("Database url:", url)

		var target string
		if len(args) > 0 {
			target = args[0]
		}

		mode, err := getNamingMode()
		if err != nil {
			return err
		}

		out, restoredCheckpoint, err := checkpoint.RestoreCheckpoint(url, getCheckpointDir(), profile, target, mode)

		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}

		if len(out) > 0 {
			ui.Detail(out)
		}
		ui.Success("Checkpoint restored:", restoredCheckpoint)
		return nil
	},
}
