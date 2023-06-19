package core

import (
	"context"
	"sync"

	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/discovery"
	"github.com/robgonnella/opi/internal/event"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/server"
)

type EventListener struct {
	id      int
	channel chan *event.Event
}

type ServerPollListener struct {
	id      int
	channel chan []*server.Server
}

// Core represents our core data structure
type Core struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	conf                config.Config
	configService       config.Service
	discovery           discovery.Service
	serverService       server.Service
	logger              logger.Logger
	evtListeners        []*EventListener
	serverPollListeners []*ServerPollListener
	nextListenerId      int
	mux                 sync.Mutex
}

// New returns new core module for given configuration
func New(
	conf config.Config,
	configService config.Service,
	serverService server.Service,
	discovery discovery.Service,
) *Core {
	logger := logger.New()

	ctx, cancel := context.WithCancel(context.Background())

	return &Core{
		ctx:                 ctx,
		cancel:              cancel,
		conf:                conf,
		configService:       configService,
		discovery:           discovery,
		serverService:       serverService,
		evtListeners:        []*EventListener{},
		serverPollListeners: []*ServerPollListener{},
		nextListenerId:      1,
		mux:                 sync.Mutex{},
		logger:              logger,
	}
}

func (c *Core) Stop() error {
	c.discovery.Stop()
	c.cancel()
	return c.ctx.Err()
}

func (c *Core) Conf() config.Config {
	return c.conf
}

func (c *Core) UpdateConfig(conf config.Config) error {
	updated, err := c.configService.Update(&conf)

	if err != nil {
		return err
	}

	c.conf = *updated

	return nil
}

func (c *Core) SetConfig(name string) error {
	conf, err := c.configService.Get(name)

	if err != nil {
		return err
	}

	c.conf = *conf

	return nil
}

func (c *Core) DeleteConfig(name string) error {
	return c.configService.Delete(name)
}

func (c *Core) GetConfigs() ([]*config.Config, error) {
	return c.configService.GetAll()
}

func (c *Core) StartDaemon() {
	go c.Monitor()
}

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

func (c *Core) RemoveServerPollListener(id int) {
	c.mux.Lock()
	defer c.mux.Unlock()

	listeners := []*ServerPollListener{}
	for _, listener := range c.serverPollListeners {
		if listener.id != id {
			listeners = append(listeners, listener)
		}
	}

	c.serverPollListeners = listeners
}
