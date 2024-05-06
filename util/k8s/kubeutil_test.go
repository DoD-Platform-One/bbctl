package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtils "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestBuildKubeConfig(t *testing.T) {
	factory := bbTestUtils.GetFakeFactory()
	v := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("kubeconfig", "../test/data/kube-config-a.yaml")
	configClient, err := factory.GetConfigClient(nil)
	assert.Nil(t, err)
	config := configClient.GetConfig()

	client, err := BuildKubeConfig(config)
	assert.Nil(t, err)
	assert.Equal(t, "https://test2.com:6443", client.Host)

	v.Set("kubeconfig", "../test/data/kube-config.yaml")
	config = configClient.GetConfig()
	client, err = BuildKubeConfig(config)
	assert.Nil(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)
}

func TestBuildDynamicClient(t *testing.T) {
	factory := bbTestUtils.GetFakeFactory()
	v := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("kubeconfig", "../test/data/kube-config.yaml")
	configClient, err := factory.GetConfigClient(nil)
	assert.Nil(t, err)
	config := configClient.GetConfig()
	client, err := BuildDynamicClient(config)
	assert.Nil(t, err)
	assert.NotNil(t, client)
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
