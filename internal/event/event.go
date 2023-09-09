package event

// Event data structure representing any event we may want to react to
type Event struct {
	Type    string
	Payload any
}
