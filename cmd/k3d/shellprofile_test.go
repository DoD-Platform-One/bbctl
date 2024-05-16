package k3d

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_ShellProfileUsage(t *testing.T) {
	// Arrange
	streams, _, _, errout := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := NewShellProfileCmd(factory, streams)
	cmd.SetArgs([]string{"-h"})
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "shellprofile", cmd.Use)
	assert.Empty(t, errout.String())
}

func TestK3d_ShellProfiile(t *testing.T) {
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
	reservationID := "r-1234567890abcdef1"
	instanceID := "i-1234567890abcdef1"
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
	factory := bbTestUtil.GetFakeFactory()
	factory.SetCallerIdentity(&callerIdentity)
	factory.SetClusterIPs(&clusterIPs)
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	kubeConfExport := fmt.Sprintf("export KUBECONFIG=~/.kube/%v-dev-config\n", callerIdentity.Username)
	privateIpExport := fmt.Sprintf("export BB_K3D_PUBLICIP=%v\n", publicIP)
	publicIpExport := fmt.Sprintf("export BB_K3D_PRIVATEIP=%v\n", privateIP)
	// Act
	cmd := NewShellProfileCmd(factory, streams)
	// Assert
	assert.Equal(t, "shellprofile", cmd.Use)
	assert.Nil(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errout.String())
	assert.Contains(t, out.String(), kubeConfExport)
	assert.Contains(t, out.String(), privateIpExport)
	assert.Contains(t, out.String(), publicIpExport)
}
