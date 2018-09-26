package hud

import (
	"fmt"
	"sort"
)

type Element interface {
	Draw(c Canvas) error
}

type hudElement struct {
}

const headerHeight = 3
const footerHeight = 3

func (h *Hud) render() {
	h.screen.Clear()
	var c Canvas
	c = h.screen
	width, height := c.Size()

	headerCanvas := divideCanvas(c, 0, 0, width, headerHeight)
	p := newPen(headerCanvas, 0, 0)
	p.write("header!")

	mainPanelHeight := height - (headerHeight + footerHeight)
	resourcesCanvas := divideCanvas(c, 0, headerHeight, width/2, mainPanelHeight)
	h.renderResources(resourcesCanvas)

	streamCanvas := divideCanvas(c, width/2, headerHeight, width/2, mainPanelHeight)
	p = newPen(streamCanvas, 0, 0)
	p.write("stream")

	footerCanvas := divideCanvas(c, 0, height-footerHeight, width, footerHeight)
	p = newPen(footerCanvas, 0, 0)
	p.write("footer!")

	p = newPen(footerCanvas, 0, 1)
	p.write("second line!")

	h.screen.Show()
}

const resourceHeight = 6

func (h *Hud) renderResources(c Canvas) {
	width, _ := c.Size()
	var keys []string
	for k, _ := range h.resources {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for i, k := range keys {
		rc := divideCanvas(c, 0, resourceHeight*i, width, resourceHeight)
		p := newPen(rc, 0, 0)
		if h.nav.selectedResource == k {
			p.write("> ")
		} else {
			p.write("  ")
		}
		p.write(fmt.Sprintf("Resource: %v", k))
		p = newPen(rc, 2, 1)
		p.write(fmt.Sprintf("Queued files: %q", h.resources[k].QueuedFiles))
	}

}
