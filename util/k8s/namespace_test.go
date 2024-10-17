package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {
	objects := []runtime.Object{}
	cs := fake.NewSimpleClientset(objects...)

	_, err := CreateNamespace(cs, "ns1")
	require.NoError(t, err)

	ns, _ := cs.CoreV1().Namespaces().Get(context.TODO(), "ns1", metaV1.GetOptions{})

	if ns.Name != "ns1" {
		t.Errorf("unexpected output: %s", ns.Name)
	}
}

func TestDeleteNamespace(t *testing.T) {
	objects := []runtime.Object{}
	cs := fake.NewSimpleClientset(objects...)

	_, err := CreateNamespace(cs, "ns1")
	require.NoError(t, err)
	err = DeleteNamespace(cs, "ns1")
	require.NoError(t, err)

	ns, err := cs.CoreV1().Namespaces().Get(context.TODO(), "ns1", metaV1.GetOptions{})

	if err == nil {
		t.Errorf("unexpected output: %v", ns)
	}
}
