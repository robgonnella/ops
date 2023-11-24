package core

import (
	"errors"
	"slices"
	"time"

	"github.com/robgonnella/go-lanscan/pkg/network"
	"github.com/robgonnella/go-lanscan/pkg/oui"
	"github.com/robgonnella/go-lanscan/pkg/scanner"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/exception"

	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
)

// getDefaultConfig creates and returns a default configuration
func getDefaultConfig(networkInfo network.Network) *config.Config {
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
		CIDR: networkInfo.Cidr(),
	}
}

// CreateNewAppCore creates and returns a new instance of *core.Core
func CreateNewAppCore(networkInfo network.Network) (*Core, error) {
	configPath := viper.Get("config-path").(string)
	configRepo := config.NewJSONRepo(configPath)
	configService := config.NewConfigService(configRepo)

	conf, err := configService.GetByCIDR(networkInfo.Cidr())

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

	ports := []string{conf.SSH.Port}

	for _, c := range conf.SSH.Overrides {
		if idx := slices.Index(ports, c.Port); idx == -1 {
			ports = append(ports, c.Port)
		}
	}

	vendorRepo, err := oui.GetDefaultVendorRepo()

	if err != nil {
		return nil, err
	}

	netScanner := scanner.NewFullScanner(
		networkInfo,
		[]string{},
		ports,
		54321,
		scanResults,
		scanner.WithVendorInfo(vendorRepo),
	)

	detailScanner := discovery.NewUnameScanner(*conf)

	eventChan := make(chan *event.Event)

	scannerService := discovery.NewScannerService(
		*conf,
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
