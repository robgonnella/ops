package core

import (
	"context"
	"sync"

	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
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
	conf                *config.Config
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

func (c *Core) Stop() error {
	c.discovery.Stop()
	if c.eventSubscription != 0 {
		c.serverService.StopStream(c.eventSubscription)
	}
	c.cancel()
	return c.ctx.Err()
}

func (c *Core) Conf() config.Config {
	return *c.conf
}

func (c *Core) CreateConfig(conf config.Config) error {
	_, err := c.configService.Create(&conf)

	if err != nil {
		return err
	}

	return nil
}

func (c *Core) UpdateConfig(conf config.Config) error {
	updated, err := c.configService.Update(&conf)

	if err != nil {
		return err
	}

	if err := c.configService.SetLastLoaded(updated.ID); err != nil {
		return err
	}

	c.conf = updated

	return nil
}

func (c *Core) SetConfig(id int) error {
	conf, err := c.configService.Get(id)

	if err != nil {
		return err
	}

	if err := c.configService.SetLastLoaded(conf.ID); err != nil {
		return err
	}

	c.conf = conf

	return nil
}

func (c *Core) DeleteConfig(id int) error {
	return c.configService.Delete(id)
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
		if listener.id == id {
			close(listener.channel)
		}
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
		if listener.id == id {
			close(listener.channel)
		}
		if listener.id != id {
			listeners = append(listeners, listener)
		}
	}

	c.serverPollListeners = listeners
}
