package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates a new namespace
func CreateNamespace(k8sinterface kubernetes.Interface, namespace string) (*corev1.Namespace, error) {

	ns := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}

	return k8sinterface.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
}

// DeleteNamespace deletes a namespace
func DeleteNamespace(k8sinterface kubernetes.Interface, namespace string) error {
	return k8sinterface.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
}
