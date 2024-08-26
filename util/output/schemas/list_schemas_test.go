package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmReleaseTableOuput_Marshal(t *testing.T) {
	testObject := HelmReleaseTableOutput{
		Releases: []HelmReleaseOutput{
			{
				Name:       "test",
				Namespace:  "test",
				Revision:   1,
				Status:     "test",
				Chart:      "test",
				AppVersion: "test",
			},
		},
	}

	tests := []struct {
		name     string
		marshal  func() ([]byte, error)
		expected string
	}{
		{
			name: "YAML",
			marshal: func() ([]byte, error) {
				return testObject.MarshalYaml()
			},
			expected: "releases:\n- name: test\n  namespace: test\n  revision: 1\n  status: test\n  chart: test\n  appversion: test\n",
		},
		{
			name: "JSON",
			marshal: func() ([]byte, error) {
				return testObject.MarshalJson()
			},
			expected: "{\"Releases\":[{\"Name\":\"test\",\"Namespace\":\"test\",\"Revision\":1,\"Status\":\"test\",\"Chart\":\"test\",\"AppVersion\":\"test\"}]}",
		},
		{
			name: "HumanReadable",
			marshal: func() ([]byte, error) {
				return testObject.MarshalHumanReadable()
			},
			expected: "NAME\tNAMESPACE\tREVISION\tSTATUS\tCHART\tAPPVERSION\ntest\ttest     \t1       \ttest  \ttest \ttest      ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.marshal()
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}
