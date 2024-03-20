package core

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
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

// CreateNewAppCore creates and returns a new instance of *core.Core
func CreateNewAppCore(networkInfo network.Network, eventManager event.Manager, debug bool) (*Core, error) {
	configPath := viper.Get("config-path").(string)
	user := viper.Get("user").(string)
	identity := viper.Get("default-ssh-identity").(string)
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	defaultConf := config.Config{
		ID:   uuid.New().String(),
		Name: nameGenerator.Generate(),
		SSH: config.SSHConfig{
			User:      user,
			Identity:  identity,
			Port:      "22",
			Overrides: []config.SSHOverride{},
		},
		Interface: networkInfo.Interface().Name,
	}

	configRepo, err := config.NewJSONRepo(configPath, defaultConf)

	if err != nil {
		return nil, err
	}

	configService := config.NewConfigService(configRepo)

	conf, err := configService.GetByInterface(networkInfo.Interface().Name)

	if err != nil {
		if errors.Is(err, exception.ErrRecordNotFound) {
			conf, err = configService.Create(&defaultConf)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	netScanner, err := createScanner(networkInfo, *conf)

	if err != nil {
		return nil, err
	}

	detailScanner := discovery.NewUnameScanner(*conf)

	scannerService := discovery.NewScannerService(
		*conf,
		netScanner,
		detailScanner,
		eventManager,
	)

	return New(
		networkInfo,
		conf,
		configService,
		scannerService,
		eventManager,
		createScanner,
		debug,
	), nil
}

func createScanner(netInfo network.Network, conf config.Config) (discovery.Scanner, error) {
	vendorRepo, err := oui.GetDefaultVendorRepo()

	if err != nil {
		return nil, err
	}

	ports := []string{conf.SSH.Port}

	for _, c := range conf.SSH.Overrides {
		if idx := slices.Index(ports, c.Port); idx == -1 {
			ports = append(ports, c.Port)
		}
	}

	return scanner.NewFullScanner(
		netInfo,
		[]string{},
		ports,
		54321,
		scanner.WithVendorInfo(vendorRepo),
	), nil

}
