package schemas

import (
	"encoding/json"

	"github.com/gosuri/uitable"
	"gopkg.in/yaml.v2"
)

type HelmReleaseOutput struct {
	Name       string `json:"name"       yaml:"name"`
	Namespace  string `json:"namespace"  yaml:"namespace"`
	Revision   int    `json:"revision"   yaml:"revision"`
	Status     string `json:"status"     yaml:"status"`
	Chart      string `json:"chart"      yaml:"chart"`
	AppVersion string `json:"appVersion" yaml:"appVersion"`
}

type HelmReleaseTableOutput struct {
	Releases []HelmReleaseOutput `json:"releases" yaml:"releases"`
}

func (hrto *HelmReleaseTableOutput) EncodeYAML() ([]byte, error) {
	return yaml.Marshal(hrto)
}

func (hrto *HelmReleaseTableOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(hrto)
}

func (hrto *HelmReleaseTableOutput) EncodeText() ([]byte, error) {
	table := uitable.New()
	table.AddRow("NAME", "NAMESPACE", "REVISION", "STATUS", "CHART", "APPVERSION")

	for _, r := range hrto.Releases {
		table.AddRow(
			r.Name,
			r.Namespace,
			r.Revision,
			r.Status,
			r.Chart,
			r.AppVersion,
		)
	}
	return []byte(table.String()), nil
}
