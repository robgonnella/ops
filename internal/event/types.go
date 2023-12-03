package event

type EventType string

const (
	FatalErrorEventType EventType = "fatal-error"
)

// Event data structure representing any event we may want to react to
type Event struct {
	Type    EventType
	Payload any
}
