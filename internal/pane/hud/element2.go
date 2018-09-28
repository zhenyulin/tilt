package hud

import ()

type ScrollingElement struct {
	els []FixedWidthElement
}

func (e *ScrollingElement) RenderFixed(c Canvas) {
	width, _ := c.Size()

	var cs []Canvas
	for _, el := range e.els {
		cs = append(cs, el.RenderFixedWidth(width))
	}

}

type FixedWidthElement interface {
	RenderFixedWidth(width int) Canvas
}

type TextElement struct {
	text string
}

func NewTextElement(text string) *TextElement {
	return &TextElement{text: text}
}

func (e *TextElement) RenderFixed(c Canvas) {
	width, _ := c.Size()
	contentCanvas := e.RenderFixedWidth(width)
	copyToFrom(c, contentCanvas)
}

func (e *TextElement) RenderFixedWidth(width int) Canvas {
	c := newBufferCanvas(width)
	p := newPen(c, 0, 0)
	p.write(e.text)
	return c
}
