package k3d

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
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

func TestK3d_ShellProfileError(t *testing.T) {
	// Arrange
	streams, in, out, errout := genericIOOptions.NewTestIOStreams()
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
	factory := bbTestUtil.GetFakeFactory()
	factory.SetCallerIdentity(&callerIdentity)
	factory.SetClusterIPs(&clusterIPs)
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	cmd := NewShellProfileCmd(factory, streams)
	// Act
	err := cmd.Execute()
	// Assert
	assert.NotNil(t, err)
	assert.IsType(t, &apiWrappers.FakeWriterError{}, err)
	assert.True(t, streams.Out.(*apiWrappers.FakeWriter).ShouldError)
	assert.Equal(t, "shellprofile", cmd.Use)
	assert.Empty(t, in.String())
	assert.Empty(t, out.String())
	assert.Empty(t, errout.String())
}

func TestK3d_ShellProfileErrors(t *testing.T) {

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
			name: "ErrorGettingSortedClusterIPs",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.AWS.GetSortedClusterIPs = true
			},
			errmsg: "unable to get cluster IPs: failed to get sorted cluster IPs",
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

			err := shellProfileCluster(factory, streams)

			assert.NotNil(t, err)
			assert.Equal(t, test.errmsg, err.Error())
		})
	}
}
