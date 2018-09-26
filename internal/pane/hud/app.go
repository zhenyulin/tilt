package hud

import (
	"context"
	"fmt"
	"os"

	"log"

	"github.com/gdamore/tcell"

	"github.com/windmilleng/tilt/internal/state"
)

type Hud struct {
	evs       chan []state.Event
	resources map[string]state.Resource

	nav navigationState

	controlCh chan<- state.ControlEvent

	screen tcell.Screen
}

func NewHud(evs chan []state.Event, controlCh chan<- state.ControlEvent) (*Hud, error) {
	return &Hud{
		evs:       evs,
		resources: make(map[string]state.Resource),
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
			for _, res := range ev.Resources.Resources {
				h.resources[res.Name] = res
			}
		default:
			return fmt.Errorf("hud.HandleTiltEvents: unexpected event %T %v", ev, ev)
		}
	}
	return nil
}
