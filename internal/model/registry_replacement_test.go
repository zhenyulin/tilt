package model

import (
	"fmt"
	"testing"

	"github.com/windmilleng/tilt/internal/container"

	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
	rep      RegistryReplacement
	name     string
	expected string
}{
	{RegistryReplacement{"gcr.io", "myreg.com"}, "gcr.io/foo/bar:deadbeef", "myreg.com/foo/bar:deadbeef"},
	{RegistryReplacement{"other.com", "myreg.com"}, "gcr.io/foo/bar:deadbeef", "gcr.io/foo/bar:deadbeef"},
	{RegistryReplacement{"gcr.io/baz/foo/bar", "aws_account_id.dkr.ecr.region.amazonaws.com/bar"}, "gcr.io/baz/foo/bar:deadbeef", "aws_account_id.dkr.ecr.region.amazonaws.com/bar:deadbeef"},
}

func TestReplaceTaggedRefDomain(t *testing.T) {
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test Case #%d", i), func(t *testing.T) {
			name := container.MustParseNamedTagged(tc.name)
			actual, err := ReplaceNamedTagged(tc.rep, name)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual.String())
		})
	}
}
