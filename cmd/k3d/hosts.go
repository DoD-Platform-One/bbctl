package k3d

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbK8s "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"
)

var (
	hostsUse = `hosts`

	hostsShort = i18n.T(`Generate /etc/hosts entries for your k3d cluster`)

	hostsLong = templates.LongDesc(i18n.T(`Generate a list of hosts that reference your k3d cluster suitable for use in /etc/hosts`))

	hostsExample = templates.Examples(i18n.T(`
	    # Generate a list of hosts that reference your k3d cluster suitable for use in /etc/hosts
		bbctl k3d hosts`))
)

// NewHostsCmd - command to generate a hosts list for your k3d cluster
func NewHostsCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     hostsUse,
		Short:   hostsShort,
		Long:    hostsLong,
		Example: hostsExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(hostsListCluster(cmd, factory, streams))
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)

	loggingClient.HandleError(
		"Unable to add flags to command: %v",
		configClient.SetAndBindFlag(
			"private-ip",
			false,
			"Use the private IP instead of the public IP",
		),
	)

	return cmd
}

// hostsListCluster - command to generate a hosts list for your k3d cluster
func hostsListCluster(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	var useIP string
	var virtualServices []string

	awsClient := factory.GetAWSClient()
	loggingClient := factory.GetLoggingClient()
	cfg := awsClient.Config(context.TODO())
	stsClient := awsClient.GetStsClient(context.TODO(), cfg)
	userInfo := awsClient.GetIdentity(context.TODO(), stsClient)
	filterExposure := bbAws.FilterExposurePublic
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()
	if config.K3dSshConfiguration.PrivateIp {
		filterExposure = bbAws.FilterExposurePrivate
	}
	ec2Client := awsClient.GetEc2Client(context.TODO(), cfg)
	ips, err := awsClient.GetClusterIPs(context.TODO(), ec2Client, userInfo.Username, filterExposure)
	loggingClient.HandleError("Unable to fetch cluster information: %v", err)
	useIP = *ips[0].IP
	k8sConfig, err := bbK8s.BuildKubeConfig(config)
	loggingClient.HandleError("Unable to build k8s configuration: %v", err)
	istioClientSet, err := factory.GetIstioClientSet(k8sConfig)
	loggingClient.HandleError("Unable to create istio client: %v", err)
	istioServices, err := istioClientSet.NetworkingV1beta1().VirtualServices("").List(context.TODO(), metaV1.ListOptions{})
	loggingClient.HandleError("Unable to list istio services: %v", err)
	for _, item := range istioServices.Items {
		virtualServices = slices.Insert(virtualServices,
			0,
			item.Spec.Hosts...,
		)
	}
	msg := fmt.Sprintf("%s\t%s\n",
		useIP,
		strings.Join(virtualServices, "\t"),
	)
	_, err = streams.Out.Write([]byte(msg))
	return err
}
