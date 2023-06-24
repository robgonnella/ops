package core

import (
	"fmt"
	"time"

	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/server"
)

// Monitor starts the processes for monitoring and tracking
// devices on the configured network
func (c *Core) Monitor() error {
	evtReceiveChan := make(chan *event.Event)

	// create event subscription
	c.eventSubscription = c.serverService.StreamEvents(evtReceiveChan)

	// Start network scanner
	go c.discovery.MonitorNetwork()

	// start polling for database updates
	go c.pollForDatabaseUpdates()

	for {
		evt, ok := <-evtReceiveChan
		if !ok {
			c.mux.Lock()
			for _, listener := range c.evtListeners {
				close(listener.channel)
			}
			c.mux.Unlock()
			return nil
		}
		c.handleServerEvent(evt)
	}
}

// handles server events from the database
func (c *Core) handleServerEvent(evt *event.Event) {
	payload := evt.Payload.(*server.Server)

	fields := map[string]interface{}{
		"type":     evt.Type,
		"id":       payload.ID,
		"hostname": payload.Hostname,
		"os":       payload.OS,
		"ip":       payload.IP,
		"ssh":      payload.SshStatus,
	}

	c.log.Info().Fields(fields).Msg("Event Received")

	c.mux.Lock()
	defer c.mux.Unlock()

	for _, listener := range c.evtListeners {
		listener.channel <- evt
	}
}

// polls database for all servers within configured network targets
func (c *Core) pollForDatabaseUpdates() error {
	pollTime := time.Second * 2
	errCount := 0

	for {
		select {
		case <-c.ctx.Done():
			for _, listener := range c.serverPollListeners {
				close(listener.channel)
			}
			return c.ctx.Err()
		default:
			if errCount >= 5 {
				return fmt.Errorf("too many consecutive errors encountered")
			}

			response, err := c.serverService.GetAllServersInNetworkTargets(
				c.conf.Targets,
			)

			if err != nil {
				c.log.Error().Err(err)
				errCount++
				continue
			}

			errCount = 0

			c.mux.Lock()
			for _, listener := range c.serverPollListeners {
				listener.channel <- response
			}
			c.mux.Unlock()

			time.Sleep(pollTime)
		}
	}
}
