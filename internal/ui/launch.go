package ui

import (
	"fmt"
	"os"

	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/util"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var originalStdout = os.Stdout
var originalStderr = os.Stderr

func restoreStdout() {
	os.Stdout = originalStdout
	os.Stderr = originalStderr
}

type UI struct {
	view *view
}

func NewUI() *UI {
	return &UI{}
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

	userIP, cidr, err := util.GetNetworkInfo()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to get default network info")
	}

	appCore, err := util.CreateNewAppCore(*cidr)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to create app core")
	}

	u.view = newView(*userIP, appCore)

	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)

	return u.view.run()
}
