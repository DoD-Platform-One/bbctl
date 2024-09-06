package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type K3dOutput struct {
	Data Output
}

func (o *K3dOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(o.Data)
}

func (o *K3dOutput) MarshalJson() ([]byte, error) {
	return json.MarshalIndent(o.Data, "", "  ")
}

func (o *K3dOutput) MarshalHumanReadable() ([]byte, error) {
	return []byte(o.String()), nil
}

func (o *K3dOutput) String() string {
	var sb strings.Builder

	sb.WriteString("Actions:\n")
	for _, action := range o.Data.Actions {
		sb.WriteString(fmt.Sprintf("  %s\n", action))
	}

	if len(o.Data.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, warning := range o.Data.Warnings {
			sb.WriteString(fmt.Sprintf("  %s\n", warning))
		}
	}

	return sb.String()
}

type HostsOutput struct {
	Hosts map[string][]string `json:"hosts" yaml:"hosts"`
}

func (o *HostsOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(o)
}

func (o *HostsOutput) MarshalJson() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

func (o *HostsOutput) MarshalHumanReadable() ([]byte, error) {
	var sb strings.Builder
	for ip, hostnames := range o.Hosts {
		sb.WriteString(fmt.Sprintf("%s\t%s\n", ip, strings.Join(hostnames, "\t")))
	}
	return []byte(sb.String()), nil
}

type ShellProfileOutput struct {
	KubeConfig       string `json:"kubeconfig"       yaml:"kubeconfig"`
	BB_K3D_PUBLICIP  string `json:"bb_k3d_publicip"  yaml:"bb_k3d_publicip"`
	BB_K3D_PRIVATEIP string `json:"bb_k3d_privateip" yaml:"bb_k3d_privateip"`
}

func (o *ShellProfileOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(o)
}

func (o *ShellProfileOutput) MarshalJson() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

func (o *ShellProfileOutput) MarshalHumanReadable() ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export KUBECONFIG=%s\n", o.KubeConfig))
	sb.WriteString(fmt.Sprintf("export BB_K3D_PUBLICIP=%s\n", o.BB_K3D_PUBLICIP))
	sb.WriteString(fmt.Sprintf("export BB_K3D_PRIVATEIP=%s\n", o.BB_K3D_PRIVATEIP))
	return []byte(sb.String()), nil
}
