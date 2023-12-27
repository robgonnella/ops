package core

import (
	"errors"

	"github.com/robgonnella/go-lanscan/pkg/network"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
)

// ScannerFactory is a function that returns a new instance of a Scanner
type ScannerFactory func(netInfo network.Network, conf config.Config) (discovery.Scanner, error)

// Core represents our core data structure through which the ui can interact
// with the backend
type Core struct {
	conf                *config.Config
	networkInfo         network.Network
	configService       config.Service
	discovery           discovery.Service
	eventManager        event.Manager
	scannerFactory      ScannerFactory
	debug               bool
	registeredListeners []int
	log                 logger.Logger
}

// New returns new core module for given configuration
func New(
	networkInfo network.Network,
	conf *config.Config,
	configService config.Service,
	discovery discovery.Service,
	eventManager event.Manager,
	scannerFactory ScannerFactory,
	debug bool,
) *Core {
	log := logger.New()

	return &Core{
		networkInfo:         networkInfo,
		conf:                conf,
		configService:       configService,
		discovery:           discovery,
		eventManager:        eventManager,
		scannerFactory:      scannerFactory,
		debug:               debug,
		registeredListeners: []int{},
		log:                 log,
	}
}

// Stop stops all processes managed by Core
// The core will be useless after calling stop, a new one must be
// instantiated to continue.
func (c *Core) Stop() error {
	c.discovery.Stop()
	return nil
}

// Conf return the currently loaded configuration
func (c *Core) Conf() config.Config {
	return *c.conf
}

// NetworkInfo returns the core's network interface
func (c *Core) NetworkInfo() network.Network {
	return c.networkInfo
}

// CreateConfig creates a new config
func (c *Core) CreateConfig(conf config.Config) error {
	_, err := c.configService.Create(&conf)

	if err != nil {
		return err
	}

	return nil
}

// UpdateConfig updates an existing config
func (c *Core) UpdateConfig(conf config.Config) error {
	updated, err := c.configService.Update(&conf)

	if err != nil {
		return err
	}

	if updated.ID == c.conf.ID {
		netInfo, err := network.NewNetworkFromInterfaceName(updated.Interface)

		if err != nil {
			return err
		}

		c.conf = updated
		c.networkInfo = netInfo

		newScanner, err := c.scannerFactory(c.networkInfo, c.Conf())

		if err != nil {
			return err
		}

		c.discovery.SetConfigAndScanner(c.Conf(), newScanner)
	}

	return nil
}

// SetConfig sets the current active configuration
func (c *Core) SetConfig(id string) error {
	if id == c.conf.ID {
		return nil
	}

	conf, err := c.configService.Get(id)

	if err != nil {
		return err
	}

	c.conf = conf

	netInfo, err := network.NewNetworkFromInterfaceName(c.conf.Interface)

	if err != nil {
		return err
	}

	c.networkInfo = netInfo

	newScanner, err := c.scannerFactory(c.networkInfo, c.Conf())

	if err != nil {
		return err
	}

	c.discovery.SetConfigAndScanner(c.Conf(), newScanner)

	return nil
}

// DeleteConfig deletes a configuration
func (c *Core) DeleteConfig(id string) error {
	if id == c.conf.ID {
		return errors.New("cannot delete current active config")
	}

	err := c.configService.Delete(id)

	return err
}

// GetConfigs returns all stored configs
func (c *Core) GetConfigs() ([]*config.Config, error) {
	return c.configService.GetAll()
}

// Monitor starts the processes for monitoring and tracking
// devices on the configured network
func (c *Core) Monitor() error {
	evtChan := make(chan event.Event)
	if c.debug {
		c.registeredListeners = append(c.registeredListeners,
			c.eventManager.RegisterListener(discovery.ArpUpdateEvent, evtChan),
			c.eventManager.RegisterListener(discovery.SynUpdateEvent, evtChan),
			c.eventManager.RegisterListener(event.FatalErrorEventType, evtChan),
		)

		defer func() {
			for _, id := range c.registeredListeners {
				c.eventManager.RemoveListener(id)
			}
		}()

		go func() {
			for evt := range evtChan {
				if err, ok := evt.Payload.(error); ok {
					c.log.Fatal().Err(err).Msg("")
				}

				if result, ok := evt.Payload.(discovery.DiscoveryResult); ok {
					fields := map[string]interface{}{
						"id":         result.ID,
						"hostname":   result.Hostname,
						"ip":         result.IP,
						"os":         result.OS,
						"vendor":     result.Vendor,
						"status":     result.Status,
						"port":       result.Port.ID,
						"portStatus": result.Port.Status,
					}

					c.log.Info().Fields(fields).Msg(result.Type)
				}
			}
		}()
	}

	return c.discovery.MonitorNetwork()
}

// StartDaemon starts the network monitoring processes in a goroutine
func (c *Core) StartDaemon() {
	go func() {
		if err := c.Monitor(); err != nil {
			c.eventManager.ReportFatalError(err)
		}
	}()
}
