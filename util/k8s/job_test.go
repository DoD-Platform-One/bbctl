package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateJob(t *testing.T) {
	jobDesc := &JobDesc{
		Name:               "foo",
		ContainerName:      "bar",
		ContainerImage:     "docker.io/busybox:1.28",
		ImagePullSecret:    "secret",
		Command:            []string{"sleep", "5"},
		Args:               []string{"nop"},
		TTLSecondsOnFinish: 1,
	}

	objects := []runtime.Object{}
	cs := fake.NewSimpleClientset(objects...)

	_, err := CreateJob(cs, "default", jobDesc)
	require.NoError(t, err)

	job, _ := cs.BatchV1().Jobs("default").Get(context.TODO(), "foo", metaV1.GetOptions{})

	if job.Name != "foo" {
		t.Errorf("unexpected output: %s", job.Name)
	}

	var ttl int32 = 1

	if *job.Spec.TTLSecondsAfterFinished != ttl {
		t.Errorf("unexpected output: %v", *job.Spec.TTLSecondsAfterFinished)
	}

	if job.Spec.Template.Spec.Containers[0].Name != "bar" {
		t.Errorf("unexpected output: %s", job.Spec.Template.Spec.Containers[0].Name)
	}

	if job.Spec.Template.Spec.Containers[0].Image != "docker.io/busybox:1.28" {
		t.Errorf("unexpected output: %s", job.Spec.Template.Spec.Containers[0].Image)
	}

	if job.Spec.Template.Spec.Containers[0].Command[0] != "sleep" {
		t.Errorf("unexpected output: %s", job.Spec.Template.Spec.Containers[0].Command[0])
	}

	if job.Spec.Template.Spec.Containers[0].Args[0] != "nop" {
		t.Errorf("unexpected output: %s", job.Spec.Template.Spec.Containers[0].Args[0])
	}
}
