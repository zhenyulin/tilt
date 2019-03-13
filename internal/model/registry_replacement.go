package model

import (
	"fmt"

	"github.com/docker/distribution/reference"
)

type RegistryReplacement struct {
	Old string
	New string
}

// Replace takes performs a RegistryReplacement and a NamedTagged and returns a new NamedTag with the domain swapped out
func Replace(rep RegistryReplacement, name reference.NamedTagged) (reference.NamedTagged, error) {
	if reference.Domain(name) == rep.Old {
		path := reference.Path(name)

		new, err := reference.ParseNamed(fmt.Sprintf("%s/%s", rep.New, path))
		if err != nil {
			return nil, err
		}

		newTagged, err := reference.WithTag(new, name.Tag())
		if err != nil {
			return nil, err
		}

		return newTagged, nil
	}
	return name, nil
}
