package commands

import (
	"github.com/robgonnella/opi/internal/ui"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type CommandProps struct {
	UI *ui.UI
}

// flag var to set verbose logging

// Root builds and returns our root command
func Root(props *CommandProps) *cobra.Command {
	var verbose bool
	var silent bool

	cmd := &cobra.Command{
		Use:     "opi",
		Short:   "The opi command line utility",
		Aliases: []string{"clean"},
		// This runs before all commands and all sub-commands
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set logging verbosity for all loggers
			zerolog.SetGlobalLevel(zerolog.InfoLevel)

			if verbose {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			if silent {
				zerolog.SetGlobalLevel(zerolog.Disabled)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return props.UI.Launch()
		},
	}

	// Persistent flags available to all commands
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show debug logs")
	cmd.PersistentFlags().BoolVar(&silent, "silent", false, "disables all logging")

	cmd.AddCommand(clear())

	return cmd
}
