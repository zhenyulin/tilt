package hud

import (
	"fmt"
	"sort"
)

type navigationState struct {
	selectedResource string
}

type action int

const (
	noopAction action = iota
	downAction
	upAction
)

func handleNavigation(c navigationState, h *Hud, a action) navigationState {
	var keys []string
	for k, _ := range h.resources {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	idx := -1
	for i, k := range keys {
		if k == c.selectedResource {
			idx = i
		}
	}

	if idx == -1 && len(keys) > 0 {
		c.selectedResource = keys[0]
		return c
	}

	switch a {
	case noopAction:
		return c
	case downAction:
		if idx+1 >= len(keys) {
			// already at the bottom
			return c
		}
		c.selectedResource = keys[idx+1]
		return c
	case upAction:
		if idx == 0 {
			return c
		}
		c.selectedResource = keys[idx-1]
		return c
	default:
		panic(fmt.Errorf("handleNavigation: unknown action %v", a))
	}

}
