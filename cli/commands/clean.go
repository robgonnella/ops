package commands

import (
	"os"

	"github.com/robgonnella/opi/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// creates and returns the "monitor" command
func clean(props *CommandProps) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clears the database file",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.New()

			dbFile, ok := viper.Get("database-file").(string)

			if ok && dbFile != "" {
				if err := os.Remove(dbFile); err != nil {
					return err
				}
				log.Info().Msg("removed database file")
			}

			logFile, ok := viper.Get("log-file").(string)

			if ok && logFile != "" {
				if err := os.RemoveAll(logFile); err != nil {
					return err
				}
				log.Info().Msg("removed log file")
			}

			if all {
				configFile := viper.Get("config-file").(string)

				if err := os.RemoveAll(configFile); err != nil {
					return err
				}

				log.Info().Msg("removed config file")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Clean all including app config")

	return cmd
}