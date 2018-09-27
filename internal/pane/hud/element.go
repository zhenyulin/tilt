package hud

import (
	"fmt"
	"sort"

	// "log"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/windmilleng/tilt/internal/k8s"
	"github.com/windmilleng/tilt/internal/state"
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
		r := h.resources[k]
		h.renderResource(rc, r)
	}

}

func (h *Hud) renderResource(c Canvas, r state.Resource) {
	p := newPen(c, 0, 0)
	if h.nav.selectedResource == r.Name {
		p.write("> ")
	} else {
		p.write("  ")
	}
	p.write(fmt.Sprintf("Resource: %v", r.Name))
	p = newPen(c, 4, 1)
	p.write(fmt.Sprintf("Queued files: %q", r.QueuedFiles))
	p = newPen(c, 4, 2)

	entities, err := k8s.ParseYAMLFromString(r.K8sYaml)
	if err != nil {
		panic(err)
	}

	// other := 0
	var deploymentEntity k8s.K8sEntity
	var serviceEntity k8s.K8sEntity
	for _, e := range entities {
		if e.Kind.Kind == "Deployment" {
			deploymentEntity = e
		}
		if e.Kind.Kind == "Service" {
			serviceEntity = e
		}
	}
	p.write("k8s: [ ")
	if deploymentEntity.Obj != nil {
		e := deploymentEntity
		p.write(fmt.Sprintf("de/%v", deploymentEntity.Name()))
		k8s, ok := h.k8s["/apis/apps/v1/namespaces/default/deployments/"+e.Name()]
		if ok {
			if d, ok := k8s.(*appsv1.Deployment); ok {
				p.write(
					fmt.Sprintf("(%d/%d) ", d.Status.AvailableReplicas, d.Status.Replicas))
			}
		}
		p.write(" ")
	}
	if serviceEntity.Obj != nil {
		p.write(fmt.Sprintf("svc/%v ", serviceEntity.Name()))
	}
	p.write("]")
}
