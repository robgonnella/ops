package event

//go:generate mockgen -destination=../mock/event/event.go -package=mock_event . Manager
type Manager interface {
	RegisterListener(eventType EventType, listener chan Event) int
	RemoveListener(id int) int
	Send(event Event)
	ReportFatalError(err error)
	ReportError(err error)
}
