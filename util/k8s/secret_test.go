package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateRegistrySecret(t *testing.T) {

	objs := []runtime.Object{}
	cs := fake.NewSimpleClientset(objs...)

	CreateRegistrySecret(cs, "ns1", "foo", "bar.com", "user", "pass")

	secret, _ := cs.CoreV1().Secrets("ns1").Get(context.TODO(), "foo", meta_v1.GetOptions{})

	if secret.Name != "foo" {
		t.Errorf("unexpected output: %s", secret.Name)
	}

	if secret.Type != core_v1.SecretTypeDockerConfigJson {
		t.Errorf("unexpected output: %s", secret.Type)
	}

	var dockerConfig DockerConfigJSON

	err := json.Unmarshal(secret.Data[core_v1.DockerConfigJsonKey], &dockerConfig)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	dockerConfigEntry := dockerConfig.Auths["bar.com"]

	if dockerConfigEntry.Username != "user" {
		t.Errorf("unexpected output: %s", dockerConfigEntry.Username)
	}

	if dockerConfigEntry.Password != "pass" {
		t.Errorf("unexpected output: %s", dockerConfigEntry.Password)
	}

	auth := base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if dockerConfigEntry.Auth != auth {
		t.Errorf("unexpected output: %s", dockerConfigEntry.Auth)
	}
}

func TestDeleteRegistrySecret(t *testing.T) {

	objs := []runtime.Object{}
	cs := fake.NewSimpleClientset(objs...)

	CreateRegistrySecret(cs, "ns1", "foo", "https://bar.com", "user", "pass")
	DeleteRegistrySecret(cs, "ns1", "foo")

	secret, err := cs.CoreV1().Secrets("ns1").Get(context.TODO(), "foo", meta_v1.GetOptions{})

	if err == nil {
		t.Errorf("unexpected output: %v", secret)
	}

}
