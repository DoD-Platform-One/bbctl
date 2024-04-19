package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
)

// NewFakeClient - returns a new Fake AWS client with the provided options
func NewFakeClient(
	clusterIPs []bbAws.ClusterIP,
	config *aws.Config,
	ec2Client *ec2.Client,
	identity *bbAws.CallerIdentity,
	stsClient *sts.Client,
) (bbAws.Client, error) {
	return &FakeClient{
		clusterIPs: clusterIPs,
		config:     config,
		ec2Client:  ec2Client,
		identity:   identity,
		stsClient:  stsClient,
	}, nil
}

// FakeClient - fake client
type FakeClient struct {
	clusterIPs []bbAws.ClusterIP
	config     *aws.Config
	ec2Client  *ec2.Client
	identity   *bbAws.CallerIdentity
	stsClient  *sts.Client
}

// Config implements aws.Client.
func (c *FakeClient) Config(ctx context.Context) *aws.Config {
	return c.config
}

// GetClusterIPs implements aws.Client.
func (c *FakeClient) GetClusterIPs(ctx context.Context, api bbAws.DescribeInstancesAPI, username string, filterExposure bbAws.FilterExposure) ([]bbAws.ClusterIP, error) {
	return c.clusterIPs, nil
}

// GetSortedClusterIPs implements aws.Client.
func (c *FakeClient) GetSortedClusterIPs(ctx context.Context, api bbAws.DescribeInstancesAPI, username string, filterExposure bbAws.FilterExposure) (bbAws.SortedClusterIPs, error) {
	var publicIPs, privateIPs []bbAws.ClusterIP
	for _, ip := range c.clusterIPs {
		if ip.IsPublic {
			publicIPs = append(publicIPs, ip)
		} else {
			privateIPs = append(privateIPs, ip)
		}
	}
	return bbAws.SortedClusterIPs{
		PublicIPs:  publicIPs,
		PrivateIPs: privateIPs,
	}, nil
}

// GetEc2Client implements aws.Client.
func (c *FakeClient) GetEc2Client(ctx context.Context, awsConfig *aws.Config) *ec2.Client {
	return c.ec2Client
}

// GetIdentity implements aws.Client.
func (c *FakeClient) GetIdentity(ctx context.Context, api bbAws.GetCallerIdentityAPI) *bbAws.CallerIdentity {
	return c.identity
}

// GetStsClient implements aws.Client.
func (c *FakeClient) GetStsClient(ctx context.Context, awsConfig *aws.Config) *sts.Client {
	return c.stsClient
}
