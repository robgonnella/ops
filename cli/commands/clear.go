package commands

import (
	"os"

	"github.com/robgonnella/ops/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/**
 * Command to remove config and log files
 */
func clear() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clears config and log files",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.New()

			configFile, ok := viper.Get("config-path").(string)

			if ok && configFile != "" {
				if err := os.RemoveAll(configFile); err != nil {
					return err
				}
				log.Info().Msg("removed config file")
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
