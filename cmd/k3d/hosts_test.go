package k3d

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	apiV1Beta1 "istio.io/api/networking/v1beta1"
	apisV1Beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

const (
	callerIdentityAccount = "123456789012"
	callerIdentityArn     = "arn:aws:iam::123456789012:user/developer"
	privateIPConst        = "192.192.192.192"
	publicIPConst         = "172.172.172.172"
)

func TestK3d_NewHostsCmd(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewHostsCmd(factory, streams)
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "hosts", cmd.Use)
}

func TestK3d_NewHostsCmd_Run(t *testing.T) {
	// Arrange
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
	account := callerIdentityAccount
	arn := callerIdentityArn
	callerIdentity := bbAwsUtil.CallerIdentity{
		GetCallerIdentityOutput: sts.GetCallerIdentityOutput{
			Account: &account,
			Arn:     &arn,
		},
		Username: "developer",
	}
	reservationID := "r-1234567890abcdef0"
	instanceID := "i-1234567890abcdef0"
	privateIP := privateIPConst
	publicIP := publicIPConst
	clusterIPs := []bbAwsUtil.ClusterIP{
		{
			ReservationID: &reservationID,
			InstanceID:    &instanceID,
			IP:            &privateIP,
			IsPublic:      false,
		},
		{
			ReservationID: &reservationID,
			InstanceID:    &instanceID,
			IP:            &publicIP,
			IsPublic:      true,
		},
	}

	vsTypeMeta := metaV1.TypeMeta{
		Kind:       "VirtualService",
		APIVersion: "networking.istio.io/v1beta1",
	}
	vsObjectMeta := metaV1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
	}
	vs := apisV1Beta1.VirtualService{
		TypeMeta:   vsTypeMeta,
		ObjectMeta: vsObjectMeta,
		Spec: apiV1Beta1.VirtualService{
			Hosts: []string{"test1", "test2"},
		},
	}
	vsList := apisV1Beta1.VirtualServiceList{
		Items: []*apisV1Beta1.VirtualService{
			&vs,
		},
	}
	factory := bbTestUtil.GetFakeFactory()
	factory.SetCallerIdentity(&callerIdentity)
	factory.SetClusterIPs(&clusterIPs)
	factory.SetVirtualServices(&vsList)
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	cmd := NewHostsCmd(factory, streams)
	cmd.SetArgs([]string{"--private-ip"})
	// Act
	assert.Nil(t, cmd.Execute())
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "hosts", cmd.Use)
	assert.Empty(t, errout.String())
	assert.Empty(t, in.String())
	assert.Equal(t, fmt.Sprintf("%v\t%v\t%v\n", privateIP, vs.Spec.Hosts[0], vs.Spec.Hosts[1]), out.String())
}
