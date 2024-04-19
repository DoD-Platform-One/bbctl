package k8s

import (
	"context"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// JobDesc - Job Descriptor
type JobDesc struct {
	Name               string
	ContainerName      string
	ContainerImage     string
	ImagePullSecret    string
	Command            []string
	Args               []string
	TTLSecondsOnFinish int32
}

// CreateJob function creates a new job
func CreateJob(client kubernetes.Interface, namespace string, jobDesc *JobDesc) (*batchV1.Job, error) {
	job := &batchV1.Job{
		ObjectMeta: metaV1.ObjectMeta{
			Name: jobDesc.Name,
		},
		Spec: batchV1.JobSpec{
			TTLSecondsAfterFinished: &jobDesc.TTLSecondsOnFinish,
			Template: coreV1.PodTemplateSpec{
				Spec: coreV1.PodSpec{
					RestartPolicy: coreV1.RestartPolicyNever,
					Containers: []coreV1.Container{
						{
							Name:    jobDesc.ContainerName,
							Image:   jobDesc.ContainerImage,
							Command: jobDesc.Command,
							Args:    jobDesc.Args,
						},
					},
					ImagePullSecrets: []coreV1.LocalObjectReference{
						{
							Name: jobDesc.ImagePullSecret,
						},
					},
				},
			},
		},
	}

	return client.BatchV1().Jobs(namespace).Create(context.TODO(), job, metaV1.CreateOptions{})
}
