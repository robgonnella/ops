package commands

import (
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/ui"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// CommandProps injected props that can be made available to all commands
type CommandProps struct {
	UI *ui.UI
}

// Root builds and returns our root command
func Root(props *CommandProps) *cobra.Command {
	var debug bool
	var verbose bool
	var silent bool

	cmd := &cobra.Command{
		Use:     "ops",
		Short:   "The ops command line utility",
		Aliases: []string{"clean"},
		// This runs before all commands and all sub-commands
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set logging verbosity for all loggers
			level := zerolog.InfoLevel

			if verbose || debug {
				level = zerolog.DebugLevel
			}

			if silent {
				level = zerolog.Disabled
			}

			logger.SetGlobalLevel(level)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return props.UI.Launch(debug)
		},
	}

	// Persistent flags available to all commands
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "run in terminal log mode - no ui")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show debug logs")
	cmd.PersistentFlags().BoolVar(&silent, "silent", false, "disables all logging")

	cmd.AddCommand(clear())
	cmd.AddCommand(version())

	return cmd
}
