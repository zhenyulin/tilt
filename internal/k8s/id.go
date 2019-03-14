package k8s

import (
	"fmt"
)

type Id struct {
	Name      string
	Kind      string
	Namespace string
	Group     string
}

// TODO(dbentley): use shortnames to shorten kind

func (i Id) NameAndKind() string {
	return fmt.Sprintf("%s|%s", i.Name, i.Kind)
}

func (i Id) NameKindNamespace() string {
	return fmt.Sprintf("%s|%s|%s", i.Name, i.Kind, i.Namespace)
}

func (i Id) Full() string {
	return fmt.Sprintf("%s|%s|%s|%s", i.Name, i.Kind, i.Namespace, i.Group)
}

func (i Id) Forms() []string {
	return []string{i.Name, i.NameAndKind(), i.NameKindNamespace(), i.Full()}
}

func FindShortUnambiguousNames(ids []Id) []string {
	seen := make(map[string]int)

	for _, id := range ids {
		for _, s := range id.Forms() {
			seen[s]++
		}
	}

	var r []string
	for _, id := range ids {
		for _, s := range id.Forms() {
			if seen[s] == 1 {
				r = append(r, s)
				break
			}
		}
	}

	return r
}
