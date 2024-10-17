package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyListOutput_Marshal(t *testing.T) {
	var testObject = PolicyListOutput{
		Messages: []string{"Test message 1", "Test message 2"},
		CrdPolicies: []CrdPolicyOutput{
			{
				CrdName: "restricted-test-toleration",
				Policies: []PolicyOutput{
					{
						Name:        "restricted-test-toleration",
						Namespace:   "test",
						Kind:        "RestrictedTest",
						Description: "Test policy description",
						Action:      "deny",
					},
					{
						Name:        "Name-test",
						Kind:        "Kindtest",
						Description: "Policy test description",
						Action:      "deny",
					},
				},
				Message: "No constraints found for CRD",
			},
			{
				CrdName:  "test-crd",
				Policies: []PolicyOutput{},
			},
		},
	}

	tests := []struct {
		name     string
		marshal  func() ([]byte, error)
		expected string
	}{
		{
			name:    "YAML",
			marshal: testObject.EncodeYAML,
			expected: `messages:
  - Test message 1
  - Test message 2
crdPolicies:
  - crdName: restricted-test-toleration
    policies:
      - name: restricted-test-toleration
        namespace: test
        kind: RestrictedTest
        description: Test policy description
        action: deny
      - name: Name-test
        namespace: ""
        kind: Kindtest
        description: Policy test description
        action: deny
    message: No constraints found for CRD
  - crdName: test-crd
    policies: []
    message: ""
`,
		},
		{
			name:     "JSON",
			marshal:  testObject.EncodeJSON,
			expected: "{\"messages\":[\"Test message 1\",\"Test message 2\"],\"crdPolicies\":[{\"crdName\":\"restricted-test-toleration\",\"policies\":[{\"name\":\"restricted-test-toleration\",\"namespace\":\"test\",\"kind\":\"RestrictedTest\",\"description\":\"Test policy description\",\"action\":\"deny\"},{\"name\":\"Name-test\",\"namespace\":\"\",\"kind\":\"Kindtest\",\"description\":\"Policy test description\",\"action\":\"deny\"}],\"message\":\"No constraints found for CRD\"},{\"crdName\":\"test-crd\",\"policies\":[],\"message\":\"\"}]}",
		},
		{
			name:     "Text",
			marshal:  testObject.EncodeText,
			expected: "\nTest message 1\n\nTest message 2\n\n\nrestricted-test-toleration\n\nKind: RestrictedTest, Name: restricted-test-toleration, Namespace: test, EnforcementAction: deny\n\nTest policy description\n\n\nKind: Kindtest, Name: Name-test, EnforcementAction: deny\n\nPolicy test description\n\n\n\n\n\ntest-crd\n\n\n\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.marshal()
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}
