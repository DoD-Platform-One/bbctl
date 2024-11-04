package k8s

import (
	"context"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// JobDesc contains all the information needed to create a new k8s Job object in a cluster using the CreateJob function
type JobDesc struct {
	Name               string
	ContainerName      string
	ContainerImage     string
	ImagePullSecret    string
	Command            []string
	Args               []string
	TTLSecondsOnFinish int32
}

// CreateJob function creates a new k8s job and deploys it into the cluster using the given parameters
//
// Returns the job and an error if there were any issues creating the job
func CreateJob(client kubernetes.Interface, namespace string, jobDesc *JobDesc) (*batchV1.Job, error) {
	runAsNonRoot := true
	runAsGroup := int64(1000)
	runAsUser := int64(1000)

	job := &batchV1.Job{
		ObjectMeta: metaV1.ObjectMeta{
			Name: jobDesc.Name,
			Labels: map[string]string{
				"job-name": jobDesc.Name,
			},
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
							SecurityContext: &coreV1.SecurityContext{
								RunAsUser:    &runAsUser,
								RunAsGroup:   &runAsGroup,
								RunAsNonRoot: &runAsNonRoot,
								Capabilities: &coreV1.Capabilities{
									Drop: []coreV1.Capability{"ALL"},
								},
							},
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
