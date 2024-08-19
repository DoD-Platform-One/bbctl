package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type CheckStepOutput struct {
	Name   string
	Output []string
	Status string
}

// Outputable interface implementations
func (cso *CheckStepOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(cso)
}

func (cso *CheckStepOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(cso)
}

func (cso *CheckStepOutput) MarshalHumanReadable() ([]byte, error) {
	return []byte(cso.String()), nil
}

func (cso *CheckStepOutput) String() string {
	return fmt.Sprintf("\n\nName: %s\nOutput:\n    %s\nStatus: %s\n", cso.Name, strings.Join(cso.Output[:], "\n    "), cso.Status)
}

type PreflightCheckOutput struct {
	Name  string
	Steps []CheckStepOutput
}

// Outputable interface implementations
func (pco *PreflightCheckOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(pco)
}

func (pco *PreflightCheckOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(pco)
}

func (pco *PreflightCheckOutput) MarshalHumanReadable() ([]byte, error) {
	steps := []string{}
	for _, step := range pco.Steps {
		steps = append(steps, step.String())
	}
	return []byte(fmt.Sprintf("\n\n%s\n\nSteps: %v\n", pco.Name, steps)), nil
}
