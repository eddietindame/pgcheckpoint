package cmd

import (
	"fmt"

	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
	"github.com/eddietindame/pgcheckpoint/internal/db"
	"github.com/eddietindame/pgcheckpoint/internal/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var checkpointName string

func init() {
	createCmd.Flags().StringVarP(&checkpointName, "name", "n", "", "optional human-readable name for the checkpoint")
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new checkpoint.",
	Long: `Create a new database checkpoint by running pg_dump against the configured
PostgreSQL database. The resulting SQL file is saved to the checkpoints
directory under the active profile.

Use --naming-mode to choose between sequential, timestamp, compact, or
unix naming.
This is the default command when pgcheckpoint is called without a subcommand.`,
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

		mode, err := getNamingMode()
		if err != nil {
			return err
		}

		out, path, err := checkpoint.CreateCheckpoint(url, getCheckpointDir(), profile, mode, checkpointName)

		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}

		if len(out) > 0 {
			ui.Detail(out)
		}
		ui.Success("Created checkpoint:", path)
		return nil
	},
}
