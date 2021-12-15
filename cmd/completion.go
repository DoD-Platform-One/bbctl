package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
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
func NewCompletionCmd(factory bbutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   completionUse,
		Short:                 completionShort,
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(streams.Out)
			case "zsh":
				cmd.Root().GenZshCompletion(streams.Out)
			case "fish":
				cmd.Root().GenFishCompletion(streams.Out, true)
			}
		},
	}

	return cmd
}
