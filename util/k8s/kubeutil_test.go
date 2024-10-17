package k8s

import (
	"bytes"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bbUtilConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbUtilLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
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
				t.Setenv("HOME", tempDir)
				require.NoError(t, os.MkdirAll(path.Join(tempDir, ".kube"), 0755))
			}
			cmd := &cobra.Command{}
			v := viper.New()
			v.Set("big-bang-repo", "test")
			stream := &bytes.Buffer{}
			logger := slog.New(slog.NewJSONHandler(stream, &slog.HandlerOptions{}))
			loggingClientGetter := bbUtilLog.ClientGetter{}
			loggingClient := loggingClientGetter.GetClient(logger)
			configClientGetter := bbUtilConfig.ClientGetter{}
			configClient, _ := configClientGetter.GetClient(cmd, &loggingClient, v)

			for _, kubeconfig := range tt.kubeconfigs {
				if !tt.homedir {
					v.Set("kubeconfig", kubeconfig[0])
				} else {
					v.Set("kubeconfig", "")
					data, err := os.ReadFile(kubeconfig[0])
					require.NoError(t, err)
					require.NoError(t, os.WriteFile(path.Join(tempDir, ".kube", "config"), data, 0600))
				}
				config, _ := configClient.GetConfig()

				// Act
				client, err := BuildKubeConfig(config)

				// Assert
				if tt.shouldErr {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorString)
					assert.Nil(t, client)
				} else {
					require.NoError(t, err)
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
				t.Setenv("HOME", tempDir)
				require.NoError(t, os.MkdirAll(path.Join(tempDir, ".kube"), 0755))
			}
			v := viper.New()
			v.Set("big-bang-repo", "test")
			for _, kubeconfig := range tt.kubeconfigs {
				if !tt.homedir {
					v.Set("kubeconfig", kubeconfig[0])
				} else {
					v.Set("kubeconfig", "")
					data, err := os.ReadFile(kubeconfig[0])
					require.NoError(t, err)
					require.NoError(t, os.WriteFile(path.Join(tempDir, ".kube", "config"), data, 0600))
				}
				cmd := &cobra.Command{}
				stream := &bytes.Buffer{}
				logger := slog.New(slog.NewJSONHandler(stream, &slog.HandlerOptions{}))
				loggingClientGetter := bbUtilLog.ClientGetter{}
				loggingClient := loggingClientGetter.GetClient(logger)
				configClientGetter := bbUtilConfig.ClientGetter{}
				configClient, _ := configClientGetter.GetClient(cmd, &loggingClient, v)
				config, _ := configClient.GetConfig()

				// Act
				client, err := BuildDynamicClient(config)

				// Assert
				if tt.shouldErr {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorString)
					assert.Nil(t, client)
				} else {
					require.NoError(t, err)
					assert.NotNil(t, client)
				}
			}
		})
	}
}

func TestGetKubeConfigFromPathList(t *testing.T) {
	configPaths := "../test/data/kube-config.yaml"
	client, err := GetKubeConfigFromPathList(configPaths)
	require.NoError(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)

	configPaths = "../test/data/kube-config.yaml:no-kube-config.yaml"
	client, err = GetKubeConfigFromPathList(configPaths)
	require.NoError(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)
}
