package event

import (
	"slices"
	"sync"
)

type EventListener struct {
	ID        int
	eventType EventType
	channel   chan Event
}

type EventManager struct {
	listeners []*EventListener
	mux       sync.RWMutex
	nextID    int
}

func NewEventManager() *EventManager {
	return &EventManager{
		listeners: []*EventListener{},
		mux:       sync.RWMutex{},
		nextID:    1,
	}
}

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

func (m *EventManager) RemoveListener(id int) int {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.listeners = slices.DeleteFunc(m.listeners, func(l *EventListener) bool {
		return l.ID == id
	})

	return id
}

func (m *EventManager) Send(evt Event) {
	for _, l := range m.listeners {
		if l.eventType == EventType(evt.Type) {
			go func(listener *EventListener, event Event) {
				listener.channel <- event
			}(l, evt)
		}
	}
}

func (m *EventManager) SendFatalError(err error) {
	m.Send(Event{
		Type:    FatalErrorEventType,
		Payload: err,
	})
}
