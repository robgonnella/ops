package core

import (
	"errors"
	"time"

	"github.com/robgonnella/go-lanscan/network"
	"github.com/robgonnella/go-lanscan/scanner"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/exception"

	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
)

// getDefaultConfig creates and returns a default configuration
func getDefaultConfig(networkInfo *network.NetworkInfo) *config.Config {
	user := viper.Get("user").(string)
	identity := viper.Get("default-ssh-identity").(string)
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	return &config.Config{
		Name: nameGenerator.Generate(),
		SSH: config.SSHConfig{
			User:      user,
			Identity:  identity,
			Port:      "22",
			Overrides: []config.SSHOverride{},
		},
		CIDR: networkInfo.Cidr,
	}
}

// CreateNewAppCore creates and returns a new instance of *core.Core
func CreateNewAppCore(networkInfo *network.NetworkInfo) (*Core, error) {
	configPath := viper.Get("config-path").(string)
	configRepo := config.NewJSONRepo(configPath)
	configService := config.NewConfigService(configRepo)

	conf, err := configService.GetByCIDR(networkInfo.Cidr)

	if err != nil {
		if errors.Is(err, exception.ErrRecordNotFound) {
			conf = getDefaultConfig(networkInfo)
			conf, err = configService.Create(conf)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	scanResults := make(chan *scanner.ScanResult)

	netScanner := scanner.NewFullScanner(
		networkInfo,
		[]string{},
		[]string{"22"},
		54321,
		scanResults,
	)

	detailScanner := discovery.NewUnameScanner(*conf)

	eventChan := make(chan *event.Event)

	scannerService := discovery.NewScannerService(
		netScanner,
		detailScanner,
		scanResults,
		eventChan,
	)

	return New(
		networkInfo,
		conf,
		configService,
		scannerService,
		eventChan,
	), nil
}
