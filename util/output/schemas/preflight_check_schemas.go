package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/yamler"
)

type CheckStepOutput struct {
	Name   string   `json:"name"   yaml:"name"`
	Output []string `json:"output" yaml:"output"`
	Status string   `json:"status" yaml:"status"`
}

// Outputable interface implementations
func (cso *CheckStepOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(cso)
}

func (cso *CheckStepOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(cso)
}

func (cso *CheckStepOutput) EncodeText() ([]byte, error) {
	return []byte(cso.String()), nil
}

func (cso *CheckStepOutput) String() string {
	return fmt.Sprintf("\n\nName: %s\nOutput:\n    %s\nStatus: %s\n", cso.Name, strings.Join(cso.Output, "\n    "), cso.Status)
}

type PreflightCheckOutput struct {
	Name  string            `json:"name"   yaml:"name"`
	Steps []CheckStepOutput `json:"steps"  yaml:"steps"`
}

// Outputable interface implementations
func (pco *PreflightCheckOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(pco)
}

func (pco *PreflightCheckOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(pco)
}

func (pco *PreflightCheckOutput) EncodeText() ([]byte, error) {
	steps := []string{}
	for _, step := range pco.Steps {
		steps = append(steps, step.String())
	}
	return []byte(fmt.Sprintf("\n\n%s\n\nSteps: %v\n", pco.Name, steps)), nil
}
