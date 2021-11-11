package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/cmd"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
	bbk8sutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/k8s"
)

func main() {
	flags := pflag.NewFlagSet("bbctl", pflag.ExitOnError)

	cobra.OnInitialize(func() {
		// automatically read in environment variables that match supported flags
		// e.g. kubeconfig is a recognized flag so the corresponding env variable is KUBECONFIG
		viper.AutomaticEnv()
	})

	factory := bbutil.NewFactory(flags)

	bbctlCmd := cmd.NewRootCmd(factory, bbk8sutil.GetIOStream())

	flags.AddFlagSet(bbctlCmd.PersistentFlags())
	pflag.CommandLine = flags

	// This set of flags is the one used for the kubectl configuration such as:
	// namespace, kube-config, insecure, and so on
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(flags)

	// It is a set of flags related to a specific resource such as: label selector
	kubeResouceBuilderFlags := genericclioptions.NewResourceBuilderFlags()
	kubeResouceBuilderFlags.AddFlags(flags)

	cobra.CheckErr(bbctlCmd.Execute())
}
