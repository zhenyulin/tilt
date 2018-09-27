package hud

import (
	"context"
	"fmt"
	"os"

	"log"

	"github.com/gdamore/tcell"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/windmilleng/tilt/internal/k8s"
	"github.com/windmilleng/tilt/internal/state"
)

type Hud struct {
	resources state.Resources
	k8s       map[string]interface{}
	spans     map[state.SpanID]state.Span

	nav navigationState

	evs       chan []state.Event
	controlCh chan<- state.ControlEvent

	screen tcell.Screen
}

func NewHud(evs chan []state.Event, controlCh chan<- state.ControlEvent) (*Hud, error) {
	return &Hud{
		evs:       evs,
		k8s:       make(map[string]interface{}),
		spans:     make(map[state.SpanID]state.Span),
		controlCh: controlCh,
	}, nil
}

func (h *Hud) Run(outerCtx context.Context, readTty *os.File, writeTty *os.File, winchCh chan os.Signal) error {
	// initialize the screen
	screen, err := tcell.NewTerminfoScreenFromTty(readTty, writeTty, winchCh)
	if err != nil {
		log.Printf("can't start screen")
	}

	if err := screen.Init(); err != nil {
		log.Printf("can't init screen %v", err)
	}
	defer screen.Fini()
	h.screen = screen

	innerCtx, cancel := context.WithCancel(outerCtx)
	defer cancel()

	evCh := make(chan tcell.Event, 1)
	go func() {
		for {
			ev := screen.PollEvent()
			if ev == nil {
				close(evCh)
				return
			}
			select {
			case <-innerCtx.Done():
				close(evCh)
				return
			case evCh <- ev:
			}
		}
	}()

	for {
		select {
		case <-outerCtx.Done():
			// TODO(dbentley): cleanup
			return nil
		case ev := <-evCh:
			if err := h.handleTcellEvent(ev); err != nil {
				return err
			}
		case evs := <-h.evs:
			if err := h.handleTiltEvents(evs); err != nil {
				return err
			}
		}

		h.nav = handleNavigation(h.nav, h, noopAction)
		h.render()
	}
}

func (h *Hud) handleTcellEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC:
			return fmt.Errorf("quit")
		case tcell.KeyDown:
			h.nav = handleNavigation(h.nav, h, downAction)
		case tcell.KeyUp:
			h.nav = handleNavigation(h.nav, h, upAction)
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				return fmt.Errorf("quit")
			case 'r':
				h.controlCh <- state.RunWorkflowEvent{ResourceName: h.nav.selectedResource}
			}
		}
	}
	return nil
}

func (h *Hud) handleTiltEvents(evs []state.Event) error {
	for _, ev := range evs {
		switch ev := ev.(type) {
		case state.ResourcesEvent:
			h.resources = ev.Resources
		case state.KubeEvent:
			switch ev := ev.Event.(type) {
			case k8s.AddInformEvent:
				h.setK8sObject(ev.Obj)
			case k8s.UpdateInformEvent:
				h.setK8sObject(ev.NewObj)
			case k8s.DeleteInformEvent:
				log.Printf("k8s delete!!! Unhandled!: %T %+v", ev.Last, ev.Last)
			}
		case state.SpanEvent:
			h.spans[ev.Span.ID] = ev.Span
		default:
			return fmt.Errorf("hud.HandleTiltEvents: unexpected event %T %v", ev, ev)
		}
	}
	return nil
}

func (h *Hud) setK8sObject(obj interface{}) {
	switch obj := obj.(type) {
	case *appsv1.Deployment:
		h.k8s[obj.SelfLink] = obj
	default:
		log.Printf("unknown k8s object type: %T %+v", obj)
	}
}
