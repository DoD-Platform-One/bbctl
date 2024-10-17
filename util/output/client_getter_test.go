package output

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestGetClient(t *testing.T) {
	tests := []struct {
		name      string
		marshaler Outputable
		writer    io.Writer
		data      interface{}
		format    Format
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
			name:     "TextOutput",
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
			name:      "Text_MarshalError",
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
			name:     "Text_WriterError",
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

			clientGetter := ClientGetter{}
			client := clientGetter.GetClient(tt.format, streams)

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
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expected)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, streams.Out.(*bytes.Buffer).String())
		})
	}
}
