package states

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/windmilleng/tilt/internal/state"
)

func NewStateStore(ctx context.Context) (*StateStore, error) {
	s := &StateStore{
		inCh:  make(chan state.Event),
		outCh: make(chan []state.Event),
		ider:  &ider{},
	}
	go s.loop()
	return s, nil
}

type StateStore struct {
	inCh  chan state.Event
	outCh chan []state.Event

	ider *ider
}

func (s *StateStore) Write(ctx context.Context, ev state.Event) error {
	s.inCh <- ev
	return nil
}

func (s *StateStore) StartRootSpan(name string) state.SingleSpanWriter {
	return newSpan(state.NoSpanID, name, s.ider, s.inCh)
}

func (s *StateStore) Subscribe(ctx context.Context) (chan []state.Event, error) {
	// TODO(dbentley): what if someone else has already read?
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

type StoreSpanWriter struct {
	span    state.Span
	ider    *ider
	writeCh chan state.Event
}

func (w *StoreSpanWriter) ID() state.SpanID {
	return w.span.ID
}

func (w *StoreSpanWriter) LogKV(kvs ...string) {
	if len(kvs)%2 != 0 {
		panic(fmt.Errorf("StoreSpanWriter called with an odd number of arguments: %d %+v", len(kvs), kvs))
	}

	if w.span.Fields == nil {
		w.span.Fields = make(map[string]string)
	}
	for i := 0; i < len(kvs); i += 2 {
		k, v := kvs[i], kvs[i+1]
		w.span.Fields[k] = v
	}

	w.send()
}

func (w *StoreSpanWriter) StartChild(name string) state.SingleSpanWriter {
	return newSpan(w.span.ID, name, w.ider, w.writeCh)
}

func (w *StoreSpanWriter) Finish() {
	w.FinishErr(nil)
}

func (w *StoreSpanWriter) FinishErr(err error) {
	w.span.End = time.Now()
	if err != nil {
		w.span.Err = err.Error()
	}
	w.send()
}

func (w *StoreSpanWriter) send() {
	span := w.span
	if span.Fields != nil {
		n := make(map[string]string, len(w.span.Fields))
		for k, v := range w.span.Fields {
			n[k] = v
		}
		span.Fields = n
	}
	w.writeCh <- state.SpanEvent{Span: span}
}

func newSpan(parent state.SpanID, name string, ider *ider, writeCh chan state.Event) *StoreSpanWriter {
	w := &StoreSpanWriter{
		span: state.Span{
			ID:     ider.spanID(),
			Parent: parent,
			Name:   name,
			Begin:  time.Now(),
		},
		ider:    ider,
		writeCh: writeCh,
	}

	w.send()
	return w
}

type ider struct {
	mu     sync.Mutex
	lastID state.SpanID
}

func (s *ider) spanID() state.SpanID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastID++
	return s.lastID
}
