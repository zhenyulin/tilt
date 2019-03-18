package k8s

import (
	"testing"
)

func TestFindShortUnambiguousNames(t *testing.T) {

	cases := []IdTestCase{
		{"simple",
			ids(
				Id{"foo", "Deployment", "default", ""},
			),
			[]string{"foo"}},
		{"no conflicts",
			ids(
				Id{"foo", "Deployment", "default", ""},
				Id{"bar", "Deployment", "default", ""},
			),
			[]string{"foo", "bar"}},
		{"complex",
			ids(
				Id{"foo", "Deployment", "default", ""},
				// bar: two different kinds
				Id{"bar", "Deployment", "default", ""},
				Id{"bar", "Service", "default", ""},
				// baz: two deployments in different namespaces
				Id{"baz", "Deployment", "default", ""},
				Id{"baz", "Deployment", "user-dev", ""},
				// quux: two CRDs with the same Kind but different apigroups
				Id{"quux", "Config", "default", "foo.com"},
				Id{"quux", "Config", "default", "bar.com"},
			),
			[]string{
				"foo",
				"bar|Deployment", "bar|Service",
				"baz|Deployment|default", "baz|Deployment|user-dev",
				"quux|Config|default|foo.com", "quux|Config|default|bar.com",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Desc, func(t *testing.T) {
			actual := FindShortUnambiguousNames(c.Ids)
			expected := c.Expected
			if len(actual) != len(expected) {
				t.Fatalf("got short names %q for input %q; expected %q", actual, c.Ids, expected)
			}
			for i, a := range actual {
				e := expected[i]
				if a != e {
					t.Fatalf("got short name %q at index %d; expected %q for input %q; %q %q %q", a, i, e, c.Ids[i], actual, expected, c.Ids)
				}
			}
		})
	}
}

type IdTestCase struct {
	Desc     string
	Ids      []Id
	Expected []string
}

func ids(ids ...Id) []Id {
	return ids
}
