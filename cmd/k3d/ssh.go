package k3d

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
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

	sshUsername = ""
)

// NewSSHCmd - command to ssh to your k3d cluster
func NewSSHCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     sshUse,
		Short:   sshShort,
		Long:    sshLong,
		Example: sshExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(sshToK3dCluster(factory, args))
		},
	}

	cmd.Flags().StringVar(&sshUsername, "ssh-username", "", "Username to use for SSH connection")

	return cmd
}

// sshToK3dCluster - Open an SSH session to your cluster
func sshToK3dCluster(factory bbUtil.Factory, args []string) error {
	if sshUsername != "" {
		viper.Set("ssh-username", sshUsername)
	}
	username := viper.GetString("ssh-username")
	if username == "" {
		username = "ubuntu"
	}
	awsClient := factory.GetAWSClient()
	loggingClient := factory.GetLoggingClient()
	cfg := awsClient.Config(context.TODO())
	stsClient := awsClient.GetStsClient(context.TODO(), cfg)
	userInfo := awsClient.GetIdentity(context.TODO(), stsClient)
	ec2Client := awsClient.GetEc2Client(context.TODO(), cfg)
	ips, err := awsClient.GetClusterIPs(context.TODO(), ec2Client, userInfo.Username, bbAws.FilterExposurePublic)
	loggingClient.HandleError("Unable to fetch cluster information: %v", err)
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
		fmt.Sprintf("%v@%v", username, *ips[0].IP),
	)
	cmd := exec.Command("ssh", sshOpts...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}
