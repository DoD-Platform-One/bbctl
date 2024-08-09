package cmd

import (
	"fmt"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	completionUse = `completion [bash|zsh|fish|powershell]`

	completionShort = i18n.T(`Generates a completion script for a specified shell environment.`)

	completionLong = templates.LongDesc(i18n.T(`
		Generates a completion script for a specified shell environment.

		Following script generation, it must be loaded into that environment in order to enable completions. 

		See command examples for shell-specific instructions on enabling completions.
	`))

	completionExample = templates.Examples(i18n.T(`		
		bash

		# To load completions for single session
		$ source <(bbctl completion bash)

		# To load completions for each session on linux
		# First install "bash-completion" with your OS's package manager
		$ bbctl completion bash > /etc/bash_completion.d/bbctl
		
		# To load completions for each session on macOS
		# First install "bash-completion" with your OS's package manager
		$ bbctl completion bash > /usr/local/etc/bash_completion.d/bbctl


		zsh

		# Enable shell completion if it is not already enabled in your environment
		$ echo "autoload -U compinit; compinit" >> ~/.zshrc

		# To load completions for each session
		$ bbctl completion zsh > "${fpath[1]}/_bbctl"

		Note: You may need to create the directory with "mkdir -p ${fpath[1]}" if it does not already exist.

		Note: You will need to start a new shell for this setup to take effect.

		
		fish

		# To load completions for single session
		$ bbctl completion fish | source
		
		# To load completions for each session
		$ bbctl completion fish > ~/.config/fish/completions/bbctl.fish


		PowerShell

		# To load completions for single session
 		PS> bbctl completion powershell | Out-String | Invoke -Expression

		# To load completions for every new session, run:
		PS> bbctl completion powershell > bbctl.ps1
		# and source this file from your PowerShell profile. `))
)

// NewCompletionCmd create a new Cobra completion command which generates a completion script
//
// Returns a cobra.Command configured to return a completion script for a specified shell environment
func NewCompletionCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	streams, err := factory.GetIOStream()
	if err != nil {
		return nil, fmt.Errorf("unable to get IO streams: %w", err)
	}
	includeDesc := true
	cmd := &cobra.Command{
		Use:                   completionUse,
		Short:                 completionShort,
		Long:                  completionLong,
		Example:               completionExample,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletionV2(streams.Out, includeDesc)
			case "zsh":
				err = cmd.Root().GenZshCompletion(streams.Out)
			case "fish":
				err = cmd.Root().GenFishCompletion(streams.Out, includeDesc)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletionWithDesc(streams.Out)
			default:
				return fmt.Errorf("unknown shell: %s", args[0])
			}

			if err != nil {
				return fmt.Errorf("Unable to generate completion script: %v", err)
			}

			return nil

		},
	}
	return cmd, nil
}
