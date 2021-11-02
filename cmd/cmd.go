package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	cmdUse   = `bbctl`
	cmdShort = `BigBang command-line tool.`
	cmdLong  = `BigBang command-line tool allows you to run commands against Kubernetes clusters 
		to simplify development, deployment, auditing, and troubleshooting of BigBang.`
	cmdExample = `
		# Print version  
		bbctl version`
)

// bbctlCmd represents the base command when called without any subcommands.
var bbctlCmd = &cobra.Command{
	Use:     cmdUse,
	Short:   cmdShort,
	Long:    cmdLong,
	Example: cmdExample,
}

// Execute adds all bbctl child commands to the base command and sets flags appropriately.
func Execute() {
	cobra.CheckErr(bbctlCmd.Execute())
}

// Add kubectl cli flags to bbctl to allow reuse, e.g., leverage --namespace flag available
// in kubectl rather than creating a similar flag for bbctl.
func init() {

	cobra.OnInitialize(initConfig)

	flags := pflag.NewFlagSet("bbctl", pflag.ExitOnError)
	flags.AddFlagSet(bbctlCmd.PersistentFlags())
	pflag.CommandLine = flags

	// This set of flags is the one used for the kubectl configuration such as:
	// namespace, kube-config, insecure, and so on
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(flags)

	// It is a set of flags related to a specific resource such as: label selector
	kubeResouceBuilderFlags := genericclioptions.NewResourceBuilderFlags()
	kubeResouceBuilderFlags.AddFlags(flags)
}

// initConfig reads in ENV variables if set.
func initConfig() {
	// automatically read in environment variables that match supported flags
	// e.g. kubeconfig is a recognized flag so the corresponding env variable is KUBECONFIG
	viper.AutomaticEnv()
}
