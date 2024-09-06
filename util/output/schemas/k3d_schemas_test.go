package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	output "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

func TestK3dOutputFormat(t *testing.T) {
	k3dOutput := K3dOutput{
		Data: Output{
			Actions:  []string{"Action 1", "Action 2"},
			Warnings: []string{"Warning 1"},
		},
	}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "general_info: {}\nactions:\n- Action 1\n- Action 2\nwarnings:\n- Warning 1\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\n  \"general_info\": null,\n  \"actions\": [\n    \"Action 1\",\n    \"Action 2\"\n  ],\n  \"warnings\": [\n    \"Warning 1\"\n  ]\n}",
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "Actions:\n  Action 1\n  Action 2\n\nWarnings:\n  Warning 1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual []byte
			var err error

			switch tt.format {
			case output.YAML:
				actual, err = k3dOutput.MarshalYaml()
			case output.JSON:
				actual, err = k3dOutput.MarshalJson()
			case output.TEXT:
				actual, err = k3dOutput.MarshalHumanReadable()
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestHostsOutputFormat(t *testing.T) {
	hostsOutput := HostsOutput{
		Hosts: map[string][]string{
			"192.168.1.1": {"host1.local", "host2.local"},
		},
	}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "hosts:\n  192.168.1.1:\n  - host1.local\n  - host2.local\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\n  \"hosts\": {\n    \"192.168.1.1\": [\n      \"host1.local\",\n      \"host2.local\"\n    ]\n  }\n}",
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "192.168.1.1\thost1.local\thost2.local\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual []byte
			var err error

			switch tt.format {
			case output.YAML:
				actual, err = hostsOutput.MarshalYaml()
			case output.JSON:
				actual, err = hostsOutput.MarshalJson()
			case output.TEXT:
				actual, err = hostsOutput.MarshalHumanReadable()
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestShellProfileOutputFormat(t *testing.T) {
	shellProfileOutput := ShellProfileOutput{
		KubeConfig:       "~/.kube/developer-dev-config",
		BB_K3D_PUBLICIP:  "172.16.1.1",
		BB_K3D_PRIVATEIP: "10.0.0.1",
	}
	tests := []struct {
		name     string
		format   output.OutputFormat
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "kubeconfig: ~/.kube/developer-dev-config\nbb_k3d_publicip: 172.16.1.1\nbb_k3d_privateip: 10.0.0.1\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\n  \"kubeconfig\": \"~/.kube/developer-dev-config\",\n  \"bb_k3d_publicip\": \"172.16.1.1\",\n  \"bb_k3d_privateip\": \"10.0.0.1\"\n}",
		},
		{
			name:     "Text Output",
			format:   output.TEXT,
			expected: "export KUBECONFIG=~/.kube/developer-dev-config\nexport BB_K3D_PUBLICIP=172.16.1.1\nexport BB_K3D_PRIVATEIP=10.0.0.1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual []byte
			var err error

			switch tt.format {
			case output.YAML:
				actual, err = shellProfileOutput.MarshalYaml()
			case output.JSON:
				actual, err = shellProfileOutput.MarshalJson()
			case output.TEXT:
				actual, err = shellProfileOutput.MarshalHumanReadable()
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}
