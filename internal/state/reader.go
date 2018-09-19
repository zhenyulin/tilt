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

func (ResourcesEvent) event() {}

type StateWriter interface {
	WriteState(ctx context.Context, resources Resources) error
}

type StateReader interface {
	Subscribe(ctx context.Context) (Subscription, error)
}

type Subscription interface {
	Ch() chan Event
}
