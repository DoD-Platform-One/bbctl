package k3d

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
)

var (
	shellProfileUse = `shellprofile`

	shellProfileShort = i18n.T(`Generate shell profile for k3d cluster`)

	shellProfileLong = templates.LongDesc(i18n.T(`Generate a shell profile (BASH compatible) to set up your environment for a k3d cluster`))

	shellProfileExample = templates.Examples(i18n.T(`
	    # Generate a profile suitable for inclusion in your ~/.profile
		bbctl k3d shellprofile`))
)

// NewShellProfileCmd - command to generate a shell profile for a k3d cluster
func NewShellProfileCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     shellProfileUse,
		Short:   shellProfileShort,
		Long:    shellProfileLong,
		Example: shellProfileExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(shellProfileCluster(factory, streams))
		},
	}

	return cmd
}

// shellProfileCluster - generate a BASH compatible shell profile for your cluster
func shellProfileCluster(factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	awsClient := factory.GetAWSClient()
	loggingClient := factory.GetLoggingClient()
	cfg := awsClient.Config(context.TODO())
	stsClient := awsClient.GetStsClient(context.TODO(), cfg)
	userInfo := awsClient.GetIdentity(context.TODO(), stsClient)
	ec2Client := awsClient.GetEc2Client(context.TODO(), cfg)
	ips, err := awsClient.GetSortedClusterIPs(context.TODO(), ec2Client, userInfo.Username, bbAws.FilterExposureAll)
	loggingClient.HandleError("Unable to fetch cluster information: %v", err)
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
