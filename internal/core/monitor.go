package core

import (
	"fmt"
	"time"

	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/server"
)

// Run runs the sequence driver for the HostInstallStage
func (c *Core) Monitor() error {
	evtReceiveChan := make(chan *event.Event, 100)

	// create event subscription
	subscription := c.serverService.StreamEvents(evtReceiveChan)

	defer c.serverService.StopStream(subscription)

	// Start network scanner
	go c.discovery.MonitorNetwork()

	// start polling for database updates
	go c.pollForDatabaseUpdates()

	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case evt := <-evtReceiveChan:
			c.handleServerEvent(evt)
		}
	}
}

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

	c.logger.Info().Fields(fields).Msg("Event Received")

	for _, listener := range c.evtListeners {
		listener.channel <- evt
	}
}

func (c *Core) pollForDatabaseUpdates() error {
	pollTime := time.Second * 2
	errCount := 0

	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
			if errCount >= 5 {
				return fmt.Errorf("too many consecutive errors encountered")
			}

			response, err := c.serverService.GetAllServersInNetworkTargets(
				c.conf.Targets,
			)

			if err != nil {
				c.logger.Error().Err(err)
				errCount++
				continue
			}

			errCount = 0

			for _, listener := range c.serverPollListeners {
				listener.channel <- response
			}

			time.Sleep(pollTime)
		}
	}
}
