package cmd

import (
	"fmt"
	"github.com/eddietindame/pgcheckpoint/internal/checkpoint"
	"github.com/eddietindame/pgcheckpoint/internal/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new checkpoint.",
	Long: `Create a new database checkpoint by running pg_dump against the configured
PostgreSQL database. The resulting SQL file is saved to the checkpoints
directory under the active profile.

The checkpoint filename can be controlled with the --filename flag.
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
		fmt.Println("Database url:", url)

		out, path, err := checkpoint.CreateCheckpoint(filename, url, getCheckpointDir(), profile)

		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}

		if len(out) > 0 {
			fmt.Printf("\n%s\n", out)
		}
		fmt.Println("Created checkpoint:", path)
		return nil
	},
}
