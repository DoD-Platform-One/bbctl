package aws

import (
	"context"
	"errors"

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
func (c *FakeClient) Config(_ context.Context) (*aws.Config, error) {
	if c.setFail.Config {
		return nil, errors.New("failed to get AWS config")
	}
	return c.config, nil
}

// GetClusterIPs returns the configured client clusterIPs object
//
// Cannot return an error
func (c *FakeClient) GetClusterIPs(_ context.Context, _ bbAws.DescribeInstancesAPI, _ string, _ bbAws.FilterExposure) ([]bbAws.ClusterIP, error) {
	if c.setFail.GetClusterIPs {
		return nil, errors.New("failed to get cluster IPs")
	}
	return c.clusterIPs, nil
}

// GetSortedClusterIPs returns the configured client cluster IPs divded into private and public
//
// Cannot return an error
func (c *FakeClient) GetSortedClusterIPs(_ context.Context, _ bbAws.DescribeInstancesAPI, _ string, _ bbAws.FilterExposure) (bbAws.SortedClusterIPs, error) {
	var publicIPs, privateIPs []bbAws.ClusterIP
	if c.setFail.GetSortedClusterIPs {
		return bbAws.SortedClusterIPs{}, errors.New("failed to get sorted cluster IPs")
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
func (c *FakeClient) GetEc2Client(_ context.Context, _ *aws.Config) (*ec2.Client, error) {
	if c.setFail.GetEc2Client {
		return nil, errors.New("failed to get EC2 client")
	}
	return c.ec2Client, nil
}

// GetIdentity returns the configured client identity object
func (c *FakeClient) GetIdentity(_ context.Context, _ bbAws.GetCallerIdentityAPI) (*bbAws.CallerIdentity, error) {
	if c.setFail.GetIdentity {
		return nil, errors.New("failed to get AWS identity")
	}
	return c.identity, nil
}

// GetStsClient returns the configured client stsClient object
func (c *FakeClient) GetStsClient(_ context.Context, _ *aws.Config) (*sts.Client, error) {
	if c.setFail.GetStsClient {
		return nil, errors.New("failed to get STS client")
	}
	return c.stsClient, nil
}
