package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type CheckStatusOutput struct {
	Name   string   `json:"name"   yaml:"name"`
	Output []string `json:"output" yaml:"output"`
}

// Outputable interface implementations
func (cso *CheckStatusOutput) EncodeYAML() ([]byte, error) {
	return yaml.Marshal(cso)
}

func (cso *CheckStatusOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(cso)
}

func (cso *CheckStatusOutput) EncodeText() ([]byte, error) {
	return []byte(cso.String()), nil
}

func (cso *CheckStatusOutput) String() string {
	return fmt.Sprintf("\n\nName: %s\nOutput:\n    %s\n", cso.Name, strings.Join(cso.Output, "\n    "))
}

type StatusOutput struct {
	Name     string              `json:"name"     yaml:"name"`
	Statuses []CheckStatusOutput `json:"statuses" yaml:"statuses"`
}

// Outputable interface implementations
func (so *StatusOutput) EncodeYAML() ([]byte, error) {
	return yaml.Marshal(so)
}

func (so *StatusOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(so)
}

func (so *StatusOutput) EncodeText() ([]byte, error) {
	statuses := []string{}
	for _, status := range so.Statuses {
		statuses = append(statuses, status.String())
	}
	return []byte(fmt.Sprintf("\n\n%s\n\nStatuses: %v\n", so.Name, statuses)), nil
}
