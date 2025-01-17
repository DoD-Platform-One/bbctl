package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageVersion_Encode(t *testing.T) {
	allFields := PackageVersion{
		Version:         "1.0.0",
		LatestVersion:   "1.2.3",
		UpdateAvailable: true,
		SHAsMatch:       "All SHAs match",
	}

	noLatest := PackageVersion{
		Version:         "1.0.0",
		UpdateAvailable: true,
		SHAsMatch:       "All SHAs match",
	}

	noSHAsMatch := PackageVersion{
		Version:         "1.0.0",
		LatestVersion:   "1.2.3",
		UpdateAvailable: true,
	}

	tests := []struct {
		name       string
		testObject PackageVersion
		marshal    func(testObject PackageVersion) ([]byte, error)
		expected   string
	}{
		{
			name: "YAML All Fields",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeYAML()
			},
			testObject: allFields,
			expected:   "latestVersion: 1.2.3\nshasMatch: All SHAs match\nupdateAvailable: true\nversion: 1.0.0\n",
		},
		{
			name: "JSON All Fields",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeJSON()
			},
			testObject: allFields,
			expected:   "{\"latestVersion\":\"1.2.3\",\"shasMatch\":\"All SHAs match\",\"updateAvailable\":true,\"version\":\"1.0.0\"}",
		},
		{
			name: "HumanReadable All Fields",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeText()
			},
			testObject: allFields,
			expected:   "Version: 1.0.0\nLatest Version: 1.2.3\nUpdate Available: true\nSHAs Match: All SHAs match\n",
		},
		{
			name: "YAML No Latest",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeYAML()
			},
			testObject: noLatest,
			expected:   "shasMatch: All SHAs match\nversion: 1.0.0\n",
		},
		{
			name: "JSON No Latest",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeJSON()
			},
			testObject: noLatest,
			expected:   "{\"shasMatch\":\"All SHAs match\",\"version\":\"1.0.0\"}",
		},
		{
			name: "HumanReadable No Latest",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeText()
			},
			testObject: noLatest,
			expected:   "Version: 1.0.0\nSHAs Match: All SHAs match\n",
		},
		{
			name: "YAML No SHAs",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeYAML()
			},
			testObject: noSHAsMatch,
			expected:   "latestVersion: 1.2.3\nupdateAvailable: true\nversion: 1.0.0\n",
		},
		{
			name: "JSON No SHAs",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeJSON()
			},
			testObject: noSHAsMatch,
			expected:   "{\"latestVersion\":\"1.2.3\",\"updateAvailable\":true,\"version\":\"1.0.0\"}",
		},
		{
			name: "HumanReadable No SHAs",
			marshal: func(testObject PackageVersion) ([]byte, error) {
				return testObject.EncodeText()
			},
			testObject: noSHAsMatch,
			expected:   "Version: 1.0.0\nLatest Version: 1.2.3\nUpdate Available: true\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.marshal(test.testObject)
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}

func TestVersionOutput_Encode(t *testing.T) {
	testObject := VersionOutput{
		"grafana": PackageVersion{
			Version:         "1.0.0",
			LatestVersion:   "1.2.3",
			UpdateAvailable: true,
			SHAsMatch:       "All SHAs match",
		},
		"bigbang": PackageVersion{
			Version:         "1.0.0",
			LatestVersion:   "1.2.3",
			UpdateAvailable: true,
		},
	}

	tests := []struct {
		name     string
		marshal  func() ([]byte, error)
		expected []string
	}{
		{
			name: "YAML",
			marshal: func() ([]byte, error) {
				return testObject.EncodeYAML()
			},
			expected: []string{"bigbang:\n  latestVersion: 1.2.3\n  updateAvailable: true\n  version: 1.0.0\ngrafana:\n  latestVersion: 1.2.3\n  shasMatch: All SHAs match\n  updateAvailable: true\n  version: 1.0.0\n"},
		},
		{
			name: "JSON",
			marshal: func() ([]byte, error) {
				return testObject.EncodeJSON()
			},
			expected: []string{"{\"bigbang\":{\"latestVersion\":\"1.2.3\",\"updateAvailable\":true,\"version\":\"1.0.0\"},\"grafana\":{\"latestVersion\":\"1.2.3\",\"shasMatch\":\"All SHAs match\",\"updateAvailable\":true,\"version\":\"1.0.0\"}}"},
		},
		{
			name: "HumanReadable",
			marshal: func() ([]byte, error) {
				return testObject.EncodeText()
			},
			expected: []string{"grafana:\nVersion: 1.0.0\nLatest Version: 1.2.3\nUpdate Available: true\nSHAs Match: All SHAs match", "bigbang:\nVersion: 1.0.0\nLatest Version: 1.2.3\nUpdate Available: true\n"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.marshal()
			require.NoError(t, err)
			for _, value := range test.expected {
				assert.Contains(t, string(actual), value)
			}
		})
	}
}
