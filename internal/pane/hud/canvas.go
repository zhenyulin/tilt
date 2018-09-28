package hud

import (
	"github.com/gdamore/tcell"
)

type Canvas interface {
	SetContent(x int, y int, ch rune, comb []rune, style tcell.Style)
	GetContent(x, y int) (rune, []rune, tcell.Style, int)

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

func (c *subCanvas) GetContent(x, y int) (rune, []rune, tcell.Style, int) {
	return c.back.GetContent(x+c.x, y+c.y)
}

func (c *subCanvas) Size() (int, int) {
	return c.width, c.height
}

func divideCanvas(c Canvas, x, y, width, height int) Canvas {
	return &subCanvas{c, x, y, width, height}
}

type cell struct {
	ch    rune
	style tcell.Style
}

type bufferCanvas struct {
	width int
	lines [][]cell
}

func newBufferCanvas(width int) *bufferCanvas {
	return &bufferCanvas{width: width}
}

func (c *bufferCanvas) SetContent(x int, y int, ch rune, comb []rune, style tcell.Style) {
	c.ensure(y)

	cell := c.lines[y][x]
	cell.ch = ch
	cell.style = style
	c.lines[y][x] = cell
}

func (c *bufferCanvas) GetContent(x, y int) (rune, []rune, tcell.Style, int) {
	c.ensure(y)
	cell := c.lines[y][x]
	return cell.ch, nil, cell.style, 1
}

func (c *bufferCanvas) ensure(y int) {
	for y >= len(c.lines) {
		line := make([]cell, c.width)
		c.lines = append(c.lines, line)
	}
}

func (c *bufferCanvas) Size() (int, int) {
	return c.width, len(c.lines)
}

func copyToFrom(dest, src Canvas) {
	width, height := src.Size()
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			ch, comb, style, _ := src.GetContent(i, j)
			dest.SetContent(i, j, ch, comb, style)
		}
	}
}
