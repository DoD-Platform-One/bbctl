package k8s

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtils "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestBuildKubeConfig(t *testing.T) {
	tests := []struct {
		name        string
		homedir     bool
		kubeconfigs [][]string
		shouldErr   bool
		errorString string
	}{
		{
			name:    "valid kubeconfig",
			homedir: false,
			kubeconfigs: [][]string{
				{
					"../test/data/kube-config.yaml",
					"https://test.com:6443",
				},
				{
					"../test/data/kube-config-a.yaml",
					"https://test2.com:6443",
				},
			},
			shouldErr: false,
		},
		{
			name:    "valid kubeconfig with HOME",
			homedir: true,
			kubeconfigs: [][]string{
				{
					"../test/data/kube-config.yaml",
					"https://test.com:6443",
				},
				{
					"../test/data/kube-config-a.yaml",
					"https://test2.com:6443",
				},
			},
			shouldErr: false,
		},
		{
			name:    "invalid kubeconfig",
			homedir: true,
			kubeconfigs: [][]string{
				{
					"../test/data/bad-kube-config.yaml",
					"",
				},
			},
			shouldErr:   true,
			errorString: "cluster has no server defined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			if tt.homedir {
				assert.NoError(t, os.Setenv("HOME", tempDir))
				assert.NoError(t, os.MkdirAll(path.Join(tempDir, ".kube"), 0755))
			}
			factory := bbTestUtils.GetFakeFactory()
			v := factory.GetViper()
			v.Set("big-bang-repo", "test")
			configClient, err := factory.GetConfigClient(nil)
			assert.Nil(t, err)

			for _, kubeconfig := range tt.kubeconfigs {
				if !tt.homedir {
					v.Set("kubeconfig", kubeconfig[0])
				} else {
					v.Set("kubeconfig", "")
					data, err := os.ReadFile(kubeconfig[0])
					assert.Nil(t, err)
					assert.NoError(t, os.WriteFile(path.Join(tempDir, ".kube", "config"), data, 0644))
				}
				config := configClient.GetConfig()

				// Act
				client, err := BuildKubeConfig(config)

				// Assert
				if tt.shouldErr {
					assert.NotNil(t, err)
					assert.Contains(t, err.Error(), tt.errorString)
					assert.Nil(t, client)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, kubeconfig[1], client.Host)
				}
			}
		})
	}
}

func TestBuildDynamicClient(t *testing.T) {
	tests := []struct {
		name        string
		homedir     bool
		kubeconfigs [][]string
		shouldErr   bool
		errorString string
	}{
		{
			name:    "valid kubeconfig",
			homedir: false,
			kubeconfigs: [][]string{
				{
					"../test/data/kube-config.yaml",
					"https://test.com:6443",
				},
				{
					"../test/data/kube-config-a.yaml",
					"https://test2.com:6443",
				},
			},
			shouldErr: false,
		},
		{
			name:    "invalid kubeconfig",
			homedir: true,
			kubeconfigs: [][]string{
				{
					"../test/data/bad-kube-config.yaml",
					"",
				},
			},
			shouldErr:   true,
			errorString: "cluster has no server defined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			if tt.homedir {
				assert.NoError(t, os.Setenv("HOME", tempDir))
				assert.NoError(t, os.MkdirAll(path.Join(tempDir, ".kube"), 0755))
			}
			factory := bbTestUtils.GetFakeFactory()
			v := factory.GetViper()
			v.Set("big-bang-repo", "test")
			for _, kubeconfig := range tt.kubeconfigs {
				if !tt.homedir {
					v.Set("kubeconfig", kubeconfig[0])
				} else {
					v.Set("kubeconfig", "")
					data, err := os.ReadFile(kubeconfig[0])
					assert.Nil(t, err)
					assert.NoError(t, os.WriteFile(path.Join(tempDir, ".kube", "config"), data, 0644))
				}
				configClient, err := factory.GetConfigClient(nil)
				assert.Nil(t, err)
				config := configClient.GetConfig()

				// Act
				client, err := BuildDynamicClient(config)

				// Assert
				if tt.shouldErr {
					assert.NotNil(t, err)
					assert.Contains(t, err.Error(), tt.errorString)
					assert.Nil(t, client)
				} else {
					assert.Nil(t, err)
					assert.NotNil(t, client)
				}
			}
		})
	}
}

func TestGetKubeConfigFromPathList(t *testing.T) {
	configPaths := "../test/data/kube-config.yaml"
	client, err := GetKubeConfigFromPathList(configPaths)
	assert.Nil(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)

	configPaths = "../test/data/kube-config.yaml:no-kube-config.yaml"
	client, err = GetKubeConfigFromPathList(configPaths)
	assert.Nil(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)
}
