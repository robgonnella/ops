package ui

import (
	"fmt"
	"os"

	"github.com/robgonnella/ops/internal/core"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/util"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// capture original stdout & stderr to be restored when ssh-ing to a server
var originalStdout = os.Stdout
var originalStderr = os.Stderr

// restores original stdout & stderr
func restoreStdout() {
	os.Stdout = originalStdout
	os.Stderr = originalStderr
}

// maskStdout
func maskStdout() {
	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)
}

// UI wrapper around view used for initial launch
type UI struct {
	view *view
}

// NewUI returns a new instance of UI
func NewUI() *UI {
	return &UI{}
}

// Launch configures logging, creates a new instance of view and launches
// our terminal UI application
func (u *UI) Launch(debug bool) error {
	log := logger.New()

	networkInfo, err := util.GetNetworkInfo()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to get default network info")
	}

	appCore, err := core.CreateNewAppCore(networkInfo)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to create app core")
	}

	if debug {
		return appCore.Monitor()
	}

	allConfigs, err := appCore.GetConfigs()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to retrieve configs")
	}

	u.view = newView(allConfigs, appCore)

	level := zerolog.GlobalLevel()

	if level != zerolog.Disabled {
		logFile, ok := viper.Get("log-file").(string)

		if !ok || logFile == "" {
			log.Error().Err(
				fmt.Errorf("invalid log file path: %s", logFile),
			)
			log.Info().Msg("disabling logs")
			logger.SetGlobalLevel(zerolog.Disabled)
		} else {
			file, err := os.OpenFile(
				logFile,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY,
				0644,
			)

			if err != nil {
				log.Error().Err(err).Msg("")
				log.Info().Msg("disabling logs")
				logger.SetGlobalLevel(zerolog.Disabled)
			} else {
				logger.SetGlobalLogFile(file)
				logger.SetWithCaller()
			}
		}
	}

	maskStdout()

	defer restoreStdout()

	return u.view.run()
}
