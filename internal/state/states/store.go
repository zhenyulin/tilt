package states

import (
	"context"

	"github.com/windmilleng/tilt/internal/state"
)

func NewStateStore(ctx context.Context) (*StateStore, error) {
	s := &StateStore{
		inCh:  make(chan state.Event),
		outCh: make(chan []state.Event),
	}
	go s.loop()
	return s, nil
}

type StateStore struct {
	inCh  chan state.Event
	outCh chan []state.Event
}

func (s *StateStore) Write(ctx context.Context, ev state.Event) error {
	s.inCh <- ev
	return nil
}

func (s *StateStore) Subscribe(ctx context.Context) (chan []state.Event, error) {
	// TODO(dbentley): what if someone else has already read
	return s.outCh, nil
}

func (s *StateStore) loop() {
	var evs []state.Event
	for {
		outCh := s.outCh
		if len(evs) == 0 {
			// nothing to send; don't try
			outCh = nil
		}
		select {
		case ev := <-s.inCh:
			evs = append(evs, ev)
		case outCh <- evs:
			evs = nil
		}
	}
}

var _ state.StateWriter = (*StateStore)(nil)
var _ state.StateReader = (*StateStore)(nil)
