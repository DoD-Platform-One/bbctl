package schemas

import (
	"encoding/json"

	"github.com/gosuri/uitable"
	"gopkg.in/yaml.v2"
)

type HelmReleaseOutput struct {
	Name       string
	Namespace  string
	Revision   int
	Status     string
	Chart      string
	AppVersion string
}

type HelmReleaseTableOutput struct {
	Releases []HelmReleaseOutput
}

func (hrto *HelmReleaseTableOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(hrto)
}

func (hrto *HelmReleaseTableOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(hrto)
}

func (hrto *HelmReleaseTableOutput) MarshalHumanReadable() ([]byte, error) {
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
