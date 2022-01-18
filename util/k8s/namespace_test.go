package k8s

import (
	"context"
	"testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {

	objs := []runtime.Object{}
	cs := fake.NewSimpleClientset(objs...)

	CreateNamespace(cs, "ns1")

	ns, _ := cs.CoreV1().Namespaces().Get(context.TODO(), "ns1", meta_v1.GetOptions{})

	if ns.Name != "ns1" {
		t.Errorf("unexpected output: %s", ns.Name)
	}
}

func TestDeleteNamespace(t *testing.T) {

	objs := []runtime.Object{}
	cs := fake.NewSimpleClientset(objs...)

	CreateNamespace(cs, "ns1")
	DeleteNamespace(cs, "ns1")

	ns, err := cs.CoreV1().Namespaces().Get(context.TODO(), "ns1", meta_v1.GetOptions{})

	if err == nil {
		t.Errorf("unexpected output: %v", ns)
	}

}
