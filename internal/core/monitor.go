package core

import (
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
)

// Monitor starts the processes for monitoring and tracking
// devices on the configured network
func (c *Core) Monitor() error {
	done := make(chan bool)

	// Start network scanner
	go func() {
		c.discovery.MonitorNetwork()
		// bail if network monitoring stops
		done <- true
	}()

	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-done:
			return nil
		case err := <-c.errorChan:
			c.log.Error().Err(err).Msg("core error")
			return err
		case evt, ok := <-c.eventChan:
			if !ok {
				return nil
			}
			c.handleDiscoveryEvent(evt)
		}
	}
}

// handles events from discovery service
func (c *Core) handleDiscoveryEvent(evt *event.Event) {
	payload, ok := evt.Payload.(*discovery.DiscoveryResult)

	if !ok {
		return
	}

	fields := map[string]interface{}{
		"type":     evt.Type,
		"id":       payload.ID,
		"hostname": payload.Hostname,
		"os":       payload.OS,
		"ip":       payload.IP,
		"ssh":      payload.Port.Status,
	}

	c.log.Info().Fields(fields).Msg("Event Received")

	c.mux.Lock()
	defer c.mux.Unlock()

	for _, listener := range c.evtListeners {
		listener.channel <- evt
	}
}
