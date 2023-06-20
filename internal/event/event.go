package event

type EventType string

const (
	SeverUpdate EventType = "SERVER_UPDATE"
)

type Event struct {
	Type    EventType
	Payload any
}
