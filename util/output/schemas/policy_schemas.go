package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type PolicyOutput struct {
	Name        string
	Namespace   string
	Kind        string
	Description string
	Action      string
}

type CrdPolicyOutput struct {
	CrdName  string
	Policies []PolicyOutput
	Message  string
}

type PolicyListOutput struct {
	Messages    []string
	CrdPolicies []CrdPolicyOutput
}

func (plo *PolicyListOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(plo)
}

func (plo *PolicyListOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(plo)
}

func (plo *PolicyListOutput) MarshalHumanReadable() ([]byte, error) {
	var output strings.Builder

	for _, message := range plo.Messages {
		output.WriteString(fmt.Sprintf("\n%s\n", message))
	}
	output.WriteString("\n")

	for _, crdPolicy := range plo.CrdPolicies {
		output.WriteString(fmt.Sprintf("\n%s\n", crdPolicy.CrdName))

		if len(crdPolicy.Policies) > 0 {
			for _, policy := range crdPolicy.Policies {
				if policy.Namespace != "" {
					output.WriteString(fmt.Sprintf("\nKind: %s, Name: %s, Namespace: %s, EnforcementAction: %s\n", policy.Kind, policy.Name, policy.Namespace, policy.Action))
				} else {
					output.WriteString(fmt.Sprintf("\nKind: %s, Name: %s, EnforcementAction: %s\n", policy.Kind, policy.Name, policy.Action))
				}
				if policy.Description != "" {
					output.WriteString(fmt.Sprintf("\n%s\n\n", policy.Description))
				}
			}
		} else {
			output.WriteString(fmt.Sprintf("\n%s", crdPolicy.Message))
		}
		output.WriteString("\n\n\n")
	}
	return []byte(output.String()), nil
}
