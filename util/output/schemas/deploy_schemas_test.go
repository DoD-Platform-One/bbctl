package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	output "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

func TestBBDeployOutputFormat(t *testing.T) {
	helmOutput := HelmOutput{
		Message:      "Testing Output",
		Name:         "Helm Test Output",
		LastDeployed: "2024-01-01T00:00:00Z",
		Namespace:    "test",
		Status:       "running",
		Revision:     "1",
		TestSuite:    "",
		Notes:        "",
	}
	bbOutput := BigbangOutput{Data: helmOutput}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "message: Testing Output\nname: Helm Test Output\nlastdeployed: \"2024-01-01T00:00:00Z\"\nnamespace: test\nstatus: running\nrevision: \"1\"\ntestsuite: \"\"\nnotes: \"\"\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: `{"Message":"Testing Output","Name":"Helm Test Output","LastDeployed":"2024-01-01T00:00:00Z","Namespace":"test","Status":"running","Revision":"1","TestSuite":"","Notes":""}`,
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "Message: Testing Output\nName: Helm Test Output\nLast Deployed: 2024-01-01T00:00:00Z\nNamespace: test\nStatus: running\nRevision: 1\nTest Suite: \nNotes:\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			switch tt.format {
			case output.YAML:
				actual, err := bbOutput.MarshalYaml()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := bbOutput.MarshalJson()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := bbOutput.MarshalHumanReadable()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}

func TestFluxDeployOutputFormat(t *testing.T) {
	fluxOutput := FluxOutput{
		Data: Output{
			GeneralInfo: map[string]string{
				"key":    "value",
				"config": "option",
			},
			Actions: []string{
				"Action 1",
				"Action 2",
			},
			Warnings: []string{
				"Warning 1",
				"Warning 2",
			},
		},
	}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "general_info:\n  config: option\n  key: value\nactions:\n- Action 1\n- Action 2\nwarnings:\n- Warning 1\n- Warning 2\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: `{"general_info":{"config":"option","key":"value"},"actions":["Action 1","Action 2"],"warnings":["Warning 1","Warning 2"]}`,
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "General Info:\n  key: value\n  config: option\n\nActions:\n  Action 1\n  Action 2\n\nWarnings:\n  Warning 1\n  Warning 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			switch tt.format {
			case output.YAML:
				actual, err := fluxOutput.MarshalYaml()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := fluxOutput.MarshalJson()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := fluxOutput.MarshalHumanReadable()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}
