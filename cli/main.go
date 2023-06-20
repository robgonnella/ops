package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/robgonnella/opi/cli/commands"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/core"
	"github.com/robgonnella/opi/internal/discovery"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/name"
	"github.com/robgonnella/opi/internal/server"
	"github.com/robgonnella/opi/internal/ui"
	"github.com/spf13/viper"
)

// get network interface associated with ip
func getIPNetByIP(ip net.IP) (*net.IPNet, error) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()

		if err != nil {
			continue
		}

		for _, addr := range addrs {
			_, ipnet, err := net.ParseCIDR(addr.String())

			if err != nil {
				continue
			}

			if ipnet.Contains(ip) {
				return ipnet, nil
			}
		}
	}

	return nil, errors.New("failed to find IPNet")
}

// get cidr for preferred outbound ip of this machine
func getDefaultCidr() (*string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ipnet, err := getIPNetByIP(localAddr.IP)

	if err != nil {
		return nil, err
	}

	size, _ := ipnet.Mask.Size()

	cidr := fmt.Sprintf("%s/%d", localAddr.IP, size)

	return &cidr, nil
}

func setConfigPaths() (string, error) {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	configDir := path.Join(userHomeDir, ".config", name.APP_NAME)

	if err := os.Mkdir(configDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}

	configFile := path.Join(configDir, name.APP_NAME+".yml")

	logFile := path.Join(configDir, name.APP_NAME+".log")

	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", err
	}

	cacheDir := path.Join(userCacheDir, name.APP_NAME)

	if err := os.Mkdir(cacheDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}

	dbFile := path.Join(cacheDir, name.APP_NAME+".db")

	// share location of files and directories globally using viper
	viper.Set("log-file", logFile)
	viper.Set("config-dir", configDir)
	viper.Set("config-file", configFile)
	viper.Set("cache-dir", cacheDir)
	viper.Set("database-file", dbFile)

	return configFile, nil
}

// Entry point for the cli.
func main() {
	log := logger.New()

	configFile, err := setConfigPaths()

	if err != nil {
		log.Fatal().Err(err)
	}

	conf, err := config.New(configFile)

	if err != nil {
		conf, err = config.Default()

		if err != nil {
			log.Fatal().Err(err)
		}
	}

	serverRepo, err := server.NewSqliteDatabase()

	if err != nil {
		log.Fatal().Err(err)
	}

	serverService := server.NewService(*conf, serverRepo)

	discoveryService, err := discovery.NewNmapService(
		*conf,
		serverService,
	)

	if err != nil {
		log.Fatal().Err(err)
	}

	appCore := core.New(*conf, serverService, discoveryService)

	defaultCidr := ""

	if len(conf.Discovery.Targets) == 0 {
		cidr, err := getDefaultCidr()

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to find default network cidr")
		}

		defaultCidr = *cidr
	}

	appUI := ui.New(appCore, defaultCidr)

	// Get the "root" cobra cli command
	cmd := commands.Root(&commands.CommandProps{
		Core: appCore,
		UI:   appUI,
	})

	// Allows "grepping" of command output
	cmd.SetOutput(os.Stdout)

	// execute the cobra command and exit with error code if necessary
	err = cmd.ExecuteContext(context.Background())

	if err != nil {
		log.Fatal().Err(err)
	}
}
