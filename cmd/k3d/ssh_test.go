package k3d

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
)

func TestK3d_SshUsage(t *testing.T) {
	// Arrange
	streams, _, _, errout := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, sshCmdError := NewSSHCmd(factory, streams)
	cmd.SetArgs([]string{"-h"})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, sshCmdError)
	assert.Equal(t, "ssh", cmd.Use)
	assert.Empty(t, errout.String())
}

func TestK3d_SshDryRun(t *testing.T) {
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

	// Act
	cmd, sshCmdError := NewSSHCmd(factory, streams)
	cmd.SetArgs([]string{"--dry-run"})
	// Assert
	assert.Equal(t, "ssh", cmd.Use)
	assert.Nil(t, cmd.Execute())
	assert.Nil(t, sshCmdError)
	assert.Empty(t, in.String())
	assert.Empty(t, errout.String())
	assert.Contains(t, out.String(), fmt.Sprintf("/usr/bin/ssh -o IdentitiesOnly=yes -i ~/.ssh/%v-dev.pem -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ubuntu@%s", callerIdentity.Username, privateIP))
}

func TestK3d_SshBadArgs(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
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

	// Act
	cmd, sshCmdError := NewSSHCmd(factory, streams)
	cmd.SetArgs([]string{"--ssh-username="})
	result := cmd.Execute()
	// Assert
	assert.Nil(t, sshCmdError)
	assert.Error(t, result)
	if exiterr, ok := result.(*exec.ExitError); ok {
		assert.Equal(t, exiterr.ExitCode(), 255)
	}
}
