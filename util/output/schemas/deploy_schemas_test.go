package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBigbangOutput_Marshall(t *testing.T) {
	testObject := &BigbangOutput{
		Data: HelmOutput{
			Message:      "test",
			Name:         "test",
			LastDeployed: "test",
			Namespace:    "test",
			Status:       "test",
			Revision:     "test",
			TestSuite:    "test",
			Notes:        "test",
		},
	}
	tests := []struct {
		name     string
		input    *BigbangOutput
		expected string
	}{
		{
			name:     "YAML",
			input:    testObject,
			expected: "message: test\nname: test\nlastdeployed: test\nnamespace: test\nstatus: test\nrevision: test\ntestsuite: test\nnotes: test\n",
		},
		{
			name:     "JSON",
			input:    testObject,
			expected: "{\"Message\":\"test\",\"Name\":\"test\",\"LastDeployed\":\"test\",\"Namespace\":\"test\",\"Status\":\"test\",\"Revision\":\"test\",\"TestSuite\":\"test\",\"Notes\":\"test\"}",
		},
		{
			name:     "HumanReadable",
			input:    testObject,
			expected: "Message: test\nName: test\nLast Deployed: test\nNamespace: test\nStatus: test\nRevision: test\nTest Suite: test\nNotes:\ntest\n",
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

func TestFluxOutput_Marshall(t *testing.T) {
	testObject := &FluxOutput{
		Data: Output{
			GeneralInfo: map[string]string{
				"test": "test",
			},
			Actions:  []string{"test"},
			Warnings: []string{"test"},
		},
	}

	tests := []struct {
		name     string
		input    *FluxOutput
		expected string
	}{
		{
			name:     "YAML",
			input:    testObject,
			expected: "general_info:\n  test: test\nactions:\n- test\nwarnings:\n- test\n",
		},
		{
			name:     "JSON",
			input:    testObject,
			expected: "{\"general_info\":{\"test\":\"test\"},\"actions\":[\"test\"],\"warnings\":[\"test\"]}",
		},
		{
			name:     "HumanReadable",
			input:    testObject,
			expected: "General Info:\n  test: test\n\nActions:\n  test\n\nWarnings:\n  test\n",
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
