package update

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:thelper // These are not helpers. I want to point to the line in the report functions.
func TestNewVersion(t *testing.T) {
	tests := map[string]struct {
		input  string
		report func(*testing.T, Version, error)
	}{
		"valid semver": {
			input: "1.2.3",
			report: func(t *testing.T, v Version, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, v.Major)
				assert.Equal(t, 2, v.Minor)
				assert.Equal(t, 3, v.Patch)
				assert.Equal(t, "1.2.3", v.String())
			},
		},
		"valid semver with leading v": {
			input: "v1.2.3",
			report: func(t *testing.T, v Version, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, v.Major)
				assert.Equal(t, 2, v.Minor)
				assert.Equal(t, 3, v.Patch)
				assert.Equal(t, "1.2.3", v.String())
			},
		},
		"valid semver with prerelease": {
			input: "1.2.3-alpha.1",
			report: func(t *testing.T, v Version, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, v.Major)
				assert.Equal(t, 2, v.Minor)
				assert.Equal(t, 3, v.Patch)
				assert.Equal(t, "alpha.1", v.Prerelease)
				assert.Equal(t, "1.2.3-alpha.1", v.String())
			},
		},
		"valid semver with build metadata": {
			input: "1.2.3+build.1",
			report: func(t *testing.T, v Version, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, v.Major)
				assert.Equal(t, 2, v.Minor)
				assert.Equal(t, 3, v.Patch)
				assert.Equal(t, "build.1", v.BuildMetadata)
				assert.Equal(t, "1.2.3+build.1", v.String())
			},
		},
		"valid semver with prerelease and build metadata": {
			input: "1.2.3-alpha.1+build.1",
			report: func(t *testing.T, v Version, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, v.Major)
				assert.Equal(t, 2, v.Minor)
				assert.Equal(t, 3, v.Patch)
				assert.Equal(t, "alpha.1", v.Prerelease)
				assert.Equal(t, "build.1", v.BuildMetadata)
				assert.Equal(t, "1.2.3-alpha.1+build.1", v.String())
			},
		},
		"invalid semver": {
			input: "1.2",
			report: func(t *testing.T, _ Version, err error) {
				assert.Error(t, err)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := NewVersion(test.input, time.Now())
			test.report(t, v, err)
		})
	}
}
