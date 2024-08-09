package k3d

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestK3d_SshUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd, sshCmdError := NewSSHCmd(factory)
	cmd.SetArgs([]string{"-h"})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, sshCmdError)
	assert.Equal(t, "ssh", cmd.Use)
	assert.Empty(t, errOut.String())
}

func TestK3d_SshDryRun(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
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
	factory.SetCallerIdentity(&callerIdentity)
	factory.SetClusterIPs(&clusterIPs)
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	// Act
	cmd, sshCmdError := NewSSHCmd(factory)
	cmd.SetArgs([]string{"--dry-run"})
	// Assert
	assert.Equal(t, "ssh", cmd.Use)
	assert.Nil(t, cmd.Execute())
	assert.Nil(t, sshCmdError)
	assert.Empty(t, in.String())
	assert.Empty(t, errOut.String())
	assert.Contains(t, out.String(), fmt.Sprintf("/usr/bin/ssh -o IdentitiesOnly=yes -i ~/.ssh/%v-dev.pem -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ubuntu@%s", callerIdentity.Username, privateIP))
}

func TestK3d_SshBadArgs(t *testing.T) {
	// Arrange
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
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	// Act
	cmd, sshCmdError := NewSSHCmd(factory)
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
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetConfigClient = true
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	// Act
	cmd, sshCmdError := NewSSHCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, sshCmdError)
	assert.Equal(t, "unable to get config client: failed to get config client", sshCmdError.Error())
}

func TestK3d_SshErrorSettingSSHUsername(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, value interface{}, description string) error {
		return expectedError
	}

	logClient, _ := factory.GetLoggingClient()
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, sshCmdError := NewSSHCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, sshCmdError)
	assert.Equal(t, fmt.Sprintf("unable to bind flags: %s", expectedError.Error()), sshCmdError.Error())
}

func TestK3d_SshErrorSettingDryRun(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, value interface{}, description string) error {
		if name == "dry-run" {
			return expectedError
		}
		return nil
	}

	logClient, _ := factory.GetLoggingClient()
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, sshCmdError := NewSSHCmd(factory)

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

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
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
			viperInstance, _ := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

			// Trigger our errors
			test.errorFunc(factory)

			cmd := NewShellProfileCmd(factory)
			err := sshToK3dCluster(factory, cmd, nil)

			assert.NotNil(t, err)
			assert.Equal(t, test.errmsg, err.Error())
		})
	}
}

func TestSSHFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}

	logClient, _ := factory.GetLoggingClient()
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &logClient, nil, viperInstance)
	factory.SetConfigClient(client)
	cmd := NewShellProfileCmd(factory)
	// Act
	err := sshToK3dCluster(factory, cmd, []string{})

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
