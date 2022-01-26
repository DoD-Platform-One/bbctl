package k8s

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func CreateJob(client kubernetes.Interface, namespace string, jobDesc *JobDesc) (*batchv1.Job, error) {

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobDesc.Name,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &jobDesc.TTLSecondsOnFinish,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    jobDesc.ContainerName,
							Image:   jobDesc.ContainerImage,
							Command: jobDesc.Command,
							Args:    jobDesc.Args,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: jobDesc.ImagePullSecret,
						},
					},
				},
			},
		},
	}

	return client.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
}
