package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLinkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Manage idea relationships (Coming Soon)",
		Long: `Manage relationships and connections between ideas.

This feature is planned for a future release.

Planned features:
  - Link related ideas
  - Create idea hierarchies
  - Track idea evolution`,
		RunE: runLink,
	}
}

func runLink(cmd *cobra.Command, args []string) error {
	warningColor.Println("⚠️  Link command is not yet implemented.")
	fmt.Println("This feature is planned for a future release.")
	fmt.Println()
	fmt.Println("In the meantime, you can:")
	fmt.Println("  • Use 'tm review' to browse ideas")
	fmt.Println("  • Use 'tm analyze --id <ID>' to review specific ideas")
	return nil
}
