package event

// nolint:revive
// EventType represents different types of events
type EventType string

const (
	// FatalErrorEventType represents a fatal error
	FatalErrorEventType EventType = "fatal-error"
	// ErrorEventType represents a regular error
	ErrorEventType EventType = "error"
)

// nolint:revive
// Event data structure representing any event we may want to react to
type Event struct {
	Type    EventType
	Payload any
}
