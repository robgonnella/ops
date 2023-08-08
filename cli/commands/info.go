package commands

import (
	"fmt"
	"os/exec"

	app_info "github.com/robgonnella/ops/internal/app-info"
	"github.com/spf13/cobra"
)

func info() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print detailed app info",
		Run: func(cmd *cobra.Command, args []string) {
			ansibleCmd := exec.Command("ansible", "--version")
			ansibleInfo, _ := ansibleCmd.Output()

			fmt.Printf(
				"%s: %s\n\n%s\n",
				app_info.NAME,
				app_info.VERSION,
				ansibleInfo,
			)
		},
	}

	return cmd
}
