package state

import (
	"context"

	"github.com/windmilleng/tilt/internal/k8s"
)

type Event interface {
	event()
}

type ResourcesEvent struct {
	Resources Resources
}

type KubeEvent struct {
	Event k8s.InformEvent
}

type SpanEvent struct {
	Span Span
}

func (ResourcesEvent) event() {}
func (KubeEvent) event()      {}
func (SpanEvent) event()      {}

type StateWriter interface {
	StartRootSpan(name string) SingleSpanWriter
	Write(ctx context.Context, ev Event) error
}

type StateReader interface {
	Subscribe(ctx context.Context) (chan []Event, error)
}

type Subscription interface {
	Ch() chan Event
}

type tiltContextKeyStruct struct{}

var tiltContextKey tiltContextKeyStruct

func StartRootSpanFromContext(ctx context.Context, w StateWriter, name string) (SingleSpanWriter, context.Context) {
	s := w.StartRootSpan(name)
	return s, context.WithValue(ctx, tiltContextKey, s)
}

func StartSpanFromContext(ctx context.Context, name string) (SingleSpanWriter, context.Context) {
	w := GetSpan(ctx)
	s := w.StartChild(name)
	return s, context.WithValue(ctx, tiltContextKey, s)
}

func GetSpan(ctx context.Context) SingleSpanWriter {
	v := ctx.Value(tiltContextKey)
	if v == nil {
		return nil
	}
	return v.(SingleSpanWriter)
}
