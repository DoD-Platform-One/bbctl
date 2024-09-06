package k3d

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

func TestK3d_ShellProfileUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	errOut := streams.ErrOut.(*bytes.Buffer)
	// Act
	cmd := NewShellProfileCmd(factory)
	cmd.SetArgs([]string{"-h"})
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "shellprofile", cmd.Use)
	assert.Empty(t, errOut.String())
}

func TestK3d_ShellProfiile(t *testing.T) {
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
	viperInstance.Set("output-config.format", "text")
	kubeConfExport := fmt.Sprintf(
		"export KUBECONFIG=~/.kube/%v-dev-config\n",
		callerIdentity.Username,
	)
	privateIpExport := fmt.Sprintf("export BB_K3D_PUBLICIP=%v\n", publicIP)
	publicIpExport := fmt.Sprintf("export BB_K3D_PRIVATEIP=%v\n", privateIP)
	// Act
	cmd := NewShellProfileCmd(factory)
	// Assert
	assert.Equal(t, "shellprofile", cmd.Use)
	assert.Nil(t, cmd.Execute())
	assert.Empty(t, in.String())
	assert.Empty(t, errOut.String())
	assert.Contains(t, out.String(), kubeConfExport)
	assert.Contains(t, out.String(), privateIpExport)
	assert.Contains(t, out.String(), publicIpExport)
}

func TestK3d_ShellProfileError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	in := streams.In.(*bytes.Buffer)
	out := streams.Out.(*bytes.Buffer)
	errOut := streams.ErrOut.(*bytes.Buffer)
	streams.Out = apiWrappers.CreateFakeWriterFromReaderWriter(t, false, true, out)
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
	factory.SetCallerIdentity(&callerIdentity)
	factory.SetClusterIPs(&clusterIPs)
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("output-config.format", "text")
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	cmd := NewShellProfileCmd(factory)
	// Act
	err := cmd.Execute()
	// Assert
	assert.NotNil(t, err)
	unwrappedErr := err
	for unwrappedErr != nil {
		if _, ok := unwrappedErr.(*apiWrappers.FakeWriterError); ok {
			break
		}
		unwrappedErr = errors.Unwrap(unwrappedErr)
	}
	assert.IsType(t, &apiWrappers.FakeWriterError{}, unwrappedErr)
	assert.Equal(t, "shellprofile", cmd.Use)
	assert.Empty(t, in.String())
	assert.Empty(t, out.String())
	assert.Empty(t, errOut.String())
}

func TestK3d_ShellProfileErrors(t *testing.T) {
	tests := []struct {
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

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, _ := factory.GetIOStream()
	out := streams.Out.(*bytes.Buffer)
	streams.Out = apiWrappers.CreateFakeWriterFromReaderWriter(t, false, true, out)
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
	cmd := NewShellProfileCmd(factory)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetCallerIdentity(&callerIdentity)
			factory.SetClusterIPs(&clusterIPs)
			viperInstance, _ := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("output-config.format", "yaml")

			viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

			// Trigger our errors
			test.errorFunc(factory)

			err := shellProfileCluster(factory, cmd)

			assert.NotNil(t, err)
			assert.Equal(t, test.errmsg, err.Error())
		})
	}
}
