package cmd

import (
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	completionUse = `completion [bash|zsh|fish]`

	completionShort = i18n.T(`Generate completion script.`)

	completionLong = templates.LongDesc(i18n.T(`
		To load completions:
		
		Bash:

		$ source <(bbctl completion bash)

		To load completions for each session, execute once:
		
		Linux:
		
		$ bbctl completion bash > /etc/bash_completion.d/bbctl
		
		macOS:

		$ bbctl completion bash > /usr/local/etc/bash_completion.d/bbctl

		Zsh:

		If shell completion is not already enabled in your environment,
		you will need to enable it.  You can execute the following once:

		$ echo "autoload -U compinit; compinit" >> ~/.zshrc

		To load completions for each session, execute once:
		
		$ bbctl completion zsh > "${fpath[1]}/_bbctl"

		Note: You will need to start a new shell for this setup to take effect.

		fish:

		$ bbctl completion fish | source

		To load completions for each session, execute once:
		
		$ bbctl completion fish > ~/.config/fish/completions/bbctl.fish `))
)

// NewCompletionCmd - create a new Cobra completion command
func NewCompletionCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	var err error
	cmd := &cobra.Command{
		Use:                   completionUse,
		Short:                 completionShort,
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletion(streams.Out)
			case "zsh":
				err = cmd.Root().GenZshCompletion(streams.Out)
			case "fish":
				err = cmd.Root().GenFishCompletion(streams.Out, true)
			}
		},
	}
	factory.GetLoggingClient().HandleError("Unable to generate completion script: %v", err)

	return cmd
}
