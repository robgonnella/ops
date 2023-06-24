package main

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/robgonnella/ops/cli/commands"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/name"
	"github.com/robgonnella/ops/internal/ui"
	"github.com/spf13/viper"
)

/**
 * Main entry point for all commands
 * Here we setup environment config via viper
 */

func setRuntTimeConfig() error {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	configDir := path.Join(userHomeDir, ".config", name.APP_NAME)

	if err := os.Mkdir(configDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	logFile := path.Join(configDir, name.APP_NAME+".log")

	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return err
	}

	cacheDir := path.Join(userCacheDir, name.APP_NAME)

	if err := os.Mkdir(cacheDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	dbFile := path.Join(cacheDir, name.APP_NAME+".db")

	defaultSSHIdentity := path.Join(userHomeDir, ".ssh", "id_rsa")

	user := os.Getenv("USER")

	// share run-time config globally using viper
	viper.Set("log-file", logFile)
	viper.Set("config-dir", configDir)
	viper.Set("cache-dir", cacheDir)
	viper.Set("database-file", dbFile)
	viper.Set("default-ssh-identity", defaultSSHIdentity)
	viper.Set("user", user)

	return nil
}

// Entry point for the cli
func main() {
	log := logger.New()

	err := setRuntTimeConfig()

	if err != nil {
		log.Fatal().Err(err)
	}

	appUI := ui.NewUI()

	// Get the "root" cobra cli command
	cmd := commands.Root(&commands.CommandProps{
		UI: appUI,
	})

	// execute the cobra command and exit with error code if necessary
	err = cmd.ExecuteContext(context.Background())

	if err != nil {
		log.Fatal().Err(err)
	}
}
