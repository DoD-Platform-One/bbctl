package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateRegistrySecret(t *testing.T) {
	objects := []runtime.Object{}
	cs := fake.NewSimpleClientset(objects...)

	_, err := CreateRegistrySecret(cs, "ns1", "foo", "bar.com", "user", "pass")
	assert.Nil(t, err)

	secret, err := cs.CoreV1().Secrets("ns1").Get(context.TODO(), "foo", metaV1.GetOptions{})
	assert.Nil(t, err)

	if secret.Name != "foo" {
		t.Errorf("unexpected output: %s", secret.Name)
	}

	if secret.Type != coreV1.SecretTypeDockerConfigJson {
		t.Errorf("unexpected output: %s", secret.Type)
	}

	var dockerConfig DockerConfigJSON

	err = json.Unmarshal(secret.Data[coreV1.DockerConfigJsonKey], &dockerConfig)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	dockerConfigEntry := dockerConfig.Authorizations["bar.com"]

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
	objects := []runtime.Object{}
	cs := fake.NewSimpleClientset(objects...)

	_, err := CreateRegistrySecret(cs, "ns1", "foo", "https://bar.com", "user", "pass")
	assert.NoError(t, err)
	err = DeleteRegistrySecret(cs, "ns1", "foo")
	assert.NoError(t, err)

	secret, err := cs.CoreV1().Secrets("ns1").Get(context.TODO(), "foo", metaV1.GetOptions{})

	if err == nil {
		t.Errorf("unexpected output: %v", secret)
	}
}
