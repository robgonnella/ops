package event

// EventType represents the possible types of events
type EventType string

const (
	// SeverUpdate event type representing a server update
	SeverUpdate EventType = "SERVER_UPDATE"
)

// Event data structure representing any event we may want to react to
type Event struct {
	Type    EventType
	Payload any
}
