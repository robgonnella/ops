package commands

import (
	"os"

	"github.com/robgonnella/ops/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/**
 * Command to remove database and log files
 */
func clear() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clears database and log files",
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

			return nil
		},
	}

	return cmd
}
