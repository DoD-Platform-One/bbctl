package aws

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	bbUtilTestLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/log"
)

type testClientObjects struct {
	T                 *testing.T
	Client            Client
	ConfigFunc        ConfigFunc
	Config            *aws.Config
	GetClusterIPsFunc GetClusterIPsFunc
	ClusterIPs        []ClusterIP
	GetEc2ClientFunc  GetEc2ClientFunc
	Ec2Client         *ec2.Client
	GetIdentityFunc   GetIdentityFunc
	CallerIdentity    *CallerIdentity
	GetStsClientFunc  GetStsClientFunc
	StsClient         *sts.Client
	LoggingClient     bbLog.Client
}

func createTestClient(t *testing.T, stringBuilder strings.Builder) testClientObjects {
	t.Helper()
	publicIP := publicIPConst
	reservationID := "r-1234567890abcdef0"
	instanceID := "i-1234567890abcdef0"
	config := aws.Config{
		Region: "us-gov-west-1",
	}
	clusterIPs := []ClusterIP{
		{
			IP:            &publicIP,
			ReservationID: &reservationID,
			InstanceID:    &instanceID,
			IsPublic:      true,
		},
	}
	ec2Client := ec2.Client{}
	callerIdentity := CallerIdentity{
		Username: "test",
	}
	stsClient := sts.Client{}
	configFunc := func(context.Context) (*aws.Config, error) {
		return &config, nil
	}
	getClusterIPsFunc := func(context.Context, DescribeInstancesAPI, string, FilterExposure) ([]ClusterIP, error) {
		return clusterIPs, nil
	}
	getSortedClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) (SortedClusterIPs, error) {
		var publicIPs, privateIPs []ClusterIP
		for _, ip := range clusterIPs {
			if ip.IsPublic {
				publicIPs = append(publicIPs, ip)
			} else {
				privateIPs = append(privateIPs, ip)
			}
		}
		return SortedClusterIPs{
			PublicIPs:  publicIPs,
			PrivateIPs: privateIPs,
		}, nil
	}
	getEc2ClientFunc := func(context.Context, *aws.Config) (*ec2.Client, error) {
		return &ec2Client, nil
	}
	getIdentityFunc := func(context.Context, GetCallerIdentityAPI) (*CallerIdentity, error) {
		return &callerIdentity, nil
	}
	getStsClientFunc := func(context.Context, *aws.Config) (*sts.Client, error) {
		return &stsClient, nil
	}
	logFunc := func(args ...string) {
		for _, arg := range args {
			stringBuilder.WriteString(arg)
		}
	}
	loggingClient := bbUtilTestLog.NewFakeClient(logFunc)
	client := NewClient(configFunc, getClusterIPsFunc, getSortedClusterIPs, getEc2ClientFunc, getIdentityFunc, getStsClientFunc)
	assert.NotNil(t, client)
	return testClientObjects{
		T:                 t,
		Client:            client,
		Config:            &config,
		ConfigFunc:        configFunc,
		GetClusterIPsFunc: getClusterIPsFunc,
		ClusterIPs:        clusterIPs,
		GetEc2ClientFunc:  getEc2ClientFunc,
		Ec2Client:         &ec2Client,
		GetIdentityFunc:   getIdentityFunc,
		CallerIdentity:    &callerIdentity,
		GetStsClientFunc:  getStsClientFunc,
		StsClient:         &stsClient,
		LoggingClient:     loggingClient,
	}
}

func TestNewClient(t *testing.T) {
	// Arrange
	config := func(context.Context) (*aws.Config, error) {
		return &aws.Config{}, nil
	}
	getClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) ([]ClusterIP, error) {
		return []ClusterIP{}, nil
	}
	getSortedClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) (SortedClusterIPs, error) {
		return SortedClusterIPs{}, nil
	}
	getEc2Client := func(context.Context, *aws.Config) (*ec2.Client, error) {
		return &ec2.Client{}, nil
	}
	getIdentity := func(context.Context, GetCallerIdentityAPI) (*CallerIdentity, error) {
		return &CallerIdentity{}, nil
	}
	getStsClient := func(context.Context, *aws.Config) (*sts.Client, error) {
		return &sts.Client{}, nil
	}
	// Act
	client := NewClient(config, getClusterIPs, getSortedClusterIPs, getEc2Client, getIdentity, getStsClient)

	// Assert
	assert.NotNil(t, client)
}

func TestClient_Config(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalConfig := testClientObjects.Config

	// Act
	newConfig, err := client.Config(context.TODO())
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, originalConfig, newConfig)
	assert.Equal(t, *originalConfig, *newConfig)
	assert.Equal(t, "", stringBuilder.String())
}

func TestClient_GetClusterIPs(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalClusterIPs := testClientObjects.ClusterIPs

	// Act
	clusterIPs, err := client.GetClusterIPs(context.TODO(), nil, "", FilterExposureAll)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, clusterIPs)
	assert.Equal(t, originalClusterIPs, clusterIPs)
	assert.Equal(t, "", stringBuilder.String())
}

func TestClient_GetSortedClusterIPs(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalClusterIPs := testClientObjects.ClusterIPs

	// Act
	sortedClusterIPs, err := client.GetSortedClusterIPs(context.TODO(), nil, "test-user", FilterExposureAll)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, sortedClusterIPs)
	assert.Equal(t, len(originalClusterIPs), len(sortedClusterIPs.PublicIPs)+len(sortedClusterIPs.PrivateIPs))
	assert.Equal(t, "", stringBuilder.String())
}

func TestClient_GetEc2Client(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalEc2Client := testClientObjects.Ec2Client

	// Act
	ec2Client, err := client.GetEc2Client(context.TODO(), nil)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, ec2Client)
	assert.Equal(t, originalEc2Client, ec2Client)
	assert.Equal(t, "", stringBuilder.String())
}

func TestClient_GetIdentity(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalCallerIdentity := testClientObjects.CallerIdentity

	// Act
	callerIdentity, err := client.GetIdentity(context.TODO(), nil)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, callerIdentity)
	assert.Equal(t, originalCallerIdentity, callerIdentity)
	assert.Equal(t, "", stringBuilder.String())
}

func TestClient_GetStsClient(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalStsClient := testClientObjects.StsClient

	// Act
	stsClient, err := client.GetStsClient(context.TODO(), nil)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, stsClient)
	assert.Equal(t, originalStsClient, stsClient)
	assert.Equal(t, "", stringBuilder.String())
}
