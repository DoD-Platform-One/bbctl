package aws

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const region = "us-gov-west-1"

// GenerateTestConfig - generate a test config file and set the AWS_CONFIG_FILE environment variable
func GenerateTestConfig(t *testing.T, config *string) {
	t.Helper()
	testConfigName := "bbctlTestConfig" + strconv.FormatInt(time.Now().UnixNano(), 10)
	testConfigPath := path.Join("/tmp", testConfigName)
	var testConfig string
	if config == nil {
		testConfig = fmt.Sprintf(`[default]
region = %v
output = json`, region)
	} else {
		testConfig = *config
	}
	err := os.WriteFile(testConfigPath, []byte(testConfig), 0600)
	require.NoError(t, err)
	t.Setenv("AWS_CONFIG_FILE", testConfigPath)
}

// GenerateClusterIPs - generate a list of ClusterIPs
func GenerateClusterIPs(_ *testing.T) []ClusterIP {
	publicIP := publicIPConst
	privateIP := privateIPConst
	return []ClusterIP{
		{
			IP:       &publicIP,
			IsPublic: true,
		},
		{
			IP:       &privateIP,
			IsPublic: false,
		},
	}
}

// DescribeInstances - mock DescribeInstances API
func (m MockDescribeInstancesAPI) DescribeInstances(
	ctx context.Context,
	params *ec2.DescribeInstancesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeInstancesOutput, error) {
	return m(nil, ctx, params, optFns...)
}

// MockDescribeInstancesAPI - mock DescribeInstances API (t will always be nil)
type MockDescribeInstancesAPI func(
	t *testing.T,
	ctx context.Context,
	params *ec2.DescribeInstancesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeInstancesOutput, error)

// GetCallerIdentity - mock GetCallerIdentity API
func (m MockGetCallerIdentityAPI) GetCallerIdentity(
	ctx context.Context,
	params *sts.GetCallerIdentityInput,
	optFns ...func(*sts.Options),
) (*sts.GetCallerIdentityOutput, error) {
	return m(nil, ctx, params, optFns...)
}

// MockGetCallerIdentityAPI - mock GetCallerIdentity API (t will always be nil)
type MockGetCallerIdentityAPI func(
	t *testing.T,
	ctx context.Context,
	params *sts.GetCallerIdentityInput,
	optFns ...func(*sts.Options),
) (*sts.GetCallerIdentityOutput, error)

// Helpers for testing getClusterIPs
const (
	privateIPConst = "192.192.192.192"
	publicIPConst  = "172.172.172.172"
)

func TestConvertClusterIPsPass(t *testing.T) {
	// Arrange
	reservationID := "r-1234567890abcdef0"
	instanceID := "i-1234567890abcdef0"
	privateIPAddress := privateIPConst
	publicIPAddress := publicIPConst
	reservation := awsTypes.Reservation{
		ReservationId: &reservationID,
	}
	instance := awsTypes.Instance{
		InstanceId:       &instanceID,
		PrivateIpAddress: &privateIPAddress,
		PublicIpAddress:  &publicIPAddress,
	}
	// Act
	clusterIPs := convertClusterIPs(reservation, instance)
	// Assert
	assert.NotNil(t, clusterIPs)
	assert.Len(t, clusterIPs, 2)
	assert.Equal(t, privateIPAddress, *clusterIPs[0].IP)
	assert.Equal(t, reservationID, *clusterIPs[0].ReservationID)
	assert.Equal(t, instanceID, *clusterIPs[0].InstanceID)
	assert.False(t, clusterIPs[0].IsPublic)
	assert.Equal(t, publicIPAddress, *clusterIPs[1].IP)
	assert.Equal(t, reservationID, *clusterIPs[1].ReservationID)
	assert.Equal(t, instanceID, *clusterIPs[1].InstanceID)
	assert.True(t, clusterIPs[1].IsPublic)
}

func TestConvertClusterIPsNoIPs(t *testing.T) {
	// Arrange
	reservationID := "r-1234567890abcdef0"
	instanceID := "i-1234567890abcdef0"
	reservation := awsTypes.Reservation{
		ReservationId: &reservationID,
	}
	instance := awsTypes.Instance{
		InstanceId: &instanceID,
	}
	// Act
	clusterIPs := convertClusterIPs(reservation, instance)
	// Assert
	assert.Empty(t, clusterIPs)
}

func TestEnsureContextNil(t *testing.T) {
	// Act
	ctx := ensureContext(nil) //nolint:all // staticcheck SA1012 intentionally ensuring nil won't break
	// Assert
	assert.NotNil(t, ctx)
}

func TestEnsureContextNotNil(t *testing.T) {
	// Arrange
	ctx := context.TODO()
	// Act
	newCtx := ensureContext(ctx)
	// Assert
	assert.NotNil(t, newCtx)
	assert.Equal(t, ctx, newCtx)
}

func TestEnsureConfigNil(t *testing.T) {
	// Arrange
	GenerateTestConfig(t, nil)
	// Act
	cfg, err := ensureConfig(context.TODO(), nil)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Region)
	assert.Equal(t, region, cfg.Region)
}

func TestEnsureConfigNotNil(t *testing.T) {
	// Arrange
	GenerateTestConfig(t, nil)
	// Act
	origCfg, err := ensureConfig(context.TODO(), nil)
	require.NoError(t, err)

	cfg, err := ensureConfig(context.TODO(), origCfg)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, origCfg)
	assert.Equal(t, region, origCfg.Region)
	assert.Equal(t, origCfg, cfg)
}

func TestFilterIPsByExposurePublic(t *testing.T) {
	// Arrange
	ips := GenerateClusterIPs(t)
	// Act
	filteredIPs := filterIPsByExposure(ips, FilterExposurePublic)
	// Assert
	assert.NotNil(t, filteredIPs)
	assert.Len(t, filteredIPs, 1)
	assert.Equal(t, publicIPConst, *filteredIPs[0].IP)
	assert.True(t, filteredIPs[0].IsPublic)
}

func TestFilterIPsByExposurePrivate(t *testing.T) {
	// Arrange
	ips := GenerateClusterIPs(t)
	// Act
	filteredIPs := filterIPsByExposure(ips, FilterExposurePrivate)
	// Assert
	assert.NotNil(t, filteredIPs)
	assert.Len(t, filteredIPs, 1)
	assert.Equal(t, privateIPConst, *filteredIPs[0].IP)
	assert.False(t, filteredIPs[0].IsPublic)
}

func TestFilterIPsByExposureAll(t *testing.T) {
	// Arrange
	ips := GenerateClusterIPs(t)
	// Act
	filteredIPs := filterIPsByExposure(ips, FilterExposureAll)
	// Assert
	assert.NotNil(t, filteredIPs)
	assert.Len(t, filteredIPs, len(ips))
	assert.Equal(t, filteredIPs[0].IP, ips[0].IP)
	assert.Equal(t, filteredIPs[0].IsPublic, ips[0].IsPublic)
	assert.Equal(t, filteredIPs[1].IP, ips[1].IP)
	assert.Equal(t, filteredIPs[1].IsPublic, ips[1].IsPublic)
}

func TestToCallerIdentityPass(t *testing.T) {
	// Arrange
	arn := "arn:aws:iam::123456789012:user/test-user"
	output := &sts.GetCallerIdentityOutput{
		Arn: &arn,
	}
	// Act
	ci := toCallerIdentity(output)
	// Assert
	assert.NotNil(t, ci)
	assert.Equal(t, "test-user", ci.Username)
	assert.Equal(t, arn, *ci.Arn)
}

func TestConfigPass(t *testing.T) {
	// Arrange
	GenerateTestConfig(t, nil)
	// Act
	cfg, err := config(context.TODO())
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Region)
	assert.Equal(t, region, cfg.Region)
}

func TestGetClusterIPsPass(t *testing.T) {
	// Arrange
	bothInstanceID := "i-1234567890abcdef0"
	privateInstanceID := "i-1234567890abcdef0"
	privateIPAddress := privateIPConst
	publicInstanceID := "i-1234567890abcdef1"
	publicIPAddress := publicIPConst
	reservationID := "r-1234567890abcdef0"
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			retVal := ec2.DescribeInstancesOutput{
				Reservations: []awsTypes.Reservation{
					{
						Instances: []awsTypes.Instance{
							// private only
							{
								InstanceId:       &privateInstanceID,
								PrivateIpAddress: &privateIPAddress,
							},
							// public only
							{
								InstanceId:      &publicInstanceID,
								PublicIpAddress: &publicIPAddress,
							},
							// both
							{
								InstanceId:       &bothInstanceID,
								PrivateIpAddress: &privateIPAddress,
								PublicIpAddress:  &publicIPAddress,
							},
						},
						ReservationId: &reservationID,
					},
				},
			}
			return &retVal, nil
		},
	)
	// Act
	ips, err := getClusterIPs(context.TODO(), api, "test-user", FilterExposureAll)
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, ips)
	assert.Len(t, ips, 4)
	assert.Equal(t, privateInstanceID, *ips[0].InstanceID)
	assert.Equal(t, privateIPAddress, *ips[0].IP)
	assert.False(t, ips[0].IsPublic)
	assert.Equal(t, publicInstanceID, *ips[1].InstanceID)
	assert.Equal(t, publicIPAddress, *ips[1].IP)
	assert.True(t, ips[1].IsPublic)
	assert.Equal(t, bothInstanceID, *ips[2].InstanceID)
	assert.Equal(t, bothInstanceID, *ips[3].InstanceID)
	assert.Equal(t, publicIPAddress, *ips[3].IP)
	assert.True(t, ips[3].IsPublic)
}

func TestGetClusterIPsEmptyReservations(t *testing.T) {
	// Arrange
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []awsTypes.Reservation{},
			}, nil
		},
	)
	// Act
	ips, err := getClusterIPs(context.TODO(), api, "test-user", FilterExposureAll)
	// Assert
	assert.Empty(t, ips)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no reservations found for user test-user")
}

func TestGetClusterIPsEmptyInstances(t *testing.T) {
	// Arrange
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []awsTypes.Reservation{
					{
						Instances: []awsTypes.Instance{},
					},
				},
			}, nil
		},
	)
	// Act
	ips, err := getClusterIPs(context.TODO(), api, "test-user", FilterExposureAll)
	// Assert
	assert.Empty(t, ips)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no instances found for user test-user")
}

func TestGetClusterIPsNoMatchingInstances(t *testing.T) {
	// Arrange
	publicIP := publicIPConst
	publicInstanceID := "i-1234567890abcdef0"
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []awsTypes.Reservation{
					{
						Instances: []awsTypes.Instance{
							{
								InstanceId:      &publicInstanceID,
								PublicIpAddress: &publicIP,
							},
						},
					},
				},
			}, nil
		},
	)
	// Act
	ips, err := getClusterIPs(context.TODO(), api, "test-user", FilterExposurePrivate)
	// Assert
	assert.Empty(t, ips)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no filtered cluster IPs found for user test-user")
}

func TestGetClusterIPsError(t *testing.T) {
	// Arrange
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			return nil, assert.AnError
		},
	)
	// Act
	ips, err := getClusterIPs(context.TODO(), api, "test-user", FilterExposureAll)
	// Assert
	assert.Empty(t, ips)
	require.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestGetSortedClusterIPs(t *testing.T) {
	// Arrange
	publicIP := publicIPConst
	publicInstanceID := "i-1234567890abcdef0"
	privateIP := privateIPConst
	privateInstanceID := "i-1234567890abcdef1"
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []awsTypes.Reservation{
					{
						Instances: []awsTypes.Instance{
							{
								InstanceId:      &publicInstanceID,
								PublicIpAddress: &publicIP,
							},
							{
								InstanceId:       &privateInstanceID,
								PrivateIpAddress: &privateIP,
							},
						},
					},
				},
			}, nil
		},
	)
	// Act
	sortedClusterIPs, err := getSortedClusterIPs(context.TODO(), api, "test-user", FilterExposureAll)
	// Assert
	require.NoError(t, err)
	assert.NotNil(t, sortedClusterIPs)
	assert.Len(t, sortedClusterIPs.PublicIPs, 1)
	assert.Len(t, sortedClusterIPs.PrivateIPs, 1)
}

func TestGetSortedClusterIPsError(t *testing.T) {
	// Arrange
	api := MockDescribeInstancesAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *ec2.DescribeInstancesInput,
			_ ...func(*ec2.Options),
		) (*ec2.DescribeInstancesOutput, error) {
			return nil, assert.AnError
		},
	)
	// Act
	sortedClusterIPs, err := getSortedClusterIPs(context.TODO(), api, "test-user", FilterExposureAll)
	// Assert
	require.Error(t, err)
	assert.Equal(t, SortedClusterIPs{}, sortedClusterIPs)
	assert.Equal(t, assert.AnError, err)
}

func TestGetEc2ClientPass(t *testing.T) {
	// Arrange
	GenerateTestConfig(t, nil)
	// Act
	ec2Client, err := getEc2Client(context.TODO(), nil)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, ec2Client)
}

func TestGetIdentityPass(t *testing.T) {
	// Arrange
	arn := "arn:aws:iam::123456789012:user/test-user"
	output := &sts.GetCallerIdentityOutput{
		Arn: &arn,
	}
	api := MockGetCallerIdentityAPI(
		func(
			_ *testing.T,
			_ context.Context,
			_ *sts.GetCallerIdentityInput,
			_ ...func(*sts.Options),
		) (*sts.GetCallerIdentityOutput, error) {
			return output, nil
		},
	)
	// Act
	identity, err := getIdentity(context.TODO(), api)
	require.NoError(t, err)

	// Assert
	assert.NotNil(t, identity)
	assert.Equal(t, "test-user", identity.Username)
	assert.Equal(t, arn, *identity.Arn)
}

func TestGetStsClientPass(t *testing.T) {
	// Arrange
	GenerateTestConfig(t, nil)
	// Act
	stsClient, ert := getStsClient(context.TODO(), nil)
	require.NoError(t, ert)

	// Assert
	assert.NotNil(t, stsClient)
}
