package model

import (
	"fmt"
	"testing"

	"github.com/windmilleng/tilt/internal/container"

	"github.com/stretchr/testify/assert"
)

var namedTaggedTestCases = []struct {
	rep      RegistryReplacement
	name     string
	expected string
}{
	{RegistryReplacement{"gcr.io", "myreg.com"}, "gcr.io/foo/bar:deadbeef", "myreg.com/foo/bar:deadbeef"},
	{RegistryReplacement{"other.com", "myreg.com"}, "gcr.io/foo/bar:deadbeef", "gcr.io/foo/bar:deadbeef"},
	{RegistryReplacement{"gcr.io/baz/foo/bar", "aws_account_id.dkr.ecr.region.amazonaws.com/bar"}, "gcr.io/baz/foo/bar:deadbeef", "aws_account_id.dkr.ecr.region.amazonaws.com/bar:deadbeef"},
}

var namedTestCases = []struct {
	rep      RegistryReplacement
	name     string
	expected string
}{
	{RegistryReplacement{"gcr.io", "myreg.com"}, "gcr.io/foo/bar", "myreg.com/foo/bar"},
	{RegistryReplacement{"other.com", "myreg.com"}, "gcr.io/foo/bar", "gcr.io/foo/bar"},
	{RegistryReplacement{"gcr.io/baz/foo/bar", "aws_account_id.dkr.ecr.region.amazonaws.com/bar"}, "gcr.io/baz/foo/bar", "aws_account_id.dkr.ecr.region.amazonaws.com/bar"},
}

func TestReplaceTaggedRefDomain(t *testing.T) {
	for i, tc := range namedTaggedTestCases {
		t.Run(fmt.Sprintf("Test Case #%d", i), func(t *testing.T) {
			name := container.MustParseNamedTagged(tc.name)
			actual, err := ReplaceNamedTagged(tc.rep, name)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual.String())
		})
	}
}

func TestReplaceNamed(t *testing.T) {
	for i, tc := range namedTestCases {
		t.Run(fmt.Sprintf("Test case #%d", i), func(t *testing.T) {
			name := container.MustParseNamed(tc.name)
			actual, err := ReplaceNamed(tc.rep, name)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual.String())
		})
	}
}
