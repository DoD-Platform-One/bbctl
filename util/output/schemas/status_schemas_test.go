package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	output "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

func TestStatusOutputFormat(t *testing.T) {
	statuses := []string{"Output one, Output two"}
	checkStatus := CheckStatusOutput{
		Name:   "Status Test Output",
		Output: statuses,
	}
	statusOutput := StatusOutput{Name: "Test"}
	statusOutput.Statuses = append(statusOutput.Statuses, checkStatus)
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "name: Test\nstatuses:\n- name: Status Test Output\n  output:\n  - Output one, Output two\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: `{"Name":"Test","Statuses":[{"Name":"Status Test Output","Output":["Output one, Output two"]}]}`,
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "\n\nTest\n\nStatuses: [\n\nName: Status Test Output\nOutput:\n    Output one, Output two\n]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.format {
			case output.YAML:
				actual, err := statusOutput.MarshalYaml()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := statusOutput.MarshalJson()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := statusOutput.MarshalHumanReadable()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}

func TestCheckStatusOutputFormat(t *testing.T) {
	statuses := []string{"Output one, Output two"}
	checkStatus := CheckStatusOutput{
		Name:   "Status Test Output",
		Output: statuses,
	}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "name: Status Test Output\noutput:\n- Output one, Output two\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: `{"Name":"Status Test Output","Output":["Output one, Output two"]}`,
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "\n\nName: Status Test Output\nOutput:\n    Output one, Output two\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.format {
			case output.YAML:
				actual, err := checkStatus.MarshalYaml()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := checkStatus.MarshalJson()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := checkStatus.MarshalHumanReadable()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}
