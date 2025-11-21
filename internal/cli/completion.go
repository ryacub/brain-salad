package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for tm.

To load completions:

Bash:
  $ source <(tm completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ tm completion bash > /etc/bash_completion.d/tm
  # macOS:
  $ tm completion bash > /usr/local/etc/bash_completion.d/tm

Zsh:
  $ tm completion zsh > "${fpath[1]}/_tm"
  $ exec zsh

Fish:
  $ tm completion fish | source
  $ tm completion fish > ~/.config/fish/completions/tm.fish

PowerShell:
  PS> tm completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	// Skip initialization for completion command (doesn't need telos.md)
	// Override the parent's PersistentPreRunE with a no-op function
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return cmd
}
