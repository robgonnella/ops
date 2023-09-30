package core

import (
	"context"
	"sync"

	"github.com/robgonnella/go-lanscan/network"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
)

// EventListener represents a registered listener for events
type EventListener struct {
	id      int
	channel chan *event.Event
}

// Core represents our core data structure through which the ui can interact
// with the backend
type Core struct {
	ctx            context.Context
	cancel         context.CancelFunc
	conf           *config.Config
	networkInfo    *network.NetworkInfo
	configService  config.Service
	discovery      discovery.Service
	log            logger.Logger
	eventChan      chan *event.Event
	evtListeners   []*EventListener
	nextListenerId int
	errorChan      chan error
	mux            sync.Mutex
}

// New returns new core module for given configuration
func New(
	networkInfo *network.NetworkInfo,
	conf *config.Config,
	configService config.Service,
	discovery discovery.Service,
	discoveryEvtChan chan *event.Event,
) *Core {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())

	return &Core{
		ctx:            ctx,
		cancel:         cancel,
		networkInfo:    networkInfo,
		conf:           conf,
		configService:  configService,
		discovery:      discovery,
		eventChan:      discoveryEvtChan,
		evtListeners:   []*EventListener{},
		errorChan:      make(chan error),
		nextListenerId: 1,
		mux:            sync.Mutex{},
		log:            log,
	}
}

// Stop stops all processes managed by Core
// The core will be useless after calling stop, a new one must be
// instantiated to continue.
func (c *Core) Stop() error {
	c.discovery.Stop()
	c.cancel()
	return c.ctx.Err()
}

// Conf return the currently loaded configuration
func (c *Core) Conf() config.Config {
	return *c.conf
}

func (c *Core) NetworkInfo() network.NetworkInfo {
	return *c.networkInfo
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

	c.conf = updated

	return nil
}

// SetConfig sets the current active configuration
func (c *Core) SetConfig(id string) error {
	conf, err := c.configService.Get(id)

	if err != nil {
		return err
	}

	c.conf = conf

	return nil
}

// DeleteConfig deletes a configuration
func (c *Core) DeleteConfig(id string) error {
	return c.configService.Delete(id)
}

// GetConfigs returns all stored configs
func (c *Core) GetConfigs() ([]*config.Config, error) {
	return c.configService.GetAll()
}

// StartDaemon starts the network monitoring processes in a goroutine
func (c *Core) StartDaemon(errorReporter chan error) {
	go func() {
		if err := c.Monitor(); err != nil {
			go func() {
				c.errorChan <- err
			}()
			go func() {
				errorReporter <- err
			}()
		}
	}()
}

// RegisterEventListener registers a channel as a listener for events
func (c *Core) RegisterEventListener(channel chan *event.Event) int {
	c.mux.Lock()
	defer c.mux.Unlock()

	listener := &EventListener{
		id:      c.nextListenerId,
		channel: channel,
	}
	c.evtListeners = append(c.evtListeners, listener)
	c.nextListenerId++

	return listener.id
}

// RemoveEventListener removes and closes a channel previously
// registered as a listener
func (c *Core) RemoveEventListener(id int) {
	c.mux.Lock()
	defer c.mux.Unlock()

	listeners := []*EventListener{}
	for _, listener := range c.evtListeners {
		if listener.id != id {
			listeners = append(listeners, listener)
		}
	}

	c.evtListeners = listeners
}
