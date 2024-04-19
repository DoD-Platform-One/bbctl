package aws

import (
	"context"
	"fmt"
	"strings"

	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func convertClusterIPs(reservation types.Reservation, instance types.Instance) []ClusterIP {
	var clusterIPs []ClusterIP
	if instance.PrivateIpAddress != nil {
		clusterIPs = append(clusterIPs,
			ClusterIP{
				IP:            instance.PrivateIpAddress,
				ReservationID: reservation.ReservationId,
				InstanceID:    instance.InstanceId,
				IsPublic:      false,
			},
		)
	}
	if instance.PublicIpAddress != nil {
		clusterIPs = append(clusterIPs,
			ClusterIP{
				IP:            instance.PublicIpAddress,
				ReservationID: reservation.ReservationId,
				InstanceID:    instance.InstanceId,
				IsPublic:      true,
			},
		)
	}
	return clusterIPs
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		newContext := context.TODO()
		return newContext
	}
	return ctx
}

func ensureConfig(ctx context.Context, loggingClient bbLog.Client, cfg *aws.Config) *aws.Config {
	contextToUse := ensureContext(ctx)
	if cfg == nil {
		return config(contextToUse, loggingClient)
	}
	return cfg
}

// filterIPsByExposure - filter IPs by exposure
func filterIPsByExposure(ips []ClusterIP, filterExposure FilterExposure) []ClusterIP {
	var filteredIPs []ClusterIP
	for _, ip := range ips {
		// add the IP to the list if it is public and we want public IPs
		if filterExposure == FilterExposurePublic && ip.IsPublic && ip.IP != nil {
			filteredIPs = append(filteredIPs, ip)
		}
		// add the IP to the list if it is public and we want public IPs
		if filterExposure == FilterExposurePrivate && !ip.IsPublic && ip.IP != nil {
			filteredIPs = append(filteredIPs, ip)
		}
		// add the IP to the list if we want all IPs
		if filterExposure == FilterExposureAll && ip.IP != nil {
			filteredIPs = append(filteredIPs, ip)
		}
	}
	return filteredIPs
}

// callerIdentity - AWS caller identity
func toCallerIdentity(output *sts.GetCallerIdentityOutput) *CallerIdentity {
	return &CallerIdentity{
		GetCallerIdentityOutput: *output,
		Username:                strings.Split(*output.Arn, "/")[1],
	}
}

// config - get the AWS SDK configuration
func config(ctx context.Context, loggingClient bbLog.Client) *aws.Config {
	contextToUse := ensureContext(ctx)
	cfg, err := awsConfig.LoadDefaultConfig(contextToUse)
	loggingClient.HandleError("failed to load AWS SDK configuration", err)
	return &cfg
}

// getClusterIPs - get the cluster IPs
func getClusterIPs(ctx context.Context, api DescribeInstancesAPI, username string, filterExposure FilterExposure) ([]ClusterIP, error) {
	contextToUse := ensureContext(ctx)
	result, err := api.DescribeInstances(contextToUse, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name: aws.String("tag:Name"),
				// FIXME : This is not safe, we need to wrap this dereference in something
				// that checks for nil
				Values: []string{*aws.String(fmt.Sprintf("%v-dev", username))},
			},
			{
				Name: aws.String("instance-state-name"),
				// FIXME: Same issue as above
				Values: []string{*aws.String("running")},
			},
		},
	})
	if (err != nil) || (len(result.Reservations) == 0) {
		if err == nil {
			err = fmt.Errorf("no reservations found for user %v", username)
		}
		return nil, err
	}
	var clusterIPs []ClusterIP
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			clusterIPs = append(clusterIPs, convertClusterIPs(reservation, instance)...)
		}
	}
	if len(clusterIPs) == 0 {
		err = fmt.Errorf("no instances found for user %v", username)
		return nil, err
	}
	clusterIPs = filterIPsByExposure(clusterIPs, filterExposure)
	if len(clusterIPs) == 0 {
		err = fmt.Errorf("no filtered cluster IPs found for user %v", username)
		return nil, err
	}
	return clusterIPs, nil
}

// getSortedClusterIPs - get the sorted cluster IPs
func getSortedClusterIPs(ctx context.Context, api DescribeInstancesAPI, username string, filterExposure FilterExposure) (SortedClusterIPs, error) {
	clusterIPs, err := getClusterIPs(ctx, api, username, filterExposure)
	if err != nil {
		return SortedClusterIPs{}, err
	}
	var publicIPs, privateIPs []ClusterIP
	for _, ip := range clusterIPs {
		if ip.IsPublic && ip.IP != nil && *ip.IP != "" {
			publicIPs = append(publicIPs, ip)
		} else if !ip.IsPublic && ip.IP != nil && *ip.IP != "" {
			privateIPs = append(privateIPs, ip)
		}
	}
	return SortedClusterIPs{
		PublicIPs:  publicIPs,
		PrivateIPs: privateIPs,
	}, nil
}

// getEc2Client - get the EC2 client
func getEc2Client(ctx context.Context, loggingClient bbLog.Client, awsConfig *aws.Config) *ec2.Client {
	config := ensureConfig(ctx, loggingClient, awsConfig)
	return ec2.NewFromConfig(*config)
}

// getIdentity - get the AWS identity
func getIdentity(ctx context.Context, loggingClient bbLog.Client, api GetCallerIdentityAPI) *CallerIdentity {
	contextToUse := ensureContext(ctx)
	result, err := api.GetCallerIdentity(
		contextToUse,
		&sts.GetCallerIdentityInput{},
	)
	loggingClient.HandleError("failed to get caller identity", err)
	return toCallerIdentity(result)
}

// getStsClient - get the STS client
func getStsClient(ctx context.Context, loggingClient bbLog.Client, awsConfig *aws.Config) *sts.Client {
	config := ensureConfig(ctx, loggingClient, awsConfig)
	return sts.NewFromConfig(*config)
}
