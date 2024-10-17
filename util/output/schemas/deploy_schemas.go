package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type HelmOutput struct {
	Message      string `json:"message"      yaml:"message"`
	Name         string `json:"name"         yaml:"name"`
	LastDeployed string `json:"lastDeployed" yaml:"lastDeployed"`
	Namespace    string `json:"namespace"    yaml:"namespace"`
	Status       string `json:"status"       yaml:"status"`
	Revision     string `json:"revision"     yaml:"revision"`
	TestSuite    string `json:"testSuite"    yaml:"testSuite"`
	Notes        string `json:"notes"        yaml:"notes"`
}

type BigbangOutput struct {
	Data HelmOutput
}

func (o *BigbangOutput) EncodeYAML() ([]byte, error) {
	return yaml.Marshal(o.Data)
}

func (o *BigbangOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(o.Data)
}

func (o *BigbangOutput) EncodeText() ([]byte, error) {
	return []byte(o.String()), nil
}

func (o *BigbangOutput) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Message: %s\n", o.Data.Message))
	sb.WriteString(fmt.Sprintf("Name: %s\n", o.Data.Name))
	sb.WriteString(fmt.Sprintf("Last Deployed: %s\n", o.Data.LastDeployed))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", o.Data.Namespace))
	sb.WriteString(fmt.Sprintf("Status: %s\n", o.Data.Status))
	sb.WriteString(fmt.Sprintf("Revision: %s\n", o.Data.Revision))
	sb.WriteString(fmt.Sprintf("Test Suite: %s\n", o.Data.TestSuite))
	sb.WriteString(fmt.Sprintf("Notes:\n%s\n", o.Data.Notes))
	return sb.String()
}

type Output struct {
	GeneralInfo map[string]string `json:"generalInfo" yaml:"generalInfo"`
	Actions     []string          `json:"actions"     yaml:"actions"`
	Warnings    []string          `json:"warnings"    yaml:"warnings"`
}

type FluxOutput struct {
	Data Output
}

func (fo *FluxOutput) EncodeYAML() ([]byte, error) {
	return yaml.Marshal(fo.Data)
}

func (fo *FluxOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(fo.Data)
}

func (fo *FluxOutput) EncodeText() ([]byte, error) {
	return []byte(fo.String()), nil
}

func (fo *FluxOutput) String() string {
	var sb strings.Builder

	sb.WriteString("General Info:\n")
	for key, value := range fo.Data.GeneralInfo {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
	}

	sb.WriteString("\nActions:\n")
	for _, action := range fo.Data.Actions {
		sb.WriteString(fmt.Sprintf("  %s\n", action))
	}

	if len(fo.Data.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, warning := range fo.Data.Warnings {
			sb.WriteString(fmt.Sprintf("  %s\n", warning))
		}
	}

	return sb.String()
}
