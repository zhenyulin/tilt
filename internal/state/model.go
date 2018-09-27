package state

import (
	"context"
)

type Resources struct {
	Resources []Resource
}

type Resource struct {
	Name        string
	K8sYaml     string
	QueuedFiles []string
}

type SingleSpanWriter interface {
	LogKV(kvs ...string)

	StartChild(name string) SingleSpanWriter

	Finish()
	FinishErr(err error)
}

type SpanID int64

type Span struct {
	ID SpanID

	// set at start
	Parent SpanID
	Name   string
	End    time.Time

	// set as it goes on
	kvs map[string]string

	// set at finish
	Err string // string so it can be serialized; empty means no error
	End time.Time
}
