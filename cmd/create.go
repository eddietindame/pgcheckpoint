package cmd

import (
	"fmt"
	"pgcheckpoint/internal/checkpoint"
	"pgcheckpoint/internal/db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		out, path, err := checkpoint.CreateCheckpoint(filename, url)

		if err != nil {
			return fmt.Errorf("%w: %s", err, out)
		}

		fmt.Println()
		fmt.Println(out)
		fmt.Println("Created checkpoint:", path)
		return nil
	},
}
