package commands

import (
	"fmt"

	app_info "github.com/robgonnella/ops/internal/app-info"
	"github.com/spf13/cobra"
)

func version() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(
				"%s: %s\n",
				app_info.NAME,
				app_info.VERSION,
			)
		},
	}

	return cmd
}
