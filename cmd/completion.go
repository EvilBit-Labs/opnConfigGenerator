package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `Generate completion script for your shell.

To load completions:

Bash:
  $ source <(opnConfigGenerator completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ opnConfigGenerator completion bash > /etc/bash_completion.d/opnConfigGenerator
  # macOS:
  $ opnConfigGenerator completion bash > $(brew --prefix)/etc/bash_completion.d/opnConfigGenerator

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ opnConfigGenerator completion zsh > "${fpath[1]}/_opnConfigGenerator"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ opnConfigGenerator completion fish | source

  # To load completions for each session, execute once:
  $ opnConfigGenerator completion fish > ~/.config/fish/completions/opnConfigGenerator.fish

PowerShell:
  PS> opnConfigGenerator completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> opnConfigGenerator completion powershell > opnConfigGenerator.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}
