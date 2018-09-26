package hud

import (
	"github.com/gdamore/tcell"
)

type Canvas interface {
	SetContent(x int, y int, ch rune, comb []rune, style tcell.Style)

	Size() (int, int)
}

type subCanvas struct {
	back   Canvas
	x      int
	y      int
	width  int
	height int
}

func (c *subCanvas) SetContent(x int, y int, ch rune, comb []rune, style tcell.Style) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}
	c.back.SetContent(x+c.x, y+c.y, ch, comb, style)
}

func (c *subCanvas) Size() (int, int) {
	return c.width, c.height
}

func divideCanvas(c Canvas, x, y, width, height int) Canvas {
	return &subCanvas{c, x, y, width, height}
}
