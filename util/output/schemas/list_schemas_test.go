package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelmReleaseTableOuput_Marshal(t *testing.T) {
	testObject := HelmReleaseTableOutput{
		Releases: []HelmReleaseOutput{
			{
				Name:       "test",
				Namespace:  "test-ns",
				Revision:   1,
				Status:     "test-status",
				Chart:      "test-chart",
				AppVersion: "test-version",
			},
		},
	}

	tests := []struct {
		name     string
		marshal  func() ([]byte, error)
		expected string
	}{
		{
			name:     "YAML",
			marshal:  testObject.EncodeYAML,
			expected: "releases:\n- name: test\n  namespace: test-ns\n  revision: 1\n  status: test-status\n  chart: test-chart\n  appVersion: test-version\n",
		},
		{
			name:     "JSON",
			marshal:  testObject.EncodeJSON,
			expected: "{\"releases\":[{\"name\":\"test\",\"namespace\":\"test-ns\",\"revision\":1,\"status\":\"test-status\",\"chart\":\"test-chart\",\"appVersion\":\"test-version\"}]}",
		},
		{
			name:     "Text",
			marshal:  testObject.EncodeText,
			expected: "NAME\tNAMESPACE\tREVISION\tSTATUS     \tCHART     \tAPPVERSION  \ntest\ttest-ns  \t1       \ttest-status\ttest-chart\ttest-version",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.marshal()
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}
