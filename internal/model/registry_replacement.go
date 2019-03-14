package model

import (
	"strings"

	"github.com/docker/distribution/reference"
)

type RegistryReplacement struct {
	Old string
	New string
}

// Replace takes performs a RegistryReplacement and a NamedTagged and returns a new NamedTag with Old swapped for New
// It errors if this isn't a valid NamedTagged
func ReplaceNamedTagged(rep RegistryReplacement, name reference.NamedTagged) (reference.NamedTagged, error) {
	ns := name.String()

	if !strings.Contains(ns, rep.Old) {
		return name, nil
	}

	newNs := strings.Replace(ns, rep.Old, rep.New, 1)
	newN, err := reference.ParseNamed(newNs)
	if err != nil {
		return nil, err
	}
	newNT, err := reference.WithTag(newN, name.Tag())
	if err != nil {
		return nil, err
	}

	return newNT, nil
}
