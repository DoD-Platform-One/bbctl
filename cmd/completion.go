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

		Note: you will need to install "bash-completion" with your OS's package manager first.

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

		You may need to create the directory with "mkdir -p ${fpath[1]}" if it does not already exist.

		Note: You will need to start a new shell for this setup to take effect.

		fish:

		$ bbctl completion fish | source

		To load completions for each session, execute once:
		
		$ bbctl completion fish > ~/.config/fish/completions/bbctl.fish

		PowerShell:

 		PS> bbctl completion powershell | Out-String | Invoke -Expression

		# To load completions for every new session, run:

		PS> bbctl completion powershell > bbctl.ps1

		# and source this file from your PowerShell profile. `))
)

// NewCompletionCmd - create a new Cobra completion command
func NewCompletionCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	var err error
	includeDesc := true
	cmd := &cobra.Command{
		Use:                   completionUse,
		Short:                 completionShort,
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletionV2(streams.Out, includeDesc)
			case "zsh":
				err = cmd.Root().GenZshCompletion(streams.Out)
			case "fish":
				err = cmd.Root().GenFishCompletion(streams.Out, includeDesc)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletionWithDesc(streams.Out)
			}
		},
	}
	factory.GetLoggingClient().HandleError("Unable to generate completion script: %v", err)

	return cmd
}
