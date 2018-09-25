package states

import (
	"context"
	"fmt"
	"log"

	"github.com/windmilleng/tilt/internal/state"
)

func NewStateStore(ctx context.Context) (*StateStore, error) {
	s := &StateStore{
		ch: make(chan state.Event),
	}
	go s.loop()
	return s, nil
}

type StateStore struct {
	ctx context.Context
	ch  chan state.Event

	// below is only accessed from loop()
	resources state.Resources
}

func (s *StateStore) WriteState(ctx context.Context, resources state.Resources) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.ch <- state.ResourcesEvent{Resources: resources}:
		return nil
	}
}

func (s *StateStore) Subscribe(ctx context.Context) (state.Subscription, error) {
	return nil, fmt.Errorf("StateStore.Subscribe: not yet implemented")
}

func (s *StateStore) loop() {
	for ev := range s.ch {
		switch ev := ev.(type) {
		case state.ResourcesEvent:
			log.Printf("got resources!")
		default:
			log.Printf("BAD %T %v", ev, ev)
		}
	}
}

var _ state.StateWriter = (*StateStore)(nil)
var _ state.StateReader = (*StateStore)(nil)
