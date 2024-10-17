package schemas

import (
	"bytes"
	"encoding/json"
	"fmt"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/yamler"
)

type Violation struct {
	Name       string `json:"name"       yaml:"name"`
	Kind       string `json:"kind"       yaml:"kind"`
	Namespace  string `json:"namespace"  yaml:"namespace"`
	Policy     string `json:"policy"     yaml:"policy"`
	Constraint string `json:"constraint" yaml:"constraint"`
	Message    string `json:"message"    yaml:"message"`
	Action     string `json:"action"     yaml:"action"`
	Timestamp  string `json:"timestamp"  yaml:"timestamp"` // UTC time
}

func (v *Violation) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(v)
}

func (v *Violation) EncodeJSON() ([]byte, error) {
	return json.Marshal(v)
}

func (v *Violation) EncodeText() ([]byte, error) {
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
	Name       string      `json:"name"       yaml:"name"`
	Violations []Violation `json:"violations" yaml:"violations"`
}

func (vo *ViolationsOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(vo)
}

func (vo *ViolationsOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(vo)
}

func (vo *ViolationsOutput) EncodeText() ([]byte, error) {
	return []byte(vo.String()), nil
}

func (vo *ViolationsOutput) String() string {
	var sb bytes.Buffer

	sb.WriteString(vo.Name + ":\n")
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
