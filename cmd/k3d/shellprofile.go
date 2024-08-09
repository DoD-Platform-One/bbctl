package k3d

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
)

var (
	shellProfileUse = `shellprofile`

	shellProfileShort = i18n.T(`Generates a shell profile for k3d cluster`)

	shellProfileLong = templates.LongDesc(i18n.T(`Generates a shell profile (BASH compatible) to set up your environment for a k3d cluster`))

	shellProfileExample = templates.Examples(i18n.T(`
	    # Generate a profile suitable for inclusion in your ~/.profile
		bbctl k3d shellprofile`))
)

// NewShellProfileCmd - Returns a command to generate a shell profile for a k3d cluster using shellProfileCluster
func NewShellProfileCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     shellProfileUse,
		Short:   shellProfileShort,
		Long:    shellProfileLong,
		Example: shellProfileExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return shellProfileCluster(factory)
		},
	}

	return cmd
}

// shellProfileCluster - Returns the error (nil if no error) when generating a BASH compatible shell profile for your cluster
func shellProfileCluster(factory bbUtil.Factory) error {
	streams, err := factory.GetIOStream()
	if err != nil {
		return fmt.Errorf("unable to get IO streams: %w", err)
	}
	awsClient, err := factory.GetAWSClient()
	if err != nil {
		return fmt.Errorf("unable to get AWS client: %w", err)
	}
	cfg, err := awsClient.Config(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to get AWS SDK configuration: %w", err)
	}
	stsClient, err := awsClient.GetStsClient(context.TODO(), cfg)
	if err != nil {
		return fmt.Errorf("unable to get STS client: %w", err)
	}
	userInfo, err := awsClient.GetIdentity(context.TODO(), stsClient)
	if err != nil {
		return fmt.Errorf("unable to get AWS identity: %w", err)
	}
	ec2Client, err := awsClient.GetEc2Client(context.TODO(), cfg)
	if err != nil {
		return fmt.Errorf("unable to get EC2 client: %w", err)
	}
	ips, err := awsClient.GetSortedClusterIPs(context.TODO(), ec2Client, userInfo.Username, bbAws.FilterExposureAll)
	if err != nil {
		return fmt.Errorf("unable to get cluster IPs: %w", err)
	}
	var publicIP, privateIP string
	if len(ips.PublicIPs) > 0 {
		publicIP = *ips.PublicIPs[0].IP
	}
	if len(ips.PrivateIPs) > 0 {
		privateIP = *ips.PrivateIPs[0].IP
	}

	output := [3]string{
		fmt.Sprintf("export KUBECONFIG=~/.kube/%v-dev-config\n", userInfo.Username),
		fmt.Sprintf("export BB_K3D_PUBLICIP=%v\n", publicIP),
		fmt.Sprintf("export BB_K3D_PRIVATEIP=%v\n", privateIP),
	}
	for _, str := range output {
		_, err = streams.Out.Write([]byte(str))
		if err != nil {
			return err
		}
	}

	return nil
}
