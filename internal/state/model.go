package state

import (
	"time"
)

// XXX(dbentley): rename to State
type Resources struct {
	Resources map[string]Resource
	RunQueue  []string
	Running   SpanID
	Last      SpanID
}

type Resource struct {
	Name        string
	K8sYaml     string
	QueuedFiles []string
}

type SingleSpanWriter interface {
	ID() SpanID

	LogKV(kvs ...string)

	StartChild(name string) SingleSpanWriter

	Finish()
	FinishErr(err error)
}

type SpanID int64

const NoSpanID SpanID = 0

type Span struct {
	ID SpanID

	// set at start
	Parent SpanID
	Name   string
	Begin  time.Time

	// set as it goes on
	Fields map[string]string

	// set at finish
	Err string // string so it can be serialized; empty means no error
	End time.Time
}
