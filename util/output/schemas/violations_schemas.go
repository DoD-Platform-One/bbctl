package schemas

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Violation struct {
	Name       string // resource name
	Kind       string // resource kind
	Namespace  string // resource namespace
	Policy     string // kyverno policy name
	Constraint string // gatekeeper constraint name
	Message    string // policy violation message
	Action     string // enforcement action
	Timestamp  string // utc time
}

func (v *Violation) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(v)
}

func (v *Violation) MarshalJson() ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (v *Violation) MarshalHumanReadable() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Violation) String() string {
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf("Resource: %s\n", v.Name))
	sb.WriteString(fmt.Sprintf("Kind: %s\n", v.Kind))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", v.Namespace))
	sb.WriteString(fmt.Sprintf("Policy: %s\n", v.Policy))
	sb.WriteString(fmt.Sprintf("Constraint: %s\n", v.Constraint))
	sb.WriteString(fmt.Sprintf("Message: %s\n", v.Message))
	sb.WriteString(fmt.Sprintf("Action: %s\n", v.Action))
	sb.WriteString(fmt.Sprintf("Timestamp:\n%s\n", v.Timestamp))
	return sb.String()
}

type ViolationsOutput struct {
	Name       string
	Violations []Violation
}

func (vo *ViolationsOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(vo)
}

func (vo *ViolationsOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(vo)
}

func (vo *ViolationsOutput) MarshalHumanReadable() ([]byte, error) {
	return []byte(vo.String()), nil
}

func (vo *ViolationsOutput) String() string {
	var sb bytes.Buffer

	sb.WriteString(fmt.Sprintf("%s:\n", vo.Name))
	for _, violation := range vo.Violations {
		sb.WriteString(fmt.Sprintf("  Resource: %s\n", violation.Name))
		sb.WriteString(fmt.Sprintf("  Kind: %s\n", violation.Kind))
		sb.WriteString(fmt.Sprintf("  Namespace: %s\n", violation.Namespace))
		sb.WriteString(fmt.Sprintf("  Policy: %s\n", violation.Policy))
		sb.WriteString(fmt.Sprintf("  Constraint: %s\n", violation.Constraint))
		sb.WriteString(fmt.Sprintf("  Message: %s\n", violation.Message))
		sb.WriteString(fmt.Sprintf("  Action: %s\n", violation.Action))
		sb.WriteString(fmt.Sprintf("  Timestamp: %s\n\n", violation.Timestamp))
	}

	return sb.String()
}
