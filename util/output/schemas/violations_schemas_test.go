package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	output "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

func TestViolationsOutputFormat(t *testing.T) {
	violationsOutput := &ViolationsOutput{Name: "Violations Test Output"}
	v1 := &Violation{
		Timestamp:  "a",
		Name:       "b",
		Kind:       "c",
		Namespace:  "d",
		Constraint: "e",
		Policy:     "f",
		Message:    "g",
		Action:     "h",
	}
	v2 := &Violation{
		Timestamp:  "i",
		Name:       "j",
		Kind:       "k",
		Namespace:  "l",
		Constraint: "m",
		Policy:     "n",
		Message:    "o",
		Action:     "p",
	}
	violationsOutput.Violations = append(violationsOutput.Violations, *v1)
	violationsOutput.Violations = append(violationsOutput.Violations, *v2)
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "name: Violations Test Output\nviolations:\n- name: b\n  kind: c\n  namespace: d\n  policy: f\n  constraint: e\n  message: g\n  action: h\n  timestamp: a\n- name: j\n  kind: k\n  namespace: l\n  policy: \"n\"\n  constraint: m\n  message: o\n  action: p\n  timestamp: i\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\"Name\":\"Violations Test Output\",\"Violations\":[{\"Name\":\"b\",\"Kind\":\"c\",\"Namespace\":\"d\",\"Policy\":\"f\",\"Constraint\":\"e\",\"Message\":\"g\",\"Action\":\"h\",\"Timestamp\":\"a\"},{\"Name\":\"j\",\"Kind\":\"k\",\"Namespace\":\"l\",\"Policy\":\"n\",\"Constraint\":\"m\",\"Message\":\"o\",\"Action\":\"p\",\"Timestamp\":\"i\"}]}",
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "Violations Test Output:\n  Resource: b\n  Kind: c\n  Namespace: d\n  Policy: f\n  Constraint: e\n  Message: g\n  Action: h\n  Timestamp: a\n\n  Resource: j\n  Kind: k\n  Namespace: l\n  Policy: n\n  Constraint: m\n  Message: o\n  Action: p\n  Timestamp: i\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.format {
			case output.YAML:
				actual, err := violationsOutput.MarshalYaml()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := violationsOutput.MarshalJson()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := violationsOutput.MarshalHumanReadable()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}

func TestViolationFormat(t *testing.T) {
	//violationsOutput := &ViolationsOutput{Name: "Violations Test Output"}
	v1 := &Violation{
		Timestamp:  "a",
		Name:       "b",
		Kind:       "c",
		Namespace:  "d",
		Constraint: "e",
		Policy:     "f",
		Message:    "g",
		Action:     "h",
	}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "name: b\nkind: c\nnamespace: d\npolicy: f\nconstraint: e\nmessage: g\naction: h\ntimestamp: a\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\n  \"Name\": \"b\",\n  \"Kind\": \"c\",\n  \"Namespace\": \"d\",\n  \"Policy\": \"f\",\n  \"Constraint\": \"e\",\n  \"Message\": \"g\",\n  \"Action\": \"h\",\n  \"Timestamp\": \"a\"\n}",
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "Resource: b\nKind: c\nNamespace: d\nPolicy: f\nConstraint: e\nMessage: g\nAction: h\nTimestamp:\na\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.format {
			case output.YAML:
				actual, err := v1.MarshalYaml()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.JSON:
				actual, err := v1.MarshalJson()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			case output.TEXT:
				actual, err := v1.MarshalHumanReadable()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(actual))
			}
		})
	}
}
