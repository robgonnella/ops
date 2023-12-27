package event

import (
	"slices"
	"sync"
)

// nolint:revive
// EventListener represents a single listener for an event channel
type EventListener struct {
	ID        int
	eventType EventType
	channel   chan Event
}

// nolint:revive
// EventManager implements the event.Manager interface
type EventManager struct {
	listeners []*EventListener
	mux       sync.RWMutex
	nextID    int
}

// NewEventManager returns a new instance of EventManager
func NewEventManager() *EventManager {
	return &EventManager{
		listeners: []*EventListener{},
		mux:       sync.RWMutex{},
		nextID:    1,
	}
}

// RegisterListener registers a listener with the event manager
func (m *EventManager) RegisterListener(eventType EventType, listener chan Event) int {
	m.mux.Lock()
	defer m.mux.Unlock()

	id := m.nextID
	eventListener := &EventListener{
		ID:        id,
		eventType: EventType(eventType),
		channel:   listener,
	}
	m.listeners = append(m.listeners, eventListener)
	m.nextID++

	return id
}

// RemoveListener removes a listener from the event manager
func (m *EventManager) RemoveListener(id int) int {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.listeners = slices.DeleteFunc(m.listeners, func(l *EventListener) bool {
		return l.ID == id
	})

	return id
}

// Send sends an event to all listeners for that event
func (m *EventManager) Send(evt Event) {
	for _, l := range m.listeners {
		if l.eventType == EventType(evt.Type) {
			go func(listener *EventListener, event Event) {
				listener.channel <- event
			}(l, evt)
		}
	}
}

// ReportFatalError reports a fatal error to all listeners for that event
func (m *EventManager) ReportFatalError(err error) {
	m.Send(Event{
		Type:    FatalErrorEventType,
		Payload: err,
	})
}

// ReportError reports an error to all listeners for that event
func (m *EventManager) ReportError(err error) {
	m.Send(Event{
		Type:    ErrorEventType,
		Payload: err,
	})
}
