package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckStepOutput_Marshall(t *testing.T) {
	tests := []struct {
		name     string
		input    *CheckStepOutput
		expected string
	}{
		{
			name: "YAML",
			input: &CheckStepOutput{
				Name:   "test",
				Output: []string{"output1", "output2"},
				Status: "pass",
			},
			expected: "name: test\noutput:\n- output1\n- output2\nstatus: pass\n",
		},
		{
			name: "JSON",
			input: &CheckStepOutput{
				Name:   "test",
				Output: []string{"output1", "output2"},
				Status: "pass",
			},
			expected: "{\"Name\":\"test\",\"Output\":[\"output1\",\"output2\"],\"Status\":\"pass\"}",
		},
		{
			name: "HumanReadable",
			input: &CheckStepOutput{
				Name:   "test",
				Output: []string{"output1", "output2"},
				Status: "pass",
			},
			expected: "\n\nName: test\nOutput:\n    output1\n    output2\nStatus: pass\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			// Act
			var actual []byte
			var err error
			switch test.name {
			case "YAML":
				actual, err = test.input.MarshalYaml()
			case "JSON":
				actual, err = test.input.MarshalJson()
			case "HumanReadable":
				actual, err = test.input.MarshalHumanReadable()
			}
			// Assert
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}

func TestPreflightCheckOutput_Marshall(t *testing.T) {
	tests := []struct {
		name     string
		input    *PreflightCheckOutput
		expected string
	}{
		{
			name: "YAML",
			input: &PreflightCheckOutput{
				Name: "test",
				Steps: []CheckStepOutput{
					{
						Name:   "test",
						Output: []string{"output1", "output2"},
						Status: "pass",
					},
				},
			},
			expected: "name: test\nsteps:\n- name: test\n  output:\n  - output1\n  - output2\n  status: pass\n",
		},
		{
			name: "JSON",
			input: &PreflightCheckOutput{
				Name: "test",
				Steps: []CheckStepOutput{
					{
						Name:   "test",
						Output: []string{"output1", "output2"},
						Status: "pass",
					},
				},
			},
			expected: "{\"Name\":\"test\",\"Steps\":[{\"Name\":\"test\",\"Output\":[\"output1\",\"output2\"],\"Status\":\"pass\"}]}",
		},
		{
			name: "HumanReadable",
			input: &PreflightCheckOutput{
				Name: "test",
				Steps: []CheckStepOutput{
					{
						Name:   "test",
						Output: []string{"output1", "output2"},
						Status: "pass",
					},
				},
			},
			expected: "\n\ntest\n\nSteps: [\n\nName: test\nOutput:\n    output1\n    output2\nStatus: pass\n]\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			// Act
			var actual []byte
			var err error
			switch test.name {
			case "YAML":
				actual, err = test.input.MarshalYaml()
			case "JSON":
				actual, err = test.input.MarshalJson()
			case "HumanReadable":
				actual, err = test.input.MarshalHumanReadable()
			}
			// Assert
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}
