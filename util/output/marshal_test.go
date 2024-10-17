package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicOutput_EncodeYAML(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]interface{}
		expected  string
		wantPanic bool
	}{
		{
			name:      "ValidData",
			input:     map[string]interface{}{"key": "value"},
			expected:  "key: value\n",
			wantPanic: false,
		},
		{
			name:      "EmptyData",
			input:     map[string]interface{}{},
			expected:  "{}\n",
			wantPanic: false,
		},
		{
			name:      "InvalidData",
			input:     map[string]interface{}{"key": make(chan int)}, // Invalid data type
			expected:  "",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("panic status mismatch, got panic: %v, expected panic: %v", r != nil, tt.wantPanic)
				}
			}()

			output := &BasicOutput{Vals: tt.input}

			yamlData, err := output.EncodeYAML()

			if tt.wantPanic {
				if err != nil {
					require.Error(t, err)
					assert.Nil(t, yamlData)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(yamlData))
			}
		})
	}
}

func TestBasicOutput_EncodeJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "ValidData",
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "EmptyData",
			input:    map[string]interface{}{},
			expected: "{}",
			wantErr:  false,
		},
		{
			name:     "InvalidData",
			input:    map[string]interface{}{"key": make(chan int)}, // Invalid data type
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &BasicOutput{Vals: tt.input}

			jsonData, err := output.EncodeJSON()

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, jsonData)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(jsonData))
			}
		})
	}
}

func TestBasicOutput_EncodeText(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		expected      string
		expectedError string
	}{
		{
			name: "ValidData",
			input: map[string]interface{}{
				"key": "value",
			},
			expected:      "Vals: map[key:value]",
			expectedError: "",
		},
		{
			name:  "EmptyData",
			input: map[string]interface{}{
				// Empty map
			},
			expected:      "Vals: map[]",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &BasicOutput{Vals: tt.input}

			text, err := output.EncodeText()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Empty(t, text)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(text))
			}
		})
	}
}
