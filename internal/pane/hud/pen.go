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
	width, _ := p.back.Size()
	for _, ch := range s {
		if p.x >= width {
			p.x = 0
			p.y += 1
		}
		if ch == '\n' {
			p.x = 0
			p.y += 1
		} else {
			p.back.SetContent(p.x, p.y, ch, nil, tcell.StyleDefault)
			p.x += 1
		}
	}
}

func newPen(c Canvas, x int, y int) *pen {
	return &pen{c, x, y}
}
