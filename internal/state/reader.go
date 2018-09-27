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

type ResourceEvent struct {
	Resource Resource
}

type KubeEvent struct {
	Event k8s.InformEvent
}

type SpanEvent struct {
	Span Span
}

func (ResourcesEvent) event() {}
func (ResourceEvent) event()  {}
func (KubeEvent) event()      {}
func (SpanEvent) event()      {}

type StateWriter interface {
	StartRootSpanFromCtx(ctx context.Context, name string) (SingleSpanWriter, ctx)
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

func StartSpanFromContext(ctx context.Context, name string) (SingleSpanWriter, ctx) {
	w := GetSpan(ctx)
	s := w.StartChild(name)
	return s, context.WithValue(ctx, tiltContextKey, s)
}

func GetSpan(ctx context.Context) SingleSpanWriter {
	v := ctx.Value(tiltContextKey)
	if v == nil {
		panic("context has no span writer")
	}
	return v.(SingleSpanWriter)
}
