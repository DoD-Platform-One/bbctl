package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		format   output.Format
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "name: Test\nstatuses:\n  - name: Status Test Output\n    output:\n      - Output one, Output two\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: `{"name":"Test","statuses":[{"name":"Status Test Output","output":["Output one, Output two"]}]}`,
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
				actual, err := statusOutput.EncodeYAML()
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := statusOutput.EncodeJSON()
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := statusOutput.EncodeText()
				require.NoError(t, err)
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
		format   output.Format
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "name: Status Test Output\noutput:\n  - Output one, Output two\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: `{"name":"Status Test Output","output":["Output one, Output two"]}`,
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
				actual, err := checkStatus.EncodeYAML()
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := checkStatus.EncodeJSON()
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := checkStatus.EncodeText()
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}
