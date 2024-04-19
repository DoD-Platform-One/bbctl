package deploy

import (
	"fmt"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	deployUse = `deploy`

	deployShort = i18n.T(`Manage deployments of big bang components`)

	deployLong = templates.LongDesc(i18n.T(`Manage deployments of big bang components and prerequisites`))

	deployExample = templates.Examples(i18n.T(``))
)

// NewDeployCmd - parent for deploy commands
func NewDeployCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     deployUse,
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		Run: func(cmd *cobra.Command, args []string) {
			_, err := streams.Out.Write([]byte(fmt.Sprintln("Please provide a subcommand for deploy (see help)")))
			factory.GetLoggingClient().HandleError("Unable to write to output stream", err)
		},
	}

	cmd.AddCommand(NewDeployFluxCmd(factory, streams))
	cmd.AddCommand(NewDeployBigBangCmd(factory, streams))

	return cmd
}
