package event_test

import (
	"errors"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/robgonnella/ops/internal/event"
)

func TestEventManager(t *testing.T) {
	t.Run("registers event listener and sends event", func(st *testing.T) {
		eventManager := event.NewEventManager()

		listener := make(chan event.Event)

		eventManager.RegisterListener("test-event", listener)

		eventManager.Send(event.Event{
			Type:    "a-different-type",
			Payload: struct{}{},
		})

		eventManager.Send(event.Event{
			Type:    "test-event",
			Payload: true,
		})

		result := <-listener

		assert.Equal(st, result.Type, event.EventType("test-event"))
	})

	t.Run("removes event listener", func(st *testing.T) {
		eventManager := event.NewEventManager()

		listener := make(chan event.Event)

		id := eventManager.RegisterListener("test-event", listener)

		removedId := eventManager.RemoveListener(id)

		assert.Equal(st, removedId, id)
	})

	t.Run("reports fatal error event", func(st *testing.T) {
		eventManager := event.NewEventManager()

		listener := make(chan event.Event)

		eventManager.RegisterListener(event.FatalErrorEventType, listener)

		eventManager.Send(event.Event{
			Type:    "a-different-type",
			Payload: struct{}{},
		})

		eventManager.ReportFatalError(errors.New("fatal test error"))

		result := <-listener

		assert.Equal(st, result.Type, event.FatalErrorEventType)
	})

	t.Run("reports error event", func(st *testing.T) {
		eventManager := event.NewEventManager()

		listener := make(chan event.Event)

		eventManager.RegisterListener(event.ErrorEventType, listener)

		eventManager.Send(event.Event{
			Type:    "a-different-type",
			Payload: struct{}{},
		})

		eventManager.ReportError(errors.New("test error"))

		result := <-listener

		assert.Equal(st, result.Type, event.ErrorEventType)
	})
}
