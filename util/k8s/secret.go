package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DockerConfigEntry holds user auth information that grants access to a docker registry
type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty" datapolicy:"password"`
	Auth     string `json:"auth,omitempty" datapolicy:"token"`
}

// DockerConfig represents the config file used by the docker CLI.
// This config is mapping of reqistry server URIs to the credentials that can be used to pull images from them
type DockerConfig map[string]DockerConfigEntry

// DockerConfigJSON represents a local docker auth config file
// for pulling images.
type DockerConfigJSON struct {
	Authorizations DockerConfig `json:"auths" datapolicy:"token"`
	// +optional
	HTTPHeaders map[string]string `json:"HttpHeaders,omitempty" datapolicy:"token"`
}

// CreateRegistrySecret creates a new secret for docker registry credentials and deploys it into a k8s cluster using the given parameters
//
// Returns a nil secret and an error if there were any issues creating the secret
func CreateRegistrySecret(k8sInterface kubernetes.Interface, namespace string, name string, server string, username string, password string) (*coreV1.Secret, error) {
	return createRegistrySecret(k8sInterface, namespace, name, server, username, password, json.Marshal)
}

// Internal helper function to implement CreateRegistrySecret
//
// Returns a nil secret and an error if there were any issues creating the secret
func createRegistrySecret(k8sInterface kubernetes.Interface, namespace string, name string, server string, username string, password string, jsonMarshalFunction func(any) ([]byte, error)) (*coreV1.Secret, error) {
	secret := &coreV1.Secret{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: coreV1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: coreV1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{},
	}

	dockerConfigAuth := DockerConfigEntry{
		Username: username,
		Password: password,
		Auth:     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
	}

	dockerConfigJSON := DockerConfigJSON{
		Authorizations: map[string]DockerConfigEntry{server: dockerConfigAuth},
	}

	bytes, err := jsonMarshalFunction(dockerConfigJSON)
	if err != nil {
		return nil, err
	}

	secret.Data[coreV1.DockerConfigJsonKey] = bytes

	return k8sInterface.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metaV1.CreateOptions{})
}

// DeleteRegistrySecret deletes docker registry secret
//
// Returns an error if there were any issues deleting the secret
func DeleteRegistrySecret(k8sInterface kubernetes.Interface, namespace string, name string) error {
	return k8sInterface.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metaV1.DeleteOptions{})
}
