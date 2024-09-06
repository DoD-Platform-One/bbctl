package k3d

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
)

var (
	shellProfileUse = `shellprofile`

	shellProfileShort = i18n.T(`Generates a shell profile for k3d cluster`)

	shellProfileLong = templates.LongDesc(
		i18n.T(
			`Generates a shell profile (BASH compatible) to set up your environment for a k3d cluster`,
		),
	)

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
			return shellProfileCluster(factory, cmd)
		},
	}

	return cmd
}

// shellProfileCluster - Returns the error (nil if no error) when generating a BASH compatible shell profile for your cluster
func shellProfileCluster(factory bbUtil.Factory, cobraCmd *cobra.Command) error {
	outputClient, err := factory.GetOutputClient(cobraCmd)
	if err != nil {
		return fmt.Errorf("Unable to  create output client: %w", err)
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
	ips, err := awsClient.GetSortedClusterIPs(
		context.TODO(),
		ec2Client,
		userInfo.Username,
		bbAws.FilterExposureAll,
	)
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

	// Prepare the shell profile output
	shellProfileOutput := &outputSchema.ShellProfileOutput{
		KubeConfig:       fmt.Sprintf("~/.kube/%v-dev-config", userInfo.Username),
		BB_K3D_PUBLICIP:  publicIP,
		BB_K3D_PRIVATEIP: privateIP,
	}

	// Output the data using the outputClient
	if err := outputClient.Output(shellProfileOutput); err != nil {
		return fmt.Errorf("error outputting shell profile: %w", err)
	}

	return nil
}
