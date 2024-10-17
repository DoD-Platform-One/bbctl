package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/yamler"
)

type PolicyOutput struct {
	Name        string `json:"name"        yaml:"name"`
	Namespace   string `json:"namespace"   yaml:"namespace"`
	Kind        string `json:"kind"        yaml:"kind"`
	Description string `json:"description" yaml:"description"`
	Action      string `json:"action"      yaml:"action"`
}

type CrdPolicyOutput struct {
	CrdName  string         `json:"crdName"  yaml:"crdName"`
	Policies []PolicyOutput `json:"policies" yaml:"policies"`
	Message  string         `json:"message"  yaml:"message"`
}

type PolicyListOutput struct {
	Messages    []string          `json:"messages"    yaml:"messages"`
	CrdPolicies []CrdPolicyOutput `json:"crdPolicies" yaml:"crdPolicies"`
}

func (plo *PolicyListOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(plo)
}

func (plo *PolicyListOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(plo)
}

func (plo *PolicyListOutput) EncodeText() ([]byte, error) {
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
			output.WriteString("\n" + crdPolicy.Message)
		}
		output.WriteString("\n\n\n")
	}
	return []byte(output.String()), nil
}
