package aws

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
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
	publicIP := publicIPConst
	reservationID := "r-1234567890abcdef0"
	instanceID := "i-1234567890abcdef0"
	config := aws.Config{
		Region: "us-west-2",
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
	configFunc := func(context.Context, bbLog.Client) *aws.Config {
		return &config
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
	getEc2ClientFunc := func(context.Context, bbLog.Client, *aws.Config) *ec2.Client {
		return &ec2Client
	}
	getIdentityFunc := func(context.Context, bbLog.Client, GetCallerIdentityAPI) *CallerIdentity {
		return &callerIdentity
	}
	getStsClientFunc := func(context.Context, bbLog.Client, *aws.Config) *sts.Client {
		return &stsClient
	}
	logFunc := func(args ...string) {
		for _, arg := range args {
			stringBuilder.WriteString(arg)
		}
	}
	loggingClient := bbUtilTestLog.NewFakeClient(logFunc)
	client, err := NewClient(configFunc, getClusterIPsFunc, getSortedClusterIPs, getEc2ClientFunc, getIdentityFunc, getStsClientFunc, loggingClient)
	assert.Nil(t, err)
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
	var stringBuilder strings.Builder
	logFunc := func(args ...string) {
		for _, arg := range args {
			stringBuilder.WriteString(arg)
		}
	}
	loggingClient := bbUtilTestLog.NewFakeClient(logFunc)
	config := func(context.Context, bbLog.Client) *aws.Config {
		return &aws.Config{}
	}
	getClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) ([]ClusterIP, error) {
		return []ClusterIP{}, nil
	}
	getSortedClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) (SortedClusterIPs, error) {
		return SortedClusterIPs{}, nil
	}
	getEc2Client := func(context.Context, bbLog.Client, *aws.Config) *ec2.Client {
		return &ec2.Client{}
	}
	getIdentity := func(context.Context, bbLog.Client, GetCallerIdentityAPI) *CallerIdentity {
		return &CallerIdentity{}
	}
	getStsClient := func(context.Context, bbLog.Client, *aws.Config) *sts.Client {
		return &sts.Client{}
	}
	// Act
	client, err := NewClient(config, getClusterIPs, getSortedClusterIPs, getEc2Client, getIdentity, getStsClient, loggingClient)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.Empty(t, stringBuilder.String())
}

func TestNewClient_WithNilLoggingClient(t *testing.T) {
	// Arrange
	config := func(context.Context, bbLog.Client) *aws.Config {
		return &aws.Config{}
	}
	getClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) ([]ClusterIP, error) {
		return []ClusterIP{}, nil
	}
	getSortedClusterIPs := func(context.Context, DescribeInstancesAPI, string, FilterExposure) (SortedClusterIPs, error) {
		return SortedClusterIPs{}, nil
	}
	getEc2Client := func(context.Context, bbLog.Client, *aws.Config) *ec2.Client {
		return &ec2.Client{}
	}
	getIdentity := func(context.Context, bbLog.Client, GetCallerIdentityAPI) *CallerIdentity {
		return &CallerIdentity{}
	}
	getStsClient := func(context.Context, bbLog.Client, *aws.Config) *sts.Client {
		return &sts.Client{}
	}
	// Act
	client, err := NewClient(config, getClusterIPs, getSortedClusterIPs, getEc2Client, getIdentity, getStsClient, nil)
	// Assert
	assert.NotNil(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "loggingClient is nil, but is required for awsClient", err.Error())
}

func TestClient_Config(t *testing.T) {
	// Arrange
	var stringBuilder strings.Builder
	testClientObjects := createTestClient(t, stringBuilder)
	client := testClientObjects.Client
	originalConfig := testClientObjects.Config
	// Act
	newConfig := client.Config(context.TODO())
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
	assert.Nil(t, err)
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
	assert.Nil(t, err)
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
	ec2Client := client.GetEc2Client(context.TODO(), nil)
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
	callerIdentity := client.GetIdentity(context.TODO(), nil)
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
	stsClient := client.GetStsClient(context.TODO(), nil)
	// Assert
	assert.NotNil(t, stsClient)
	assert.Equal(t, originalStsClient, stsClient)
	assert.Equal(t, "", stringBuilder.String())
}
