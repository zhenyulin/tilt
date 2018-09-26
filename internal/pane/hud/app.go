package hud

import (
	"context"
	"fmt"
	"os"

	"log"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Hud struct {
	screen tcell.Screen
}

func NewHud() (*Hud, error) {
	return &Hud{}, nil
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
			log.Printf("event! %T %v %+v", ev, ev, ev)
			if err := h.handleTcellEvent(ev); err != nil {
				return err
			}
		}

		h.render()
	}
}

func (h *Hud) handleTcellEvent(ev tcell.Event) error {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC:
			return fmt.Errorf("quit")
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				return fmt.Errorf("quit")
			}
		}
	}
	return nil
}

func (h *Hud) render() {
	width, height := h.screen.Size()
	f := tview.NewFlex()
	f.SetRect(0, 0, width, height)
	f.SetDirection(tview.FlexRow)

	header := tview.NewBox().SetBorder(true).SetTitle("header")
	f.AddItem(header, 4, 0, false)

	inner := tview.NewFlex()
	inner.AddItem(tview.NewBox().SetBorder(true).SetTitle("resources"), 0, 1, false)
	inner.AddItem(tview.NewBox().SetBorder(true).SetTitle("stream"), 0, 1, false)
	f.AddItem(inner, 0, 1, false)

	footer := tview.NewBox().SetBorder(true).SetTitle("footer")
	f.AddItem(footer, 4, 0, false)

	f.Draw(h.screen)
	h.screen.Show()
}
