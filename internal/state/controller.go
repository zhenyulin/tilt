package state

type ControlEvent interface {
	controlEvent()
}

type RunWorkflowEvent struct {
	ResourceName string
}

func (RunWorkflowEvent) controlEvent() {}

type ControlListener interface {
	Ch() <-chan ControlEvent
}
