package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestOutputClient(t *testing.T) {
	tests := []struct {
		name      string
		marshaler Outputable
		writer    io.Writer
		data      interface{}
		format    OutputFormat
		expected  string
		wantErr   bool
	}{
		{
			name:     "UnsupportedOutput",
			data:     map[string]string{"key": "value"},
			format:   "graph",
			expected: "unsupported format:",
			wantErr:  true,
		},
		{
			name:     "HumanReadableOutput",
			data:     map[string]string{"key": "value"},
			format:   "text",
			expected: "Vals: map[key:value]\n",
			wantErr:  false,
		},
		{
			name:     "JSON",
			data:     map[string]string{"key": "value"},
			format:   "json",
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "YAML",
			data:     map[string]string{"key": "value"},
			format:   "yaml",
			expected: "key: value\n",
			wantErr:  false,
		},
		{
			name:      "HumanReadable_MarshalError",
			data:      map[string]string{"key": "value"},
			marshaler: &errorOutput{},
			format:    "text",
			expected:  "unable to marshal data",
			wantErr:   true,
		},
		{
			name:      "JSON_MarshalError",
			data:      map[string]string{"key": "value"},
			marshaler: &errorOutput{},
			format:    "json",
			expected:  "unable to marshal data",
			wantErr:   true,
		},
		{
			name:      "YAML_MarshalError",
			data:      map[string]string{"key": "value"},
			marshaler: &errorOutput{},
			format:    "yaml",
			expected:  "unable to marshal data",
			wantErr:   true,
		},
		{
			name:     "HumanReadable_WriterError",
			data:     map[string]string{"key": "value"},
			writer:   &errorWriter{},
			format:   "text",
			expected: "data is bad",
			wantErr:  true,
		},

		{
			name:     "JSON_WriterError",
			data:     map[string]string{"key": "value"},
			writer:   &errorWriter{},
			format:   "json",
			expected: "data is bad",
			wantErr:  true,
		},
		{
			name:     "YAML_WriterError",
			data:     map[string]string{"key": "value"},
			writer:   &errorWriter{},
			format:   "yaml",
			expected: "data is bad",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams, _, _, _ := genericIOOptions.NewTestIOStreams()
			if tt.writer != nil {
				streams.Out = tt.writer
			}

			client := NewOutputClient(tt.format, streams)

			var data Outputable

			if tt.marshaler != nil {
				data = tt.marshaler
			} else {
				data = &testOutput{
					Vals: tt.data,
				}
			}

			err := client.Output(data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expected)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, streams.Out.(*bytes.Buffer).String())

		})
	}
}

type errorOutput struct {
	Vals interface{}
}

func (to *errorOutput) MarshalYaml() ([]byte, error) {
	return nil, errors.New("unable to marshal data")
}

func (to *errorOutput) MarshalJson() ([]byte, error) {
	return nil, errors.New("unable to marshal data")

}

func (to *errorOutput) MarshalHumanReadable() (string, error) {
	return "", errors.New("unable to marshal data")

}

type testOutput struct {
	Vals interface{}
}

func (to *testOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(to.Vals)
}

func (to *testOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(to.Vals)
}

func (to *testOutput) MarshalHumanReadable() (string, error) {
	return to.String(), nil
}

func (to *testOutput) String() string {
	return fmt.Sprintf("Vals: %s", to.Vals)
}

type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("data is bad")
}
