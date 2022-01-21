package k8s

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBuildKubeConfigFromFlags(t *testing.T) {

	var flags *pflag.FlagSet = &pflag.FlagSet{}
	viper.Set("kubeconfig", "../test/data/kube-config-a")

	client, err := BuildKubeConfigFromFlags(flags)
	assert.Nil(t, err)
	assert.Equal(t, "https://test2.com:6443", client.Host)

	flags.String("kubeconfig", "../test/data/kube-config", "")
	client, err = BuildKubeConfigFromFlags(flags)
	assert.Nil(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)
}

func TestBuildDynamicClientFromFlags(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "../test/data/kube-config", "")
	client, err := BuildDynamicClientFromFlags(flags)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetKubeConfigFromPathList(t *testing.T) {

	configPaths := "../test/data/kube-config"
	client, err := GetKubeConfigFromPathList(configPaths)
	assert.Nil(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)

	configPaths = "../test/data/kube-config:no-kube-config"
	client, err = GetKubeConfigFromPathList(configPaths)
	assert.Nil(t, err)
	assert.Equal(t, "https://test.com:6443", client.Host)
}
