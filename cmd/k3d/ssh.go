package k3d

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
)

var (
	sshUse = `ssh`

	sshShort = i18n.T(`SSH to the k3d cluster`)

	sshLong = templates.LongDesc(i18n.T(`SSH to the primary instance of your k3d cluster`))

	sshExample = templates.Examples(i18n.T(`
	    # Open an SSH session to your K3d cluster
		bbctl k3d ssh`))
)

// NewSSHCmd - Returns a command to ssh to your k3d cluster using sshToK3dCluster
func NewSSHCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     sshUse,
		Short:   sshShort,
		Long:    sshLong,
		Example: sshExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return sshToK3dCluster(factory, cmd, args)
		},
		SilenceUsage: true,
	}

	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("unable to get config client: %w", err)
	}
	// make sure to sync this default with the one in the configuration schema
	err = configClient.SetAndBindFlag("ssh-username", "", "ubuntu", "Username to use for SSH connection")
	if err != nil {
		return nil, fmt.Errorf("unable to bind flags: %w", err)
	}
	err = configClient.SetAndBindFlag("dry-run", "", false, "Print command but don't actually establish an SSH connection")
	if err != nil {
		return nil, fmt.Errorf("unable to bind flags: %w", err)
	}

	return cmd, nil
}

// sshToK3dCluster - Returns an error (nil if no error) when opening an SSH session to your cluster
func sshToK3dCluster(factory bbUtil.Factory, command *cobra.Command, args []string) error {
	streams, err := factory.GetIOStream()
	if err != nil {
		return fmt.Errorf("unable to get IO stream: %w", err)
	}
	configClient, err := factory.GetConfigClient(command)
	if err != nil {
		return fmt.Errorf("unable to get config client: %w", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}
	awsClient, err := factory.GetAWSClient()
	if err != nil {
		return fmt.Errorf("unable to get AWS client: %w", err)
	}
	loggingClient, err := factory.GetLoggingClient()
	if err != nil {
		return fmt.Errorf("unable to get logging client: %w", err)
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
	ips, err := awsClient.GetClusterIPs(context.TODO(), ec2Client, userInfo.Username, bbAws.FilterExposurePublic)
	if err != nil {
		return fmt.Errorf("unable to get cluster IPs: %w", err)
	}
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
