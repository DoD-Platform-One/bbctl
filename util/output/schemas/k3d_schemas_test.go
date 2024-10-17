package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		format   output.Format
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "generalInfo: {}\nactions:\n- Action 1\n- Action 2\nwarnings:\n- Warning 1\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\"generalInfo\":null,\"actions\":[\"Action 1\",\"Action 2\"],\"warnings\":[\"Warning 1\"]}",
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
				actual, err = k3dOutput.EncodeYAML()
			case output.JSON:
				actual, err = k3dOutput.EncodeJSON()
			case output.TEXT:
				actual, err = k3dOutput.EncodeText()
			}

			require.NoError(t, err)
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
		format   output.Format
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
			expected: "{\"hosts\":{\"192.168.1.1\":[\"host1.local\",\"host2.local\"]}}",
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
				actual, err = hostsOutput.EncodeYAML()
			case output.JSON:
				actual, err = hostsOutput.EncodeJSON()
			case output.TEXT:
				actual, err = hostsOutput.EncodeText()
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestShellProfileOutputFormat(t *testing.T) {
	shellProfileOutput := ShellProfileOutput{
		KubeConfig:   "~/.kube/developer-dev-config",
		K3DPublicIP:  "172.16.1.1",
		K3DPrivateIP: "10.0.0.1",
	}
	tests := []struct {
		name     string
		format   output.Format
		expected string
	}{
		{
			name:     "YAML Output",
			format:   output.YAML,
			expected: "kubeconfig: ~/.kube/developer-dev-config\npublicIp: 172.16.1.1\nprivateIp: 10.0.0.1\n",
		},
		{
			name:     "JSON Output",
			format:   output.JSON,
			expected: "{\"kubeconfig\":\"~/.kube/developer-dev-config\",\"publicIp\":\"172.16.1.1\",\"privateIp\":\"10.0.0.1\"}",
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
				actual, err = shellProfileOutput.EncodeYAML()
			case output.JSON:
				actual, err = shellProfileOutput.EncodeJSON()
			case output.TEXT:
				actual, err = shellProfileOutput.EncodeText()
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}
