package k3d

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
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

func TestK3d_SshErrorGettingConfigClient(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetConfigClient = true
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	// Act
	cmd, sshCmdError := NewSSHCmd(factory, streams)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, sshCmdError)
	assert.Equal(t, "unable to get config client: failed to get config client", sshCmdError.Error())
}

func TestK3d_SshErrorSettingSSHUsername(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, value interface{}, description string) error {
		return expectedError
	}

	logClient := factory.GetLoggingClient()
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, sshCmdError := NewSSHCmd(factory, streams)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, sshCmdError)
	assert.Equal(t, fmt.Sprintf("unable to bind flags: %s", expectedError.Error()), sshCmdError.Error())
}

func TestK3d_SshErrorSettingDryRun(t *testing.T) {
	// Arrange
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, value interface{}, description string) error {
		if name == "dry-run" {
			return expectedError
		}
		return nil
	}

	logClient := factory.GetLoggingClient()
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, sshCmdError := NewSSHCmd(factory, streams)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, sshCmdError)
	assert.Equal(t, fmt.Sprintf("unable to bind flags: %s", expectedError.Error()), sshCmdError.Error())
}

func TestK3d_sshToK3dClusterErrors(t *testing.T) {
	var tests = []struct {
		name string
		// errorFunc is a function that will be called with the awsClient and factory
		// at the start of a test case to allow setting flags to force errors
		errorFunc func(factory *bbTestUtil.FakeFactory)
		errmsg    string
	}{
		{
			name: "ErrorGettingAWSClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetAWSClient = true
			},
			errmsg: "unable to get AWS client: failed to get AWS client",
		},
		{
			name: "ErrorGettingAWSConfig",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.AWS.Config = true
			},
			errmsg: "unable to get AWS SDK configuration: failed to get AWS config",
		},
		{
			name: "ErrorGettingAWSStsClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.AWS.GetStsClient = true
			},
			errmsg: "unable to get STS client: failed to get STS client",
		},
		{
			name: "ErrorGettingAWSIdentity",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.AWS.GetIdentity = true
			},
			errmsg: "unable to get AWS identity: failed to get AWS identity",
		},
		{
			name: "ErrorGettingAWSEc2Client",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.AWS.GetEc2Client = true
			},
			errmsg: "unable to get EC2 client: failed to get EC2 client",
		},
		{
			name: "ErrorGettingClusterIPs",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.AWS.GetClusterIPs = true
			},
			errmsg: "unable to get cluster IPs: failed to get cluster IPs",
		},
		{
			name: "ErrorGettingConfigClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetConfigClient = true
			},
			errmsg: "unable to get config client: failed to get config client",
		},
	}

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	streams.Out = apiWrappers.CreateFakeWriterFromStream(t, true, streams.Out)
	account := callerIdentityAccount
	arn := callerIdentityArn
	callerIdentity := bbAwsUtil.CallerIdentity{
		GetCallerIdentityOutput: sts.GetCallerIdentityOutput{
			Account: &account,
			Arn:     &arn,
		},
		Username: "developer",
	}
	clusterIPs := []bbAwsUtil.ClusterIP{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetCallerIdentity(&callerIdentity)
			factory.SetClusterIPs(&clusterIPs)
			viperInstance := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

			// Trigger our errors
			test.errorFunc(factory)

			cmd := NewShellProfileCmd(factory, streams)
			err := sshToK3dCluster(factory, cmd, streams, nil)

			assert.NotNil(t, err)
			assert.Equal(t, test.errmsg, err.Error())
		})
	}
}
