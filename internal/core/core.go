package core

import (
	"context"
	"sync"

	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/util"
)

// EventListener represents a registered listener for database events
type EventListener struct {
	id      int
	channel chan *event.Event
}

// ServerPollListener represents a registered listener for
// database server polling
type ServerPollListener struct {
	id      int
	channel chan []*server.Server
}

// Core represents our core data structure through which the ui can interact
// with the backend
type Core struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	conf                *config.Config
	networkInfo         *util.NetworkInfo
	configService       config.Service
	discovery           discovery.Service
	serverService       server.Service
	log                 logger.Logger
	eventSubscription   int
	evtListeners        []*EventListener
	serverPollListeners []*ServerPollListener
	nextListenerId      int
	mux                 sync.Mutex
}

// New returns new core module for given configuration
func New(
	networkInfo *util.NetworkInfo,
	conf *config.Config,
	configService config.Service,
	serverService server.Service,
	discovery discovery.Service,
) *Core {
	log := logger.New()

	ctx, cancel := context.WithCancel(context.Background())

	return &Core{
		ctx:                 ctx,
		cancel:              cancel,
		networkInfo:         networkInfo,
		conf:                conf,
		configService:       configService,
		discovery:           discovery,
		serverService:       serverService,
		evtListeners:        []*EventListener{},
		serverPollListeners: []*ServerPollListener{},
		nextListenerId:      1,
		mux:                 sync.Mutex{},
		log:                 log,
	}
}

// Stop stops all processes managed by Core
// The core will be useless after calling stop, a new one must be
// instantiated to continue.
func (c *Core) Stop() error {
	c.discovery.Stop()
	if c.eventSubscription != 0 {
		c.serverService.StopStream(c.eventSubscription)
	}
	c.cancel()
	return c.ctx.Err()
}

// Conf return the currently loaded configuration
func (c *Core) Conf() config.Config {
	return *c.conf
}

func (c *Core) NetworkInfo() util.NetworkInfo {
	return *c.networkInfo
}

// CreateConfig creates a new config in the database
func (c *Core) CreateConfig(conf config.Config) error {
	_, err := c.configService.Create(&conf)

	if err != nil {
		return err
	}

	return nil
}

// UpdateConfig updates an existing config in the database
func (c *Core) UpdateConfig(conf config.Config) error {
	updated, err := c.configService.Update(&conf)

	if err != nil {
		return err
	}

	c.conf = updated

	return nil
}

// SetConfig sets the current active configuration
func (c *Core) SetConfig(id int) error {
	conf, err := c.configService.Get(id)

	if err != nil {
		return err
	}

	c.conf = conf

	return nil
}

// DeleteConfig deletes a configuration
func (c *Core) DeleteConfig(id int) error {
	return c.configService.Delete(id)
}

// GetConfigs returns all stored configs
func (c *Core) GetConfigs() ([]*config.Config, error) {
	return c.configService.GetAll()
}

// StartDaemon starts the network monitoring processes in a goroutine
func (c *Core) StartDaemon() {
	go c.Monitor()
}

// RegisterEventListener registers a channel as a listener for database events
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
		if listener.id == id {
			close(listener.channel)
		}
		if listener.id != id {
			listeners = append(listeners, listener)
		}
	}

	c.evtListeners = listeners
}

// RegisterServerPollListener registers a channel as a
// listener for server polling results
func (c *Core) RegisterServerPollListener(channel chan []*server.Server) int {
	c.mux.Lock()
	defer c.mux.Unlock()

	listener := &ServerPollListener{
		id:      c.nextListenerId,
		channel: channel,
	}

	c.serverPollListeners = append(c.serverPollListeners, listener)
	c.nextListenerId++

	return listener.id
}

// RemoveServerPollListener removes and closes a channel previously
// registered to listen for server polling results
func (c *Core) RemoveServerPollListener(id int) {
	c.mux.Lock()
	defer c.mux.Unlock()

	listeners := []*ServerPollListener{}
	for _, listener := range c.serverPollListeners {
		if listener.id == id {
			close(listener.channel)
		}
		if listener.id != id {
			listeners = append(listeners, listener)
		}
	}

	c.serverPollListeners = listeners
}
