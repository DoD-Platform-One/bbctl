package util

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetHelmClient(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	client, err := factory.GetHelmClient(nil, "foo")
	assert.Nil(t, err)
	assert.NotNil(t, client.GetList)
	assert.NotNil(t, client.GetRelease)
	assert.NotNil(t, client.GetValues)
}

func TestGetHelmClientBadConfig(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	client, err := factory.GetHelmClient(nil, "foo")
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetK8sClientset(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	client, err := factory.GetK8sClientset(nil)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sClientsetBadConfig(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	client, err := factory.GetK8sClientset(nil)
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetK8sDynamicClient(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	client, err := factory.GetK8sDynamicClient(nil)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sDynamicClientBadConfig(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	client, err := factory.GetK8sDynamicClient(nil)
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetRestConfig(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	config, err := factory.GetRestConfig(nil)
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestGetRestConfigBadConfig(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	config, err := factory.GetRestConfig(nil)
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestGetCommandExecutor(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "./test/data/kube-config.yaml")
	pod := &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	var stdout, stderr bytes.Buffer
	executor, err := factory.GetCommandExecutor(nil, pod, "foo", []string{"hello"}, &stdout, &stderr)
	assert.Nil(t, err)
	assert.NotNil(t, executor)
}

func TestGetCommandExecutorBadConfig(t *testing.T) {
	factory := NewFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "no-kube-config.yaml")
	executor, err := factory.GetCommandExecutor(nil, nil, "", nil, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, executor)
}
