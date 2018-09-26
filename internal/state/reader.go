package state

import (
	"context"
)

type Event interface {
	event()
}

type ResourcesEvent struct {
	Resources Resources
}

type ResourceEvent struct {
	Resource Resource
}

func (ResourcesEvent) event() {}
func (ResourceEvent) event()  {}

type StateWriter interface {
	Write(ctx context.Context, ev Event) error
}

type StateReader interface {
	Subscribe(ctx context.Context) (chan []Event, error)
}

type Subscription interface {
	Ch() chan Event
}
