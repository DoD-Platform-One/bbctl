package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client holds the method signatures for an AWS client.
type Client interface {
	Config(context.Context) (*aws.Config, error)
	GetClusterIPs(context.Context, DescribeInstancesAPI, string, FilterExposure) ([]ClusterIP, error)
	GetSortedClusterIPs(context.Context, DescribeInstancesAPI, string, FilterExposure) (SortedClusterIPs, error)
	GetEc2Client(context.Context, *aws.Config) (*ec2.Client, error)
	GetIdentity(context.Context, GetCallerIdentityAPI) (*CallerIdentity, error)
	GetStsClient(context.Context, *aws.Config) (*sts.Client, error)
}

// ConfigFunc type
type ConfigFunc func(context.Context) (*aws.Config, error)

// GetClusterIPsFunc type
type GetClusterIPsFunc func(context.Context, DescribeInstancesAPI, string, FilterExposure) ([]ClusterIP, error)

// GetSortedClusterIPsFunc type
type GetSortedClusterIPsFunc func(context.Context, DescribeInstancesAPI, string, FilterExposure) (SortedClusterIPs, error)

// GetEc2ClientFunc type
type GetEc2ClientFunc func(context.Context, *aws.Config) (*ec2.Client, error)

// GetIdentityFunc type
type GetIdentityFunc func(context.Context, GetCallerIdentityAPI) (*CallerIdentity, error)

// GetStsClientFunc type
type GetStsClientFunc func(context.Context, *aws.Config) (*sts.Client, error)

// awsClient is composed of functions to interact with AWS API
type awsClient struct {
	config                  ConfigFunc
	getClusterIps           GetClusterIPsFunc
	getSortedClusterIPsFunc GetSortedClusterIPsFunc
	getEc2Client            GetEc2ClientFunc
	getIdentity             GetIdentityFunc
	getStsClient            GetStsClientFunc
}

// NewClient returns a new AWS client with the provided configuration
func NewClient(
	config ConfigFunc,
	getClusterIPs GetClusterIPsFunc,
	getSortedClusterIPsFunc GetSortedClusterIPsFunc,
	getEc2Client GetEc2ClientFunc,
	getIdentity GetIdentityFunc,
	getStsClient GetStsClientFunc,
) Client {
	return &awsClient{
		config:                  config,
		getClusterIps:           getClusterIPs,
		getSortedClusterIPsFunc: getSortedClusterIPsFunc,
		getEc2Client:            getEc2Client,
		getIdentity:             getIdentity,
		getStsClient:            getStsClient,
	}
}

// Config - get the AWS SDK configuration
func (c *awsClient) Config(ctx context.Context) (*aws.Config, error) {
	return c.config(ctx)
}

// GetClusterIPs - get the cluster IPs
func (c *awsClient) GetClusterIPs(ctx context.Context, api DescribeInstancesAPI, username string, filterExposure FilterExposure) ([]ClusterIP, error) {
	return c.getClusterIps(ctx, api, username, filterExposure)
}

// GetSortedClusterIPs - get the sorted cluster IPs
func (c *awsClient) GetSortedClusterIPs(ctx context.Context, api DescribeInstancesAPI, username string, filterExposure FilterExposure) (SortedClusterIPs, error) {
	return c.getSortedClusterIPsFunc(ctx, api, username, filterExposure)
}

// GetEc2Client - get the EC2 client
func (c *awsClient) GetEc2Client(ctx context.Context, awsConfig *aws.Config) (*ec2.Client, error) {
	return c.getEc2Client(ctx, awsConfig)
}

// GetIdentity - get the AWS caller identity
func (c *awsClient) GetIdentity(ctx context.Context, api GetCallerIdentityAPI) (*CallerIdentity, error) {
	return c.getIdentity(ctx, api)
}

// GetStsClient - get the STS client
func (c *awsClient) GetStsClient(ctx context.Context, awsConfig *aws.Config) (*sts.Client, error) {
	return c.getStsClient(ctx, awsConfig)
}
