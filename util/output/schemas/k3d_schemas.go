package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/yamler"
)

type K3dOutput struct {
	Data Output `json:"data" yaml:"data"`
}

func (o *K3dOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(o.Data)
}

func (o *K3dOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(o.Data)
}

func (o *K3dOutput) EncodeText() ([]byte, error) {
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

func (o *HostsOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(o)
}

func (o *HostsOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(o)
}

func (o *HostsOutput) EncodeText() ([]byte, error) {
	var sb strings.Builder
	for ip, hostnames := range o.Hosts {
		sb.WriteString(fmt.Sprintf("%s\t%s\n", ip, strings.Join(hostnames, "\t")))
	}
	return []byte(sb.String()), nil
}

type ShellProfileOutput struct {
	KubeConfig   string `json:"kubeconfig" yaml:"kubeconfig"`
	K3DPublicIP  string `json:"publicIp"   yaml:"publicIp"`
	K3DPrivateIP string `json:"privateIp"  yaml:"privateIp"`
}

func (o *ShellProfileOutput) EncodeYAML() ([]byte, error) {
	return yamler.Marshal(o)
}

func (o *ShellProfileOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(o)
}

func (o *ShellProfileOutput) EncodeText() ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export KUBECONFIG=%s\n", o.KubeConfig))
	sb.WriteString(fmt.Sprintf("export BB_K3D_PUBLICIP=%s\n", o.K3DPublicIP))
	sb.WriteString(fmt.Sprintf("export BB_K3D_PRIVATEIP=%s\n", o.K3DPrivateIP))
	return []byte(sb.String()), nil
}
