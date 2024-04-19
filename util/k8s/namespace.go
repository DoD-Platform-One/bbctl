package k8s

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates a new namespace
func CreateNamespace(k8sInterface kubernetes.Interface, namespace string) (*coreV1.Namespace, error) {
	ns := &coreV1.Namespace{
		TypeMeta:   metaV1.TypeMeta{APIVersion: coreV1.SchemeGroupVersion.String(), Kind: "Namespace"},
		ObjectMeta: metaV1.ObjectMeta{Name: namespace},
	}

	return k8sInterface.CoreV1().Namespaces().Create(context.TODO(), ns, metaV1.CreateOptions{})
}

// DeleteNamespace deletes a namespace
func DeleteNamespace(k8sInterface kubernetes.Interface, namespace string) error {
	return k8sInterface.CoreV1().Namespaces().Delete(context.TODO(), namespace, metaV1.DeleteOptions{})
}
