package ui

import (
	"fmt"
	"os"

	"github.com/robgonnella/opi/internal/core"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type UI struct {
	uiApp *app
}

var originalStdout = os.Stdout
var originalStderr = os.Stderr

func New(appCore *core.Core) *UI {
	return &UI{
		uiApp: newApp(appCore),
	}
}

func (u *UI) Launch() error {
	log := logger.New()

	level := zerolog.GlobalLevel()

	if level != zerolog.Disabled {
		logFile, ok := viper.Get("log-file").(string)

		if !ok || logFile == "" {
			log.Error().Err(
				fmt.Errorf("invalid log file path: %s", logFile),
			)
			log.Info().Msg("disabling logs")
			zerolog.SetGlobalLevel(zerolog.Disabled)
		} else {
			if err := logger.GlobalSetLogFile(logFile); err != nil {
				log.Error().Err(err)
				log.Info().Msg("disabling logs")
				zerolog.SetGlobalLevel(zerolog.Disabled)
			}
		}

	}

	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)

	return u.uiApp.run()
}
