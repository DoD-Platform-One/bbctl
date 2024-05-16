package k3d

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
)

var (
	sshUse = `ssh`

	sshShort = i18n.T(`SSH to the primary instance of your k3d cluster`)

	sshLong = templates.LongDesc(i18n.T(`SSH to the primary instance of your k3d cluster`))

	sshExample = templates.Examples(i18n.T(`
	    # Open an SSH session to your K3d cluster
		bbctl k3d ssh`))
)

// NewSSHCmd - command to ssh to your k3d cluster
func NewSSHCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     sshUse,
		Short:   sshShort,
		Long:    sshLong,
		Example: sshExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sshToK3dCluster(factory, cmd, streams, args)
		},
		SilenceUsage: true,
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)

	loggingClient.HandleError("Unable to bind flags: %v",
		// make sure to sync this default with the one in the configuration schema
		configClient.SetAndBindFlag("ssh-username", "ubuntu", "Username to use for SSH connection"),
	)
	loggingClient.HandleError("Unable to bind flags: %v",
		configClient.SetAndBindFlag("dry-run", false, "Print command but don't actually establish an SSH connection"),
	)

	return cmd
}

// sshToK3dCluster - Open an SSH session to your cluster
func sshToK3dCluster(factory bbUtil.Factory, command *cobra.Command, streams genericIOOptions.IOStreams, args []string) error {
	awsClient := factory.GetAWSClient()
	loggingClient := factory.GetLoggingClient()
	cfg := awsClient.Config(context.TODO())
	stsClient := awsClient.GetStsClient(context.TODO(), cfg)
	userInfo := awsClient.GetIdentity(context.TODO(), stsClient)
	ec2Client := awsClient.GetEc2Client(context.TODO(), cfg)
	ips, err := awsClient.GetClusterIPs(context.TODO(), ec2Client, userInfo.Username, bbAws.FilterExposurePublic)
	loggingClient.HandleError("Unable to fetch cluster information: %v", err)
	configClient, err := factory.GetConfigClient(command)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()

	loggingClient.Debug(fmt.Sprintf("Args: %v", strings.Join(args, " ")))
	sshOpts := slices.Clone(args)
	sshOpts = append(sshOpts,
		"-o",
		"IdentitiesOnly=yes",
		"-i",
		fmt.Sprintf("~/.ssh/%v-dev.pem", userInfo.Username),
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-o",
		"StrictHostKeyChecking=no",
		fmt.Sprintf("%v@%v", config.K3dSshConfiguration.User, *ips[0].IP),
	)
	loggingClient.Debug(fmt.Sprintf("Running ssh command: %v", strings.Join(sshOpts, " ")))
	cmd := exec.Command("ssh", sshOpts...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = nil
	dryRun, _ := command.Flags().GetBool("dry-run")
	if !dryRun {
		err = cmd.Run()
	} else {
		fmt.Fprint(streams.Out, cmd.String())
	}
	return err
}
