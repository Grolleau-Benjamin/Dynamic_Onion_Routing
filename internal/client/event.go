package client

type EventType int

const (
	EvLog EventType = iota
	EvErr
)

type Event struct {
	Type    EventType
	Payload interface{}
}
