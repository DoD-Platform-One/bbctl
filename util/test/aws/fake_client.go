package aws

import (
	"context"
	"fmt"

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
	setFail SetFail,
) (bbAws.Client, error) {
	return &FakeClient{
		clusterIPs: clusterIPs,
		config:     config,
		ec2Client:  ec2Client,
		identity:   identity,
		stsClient:  stsClient,
		setFail:    setFail,
	}, nil
}

// Flags to control the fake client behavior and force functions to fail
type SetFail struct {
	Config              bool
	GetStsClient        bool
	GetIdentity         bool
	GetEc2Client        bool
	GetClusterIPs       bool
	GetSortedClusterIPs bool
}

// FakeClient
type FakeClient struct {
	clusterIPs []bbAws.ClusterIP
	config     *aws.Config
	ec2Client  *ec2.Client
	identity   *bbAws.CallerIdentity
	stsClient  *sts.Client

	setFail SetFail
}

// Config returns the configured client config object
func (c *FakeClient) Config(ctx context.Context) (*aws.Config, error) {
	if c.setFail.Config {
		return nil, fmt.Errorf("failed to get AWS config")
	}
	return c.config, nil
}

// GetClusterIPs returns the configured client clusterIPs object
//
// Cannot return an error
func (c *FakeClient) GetClusterIPs(ctx context.Context, api bbAws.DescribeInstancesAPI, username string, filterExposure bbAws.FilterExposure) ([]bbAws.ClusterIP, error) {
	if c.setFail.GetClusterIPs {
		return nil, fmt.Errorf("failed to get cluster IPs")
	}
	return c.clusterIPs, nil
}

// GetSortedClusterIPs returns the configured client cluster IPs divded into private and public
//
// Cannot return an error
func (c *FakeClient) GetSortedClusterIPs(ctx context.Context, api bbAws.DescribeInstancesAPI, username string, filterExposure bbAws.FilterExposure) (bbAws.SortedClusterIPs, error) {
	var publicIPs, privateIPs []bbAws.ClusterIP
	if c.setFail.GetSortedClusterIPs {
		return bbAws.SortedClusterIPs{}, fmt.Errorf("failed to get sorted cluster IPs")
	}
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

// GetEc2Client returns the configured client ec2Client object
func (c *FakeClient) GetEc2Client(ctx context.Context, awsConfig *aws.Config) (*ec2.Client, error) {
	if c.setFail.GetEc2Client {
		return nil, fmt.Errorf("failed to get EC2 client")
	}
	return c.ec2Client, nil
}

// GetIdentity returns the configured client identity object
func (c *FakeClient) GetIdentity(ctx context.Context, api bbAws.GetCallerIdentityAPI) (*bbAws.CallerIdentity, error) {
	if c.setFail.GetIdentity {
		return nil, fmt.Errorf("failed to get AWS identity")
	}
	return c.identity, nil
}

// GetStsClient returns the configured client stsClient object
func (c *FakeClient) GetStsClient(ctx context.Context, awsConfig *aws.Config) (*sts.Client, error) {
	if c.setFail.GetStsClient {
		return nil, fmt.Errorf("failed to get STS client")
	}
	return c.stsClient, nil
}
