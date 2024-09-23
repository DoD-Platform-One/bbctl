package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestPolicyListOutput_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		marshal  func() ([]byte, error)
		expected string
	}{
		{
			name: "YAML",
			marshal: func() ([]byte, error) {
				return testObject.MarshalYaml()
			},
			expected: "messages:\n- Test message 1\n- Test message 2\ncrdpolicies:\n- crdname: restricted-test-toleration\n  policies:\n  - name: restricted-test-toleration\n    namespace: test\n    kind: RestrictedTest\n    description: Test policy description\n    action: deny\n  - name: Name-test\n    namespace: \"\"\n    kind: Kindtest\n    description: Policy test description\n    action: deny\n  message: No constraints found for CRD\n- crdname: test-crd\n  policies: []\n  message: \"\"\n",
		},
		{
			name: "JSON",
			marshal: func() ([]byte, error) {
				return testObject.MarshalJson()
			},
			expected: "{\"Messages\":[\"Test message 1\",\"Test message 2\"],\"CrdPolicies\":[{\"CrdName\":\"restricted-test-toleration\",\"Policies\":[{\"Name\":\"restricted-test-toleration\",\"Namespace\":\"test\",\"Kind\":\"RestrictedTest\",\"Description\":\"Test policy description\",\"Action\":\"deny\"},{\"Name\":\"Name-test\",\"Namespace\":\"\",\"Kind\":\"Kindtest\",\"Description\":\"Policy test description\",\"Action\":\"deny\"}],\"Message\":\"No constraints found for CRD\"},{\"CrdName\":\"test-crd\",\"Policies\":[],\"Message\":\"\"}]}",
		},
		{
			name: "HumanReadable",
			marshal: func() ([]byte, error) {
				return testObject.MarshalHumanReadable()
			},
			expected: "\nTest message 1\n\nTest message 2\n\n\nrestricted-test-toleration\n\nKind: RestrictedTest, Name: restricted-test-toleration, Namespace: test, EnforcementAction: deny\n\nTest policy description\n\n\nKind: Kindtest, Name: Name-test, EnforcementAction: deny\n\nPolicy test description\n\n\n\n\n\ntest-crd\n\n\n\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.marshal()
			assert.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}
