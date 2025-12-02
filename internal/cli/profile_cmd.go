package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/profile"
	"github.com/spf13/cobra"
)

func newProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "View your scoring profile",
		Long:  `Display your current scoring profile and priorities.`,
		RunE:  runProfile,
	}

	cmd.AddCommand(newProfileResetCommand())

	return cmd
}

func newProfileResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Re-run the setup wizard",
		Long:  `Reset your profile by running the discovery wizard again.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Just run init, which handles everything
			return runInit(cmd, args)
		},
		// Skip normal initialization since we're resetting
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func runProfile(cmd *cobra.Command, args []string) error {
	if ctx.ScoringMode != ScoringModeUniversal || ctx.Profile == nil {
		fmt.Println("No profile found. Using advanced telos.md mode.")
		fmt.Printf("Telos file: %s\n", ctx.TelosPath)
		return nil
	}

	p := ctx.Profile
	headerColor := color.New(color.FgCyan, color.Bold)
	dimColor := color.New(color.FgWhite)

	fmt.Println()
	headerColor.Println("Your Scoring Profile")
	fmt.Println(strings.Repeat("─", 40))

	// Goals
	if len(p.Goals) > 0 {
		fmt.Println()
		headerColor.Println("Goals:")
		for _, goal := range p.Goals {
			fmt.Printf("  • %s\n", goal)
		}
	}

	// Avoid
	if len(p.Avoid) > 0 {
		fmt.Println()
		headerColor.Println("Avoiding:")
		for _, avoid := range p.Avoid {
			fmt.Printf("  • %s\n", avoid)
		}
	}

	// Priorities (sorted by weight)
	fmt.Println()
	headerColor.Println("Priorities:")

	type priority struct {
		name   string
		weight float64
	}
	priorities := []priority{}
	for name, weight := range p.Priorities {
		priorities = append(priorities, priority{name, weight})
	}
	// Sort by weight descending
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].weight > priorities[j].weight
	})

	descriptions := map[string]string{
		profile.DimensionCompletionLikelihood: "Will I finish this?",
		profile.DimensionSkillFit:             "Can I do this?",
		profile.DimensionTimeToDone:           "How long?",
		profile.DimensionRewardAlignment:      "What I want?",
		profile.DimensionSustainability:       "Stay motivated?",
		profile.DimensionAvoidanceFit:         "Dodge pitfalls?",
	}

	for _, pri := range priorities {
		barLen := int(pri.weight * 20)
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", 20-barLen)
		desc := descriptions[pri.name]
		if desc == "" {
			desc = pri.name
		}
		dimColor.Printf("  %s  %3.0f%%  %s\n", bar, pri.weight*100, desc)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("Profile: %s\n", ctx.ProfilePath)
	fmt.Println("Run 'tm profile reset' to reconfigure")
	fmt.Println()

	return nil
}
