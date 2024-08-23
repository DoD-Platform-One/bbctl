package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type CheckStatusOutput struct {
	Name   string
	Output []string
}

// Outputable interface implementations
func (cso *CheckStatusOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(cso)
}

func (cso *CheckStatusOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(cso)
}

func (cso *CheckStatusOutput) MarshalHumanReadable() ([]byte, error) {
	return []byte(cso.String()), nil
}

func (cso *CheckStatusOutput) String() string {
	return fmt.Sprintf("\n\nName: %s\nOutput:\n    %s\n", cso.Name, strings.Join(cso.Output[:], "\n    "))
}

type StatusOutput struct {
	Name     string
	Statuses []CheckStatusOutput
}

// Outputable interface implementations
func (so *StatusOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(so)
}

func (so *StatusOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(so)
}

func (so *StatusOutput) MarshalHumanReadable() ([]byte, error) {
	statuses := []string{}
	for _, status := range so.Statuses {
		statuses = append(statuses, status.String())
	}
	return []byte(fmt.Sprintf("\n\n%s\n\nStatuses: %v\n", so.Name, statuses)), nil
}
