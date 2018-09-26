package hud

import (
	"github.com/gdamore/tcell"
)

type pen struct {
	back Canvas
	x    int
	y    int
}

func (p *pen) write(s string) {
	for _, ch := range s {
		p.back.SetContent(p.x, p.y, ch, nil, tcell.StyleDefault)
		p.x += 1
	}
}

func newPen(c Canvas, x int, y int) *pen {
	return &pen{c, x, y}
}
