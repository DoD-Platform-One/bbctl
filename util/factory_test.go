package util

import (
	"bytes"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetHelmClient(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "./test/data/kube-config", "")
	factory := NewFactory(flags)
	client, err := factory.GetHelmClient("foo")
	assert.Nil(t, err)
	assert.NotNil(t, client.GetList)
	assert.NotNil(t, client.GetRelease)
	assert.NotNil(t, client.GetValues)
}

func TestGetHelmClientBadConfig(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "no-kube-config", "")
	factory := NewFactory(flags)
	client, err := factory.GetHelmClient("foo")
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetK8sClientset(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "./test/data/kube-config", "")
	factory := NewFactory(flags)
	client, err := factory.GetK8sClientset()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sClientsetBadConfig(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "no-kube-config", "")
	factory := NewFactory(flags)
	client, err := factory.GetK8sClientset()
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetK8sDynamicClient(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "./test/data/kube-config", "")
	factory := NewFactory(flags)
	client, err := factory.GetK8sDynamicClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestGetK8sDynamicClientBadConfig(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "no-kube-config", "")
	factory := NewFactory(flags)
	client, err := factory.GetK8sDynamicClient()
	assert.NotNil(t, err)
	assert.Nil(t, client)
}

func TestGetRestConfig(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "./test/data/kube-config", "")
	factory := NewFactory(flags)
	config, err := factory.GetRestConfig()
	assert.Nil(t, err)
	assert.NotNil(t, config)
}

func TestGetRestConfigBadConfig(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "no-kube-config", "")
	factory := NewFactory(flags)
	config, err := factory.GetRestConfig()
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestGetCommandExecutor(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "./test/data/kube-config", "")
	factory := NewFactory(flags)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	var stdout, stderr bytes.Buffer
	executor, err := factory.GetCommandExecutor(pod, "foo", []string{"hello"}, &stdout, &stderr)
	assert.Nil(t, err)
	assert.NotNil(t, executor)
}

func TestGetCommandExecutorBadConfig(t *testing.T) {
	var flags *pflag.FlagSet = &pflag.FlagSet{}
	flags.String("kubeconfig", "no-kube-config", "")
	factory := NewFactory(flags)
	executor, err := factory.GetCommandExecutor(nil, "", nil, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, executor)
}
