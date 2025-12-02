package cli

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

const (
	ideaNotFoundMessage = "(idea not found)"
)

func newLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Manage relationships between ideas",
		Long: `Manage relationships between ideas.

Link ideas together to express dependencies, hierarchies, and associations.
This helps track how ideas relate to each other and find dependency paths.

Available relationship types:
  - depends_on: Source depends on target
  - related_to: Ideas are related
  - part_of: Source is part of target
  - parent/child: Hierarchical relationship
  - duplicate: Ideas are duplicates
  - blocks/blocked_by: Blocking relationship
  - similar_to: Similar ideas`,
	}

	cmd.AddCommand(newLinkCreateCommand())
	cmd.AddCommand(newLinkListCommand())
	cmd.AddCommand(newLinkShowCommand())
	cmd.AddCommand(newLinkRemoveCommand())
	cmd.AddCommand(newLinkPathCommand())

	return cmd
}

func newLinkCreateCommand() *cobra.Command {
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "create <source-id> <target-id> <type>",
		Short: "Create a relationship between two ideas",
		Long: `Create a relationship between two ideas.

Relationship types:
  - depends_on: Source depends on target
  - related_to: Ideas are related
  - part_of: Source is part of target
  - parent/child: Hierarchical relationship
  - duplicate: Ideas are duplicates
  - blocks/blocked_by: Blocking relationship
  - similar_to: Similar ideas

Examples:
  tm link create abc123 def456 depends_on
  tm link create abc123 ghi789 related_to --no-confirm`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkCreate(args[0], args[1], args[2], noConfirm)
		},
	}

	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompt")

	return cmd
}

func newLinkListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <idea-id>",
		Short: "List all relationships for an idea",
		Long: `List all relationships for an idea.

Shows both outgoing relationships (where this idea is the source) and
incoming relationships (where this idea is the target).

Examples:
  tm link list abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkList(args[0])
		},
	}
}

func newLinkShowCommand() *cobra.Command {
	var relType string

	cmd := &cobra.Command{
		Use:   "show <idea-id>",
		Short: "Show related ideas for an idea",
		Long: `Show related ideas (not relationships) for an idea.

Displays full idea details including content and scores.
Optionally filter by relationship type.

Examples:
  tm link show abc123
  tm link show abc123 --type depends_on`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkShow(args[0], relType)
		},
	}

	cmd.Flags().StringVar(&relType, "type", "", "Filter by relationship type")

	return cmd
}

func newLinkRemoveCommand() *cobra.Command {
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "remove <relationship-id>",
		Short: "Remove a relationship between ideas",
		Long: `Remove a relationship between ideas.

You can get the relationship ID from the 'link list' command.

Examples:
  tm link remove rel123
  tm link remove rel456 --no-confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkRemove(args[0], noConfirm)
		},
	}

	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompt")

	return cmd
}

func newLinkPathCommand() *cobra.Command {
	var maxDepth int

	cmd := &cobra.Command{
		Use:   "path <source-id> <target-id>",
		Short: "Find dependency paths between ideas",
		Long: `Find all paths between two ideas using breadth-first search.

Useful for understanding how ideas are connected and identifying
dependency chains.

Examples:
  tm link path abc123 xyz789
  tm link path abc123 xyz789 --max-depth 5`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkPath(args[0], args[1], maxDepth)
		},
	}

	cmd.Flags().IntVar(&maxDepth, "max-depth", 3, "Maximum path length")

	return cmd
}

// --- Implementation Functions ---

func runLinkCreate(sourceID, targetID, relTypeStr string, noConfirm bool) error {
	// Validate relationship type
	relType, err := models.ParseRelationshipType(relTypeStr)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Invalid relationship type: %s\n", relTypeStr); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		fmt.Println("\nValid types:")
		for _, rt := range models.AllRelationshipTypes() {
			fmt.Printf("  - %s\n", rt)
		}
		return nil
	}

	// Get both ideas for confirmation
	sourceIdea, err := ctx.Repository.GetByID(sourceID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Source idea not found: %s\n", sourceID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	targetIdea, err := ctx.Repository.GetByID(targetID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Target idea not found: %s\n", targetID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	// Show confirmation
	fmt.Println()
	if _, err := infoColor.Println("Creating relationship:"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Printf("  Source: [%s] %s\n", truncateID(sourceID), cliutil.TruncateText(sourceIdea.Content, 60))
	fmt.Printf("  Target: [%s] %s\n", truncateID(targetID), cliutil.TruncateText(targetIdea.Content, 60))
	fmt.Printf("  Type: %s\n", relType)
	fmt.Println()

	// Get user confirmation
	if !noConfirm {
		if !cliutil.Confirm("Continue?") {
			if _, err := cliutil.WarningColor.Println("❌ Cancelled."); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
			return nil
		}
	}

	// Create relationship
	relationship, err := models.NewIdeaRelationship(sourceID, targetID, relType)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	err = ctx.Repository.CreateRelationship(relationship)
	if err != nil {
		return fmt.Errorf("failed to save relationship: %w", err)
	}

	if _, err := cliutil.SuccessColor.Printf("✓ Relationship created successfully (ID: %s)\n", truncateID(relationship.ID)); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	return nil
}

func runLinkList(ideaID string) error {
	// Verify idea exists
	idea, err := ctx.Repository.GetByID(ideaID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Idea not found: %s\n", ideaID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	relationships, err := ctx.Repository.GetRelationshipsForIdea(ideaID)
	if err != nil {
		return fmt.Errorf("failed to get relationships: %w", err)
	}

	if len(relationships) == 0 {
		if _, err := cliutil.WarningColor.Printf("📭 No relationships found for idea: %s\n", truncateID(ideaID)); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return nil
	}

	fmt.Println()
	if _, err := infoColor.Printf("🔗 Relationships for idea: %s\n", truncateID(ideaID)); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Printf("   %s\n", cliutil.TruncateText(idea.Content, 60))
	fmt.Println()

	// Separate outgoing and incoming relationships
	var outgoing, incoming []*models.IdeaRelationship
	for _, rel := range relationships {
		if rel.SourceIdeaID == ideaID {
			outgoing = append(outgoing, rel)
		} else {
			incoming = append(incoming, rel)
		}
	}

	// Display outgoing relationships
	if len(outgoing) > 0 {
		if _, err := cliutil.SuccessColor.Println("Outgoing (where this idea is the source):"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for i, rel := range outgoing {
			relatedIdeaID := rel.TargetIdeaID
			relatedIdea, _ := ctx.Repository.GetByID(relatedIdeaID)

			relatedContent := ideaNotFoundMessage
			if relatedIdea != nil {
				relatedContent = cliutil.TruncateText(relatedIdea.Content, 60)
			}

			fmt.Printf("  %d. %s → [%s] %s\n",
				i+1,
				rel.RelationshipType,
				truncateID(relatedIdeaID),
				relatedContent,
			)
			fmt.Printf("     ID: %s | Created: %s\n",
				truncateID(rel.ID),
				rel.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		fmt.Println()
	}

	// Display incoming relationships
	if len(incoming) > 0 {
		if _, err := infoColor.Println("Incoming (where this idea is the target):"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for i, rel := range incoming {
			relatedIdeaID := rel.SourceIdeaID
			relatedIdea, _ := ctx.Repository.GetByID(relatedIdeaID)

			relatedContent := ideaNotFoundMessage
			if relatedIdea != nil {
				relatedContent = cliutil.TruncateText(relatedIdea.Content, 60)
			}

			fmt.Printf("  %d. %s ← [%s] %s\n",
				i+1,
				rel.RelationshipType,
				truncateID(relatedIdeaID),
				relatedContent,
			)
			fmt.Printf("     ID: %s | Created: %s\n",
				truncateID(rel.ID),
				rel.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d relationships\n", len(relationships))
	return nil
}

func runLinkShow(ideaID, relTypeStr string) error {
	// Verify idea exists
	idea, err := ctx.Repository.GetByID(ideaID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Idea not found: %s\n", ideaID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	// Parse relationship type if provided
	var relType *models.RelationshipType
	if relTypeStr != "" {
		rt, err := models.ParseRelationshipType(relTypeStr)
		if err != nil {
			if _, printErr := cliutil.ErrorColor.Printf("❌ Invalid relationship type: %s\n", relTypeStr); printErr != nil {
				log.Warn().Err(printErr).Msg("failed to print error message")
			}
			fmt.Println("\nValid types:")
			for _, t := range models.AllRelationshipTypes() {
				fmt.Printf("  - %s\n", t)
			}
			return nil
		}
		relType = &rt
	}

	relatedIdeas, err := ctx.Repository.GetRelatedIdeas(ideaID, relType)
	if err != nil {
		return fmt.Errorf("failed to get related ideas: %w", err)
	}

	if len(relatedIdeas) == 0 {
		if _, err := cliutil.WarningColor.Printf("📭 No related ideas found for: %s\n", truncateID(ideaID)); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		return nil
	}

	filterText := ""
	if relType != nil {
		filterText = fmt.Sprintf(" (filtered by: %s)", *relType)
	}

	fmt.Println()
	if _, err := infoColor.Printf("🔗 Related ideas for: %s%s\n", truncateID(ideaID), filterText); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Printf("   %s\n", cliutil.TruncateText(idea.Content, 60))
	fmt.Println()

	for i, relatedIdea := range relatedIdeas {
		scoreDisplay := ""
		if relatedIdea.FinalScore > 0 {
			scoreDisplay = fmt.Sprintf(" 📊 %.1f/10", relatedIdea.FinalScore)
		}

		fmt.Printf("%d. %s\n", i+1, relatedIdea.Content)
		fmt.Printf("   ID: %s | Status: %s%s\n",
			truncateID(relatedIdea.ID),
			relatedIdea.Status,
			scoreDisplay,
		)
		fmt.Printf("   Created: %s\n", relatedIdea.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Println()
	}

	fmt.Printf("Found %d related ideas\n", len(relatedIdeas))
	return nil
}

func runLinkRemove(relationshipID string, noConfirm bool) error {
	// Get relationship details
	rel, err := ctx.Repository.GetRelationship(relationshipID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Relationship not found: %s\n", relationshipID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	// Get idea details for display
	sourceIdea, _ := ctx.Repository.GetByID(rel.SourceIdeaID)
	targetIdea, _ := ctx.Repository.GetByID(rel.TargetIdeaID)

	sourceContent := ideaNotFoundMessage
	if sourceIdea != nil {
		sourceContent = cliutil.TruncateText(sourceIdea.Content, 50)
	}

	targetContent := ideaNotFoundMessage
	if targetIdea != nil {
		targetContent = cliutil.TruncateText(targetIdea.Content, 50)
	}

	fmt.Println()
	if _, err := cliutil.WarningColor.Println("Removing relationship:"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Printf("  ID: %s\n", truncateID(relationshipID))
	fmt.Printf("  [%s] %s\n", truncateID(rel.SourceIdeaID), sourceContent)
	fmt.Printf("    %s →\n", rel.RelationshipType)
	fmt.Printf("  [%s] %s\n", truncateID(rel.TargetIdeaID), targetContent)
	fmt.Println()

	if !noConfirm {
		if !cliutil.Confirm("Are you sure?") {
			if _, err := cliutil.WarningColor.Println("❌ Removal cancelled."); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
			return nil
		}
	}

	if err := ctx.Repository.DeleteRelationship(relationshipID); err != nil {
		return fmt.Errorf("failed to remove relationship: %w", err)
	}

	if _, err := cliutil.SuccessColor.Printf("✓ Relationship removed successfully\n"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	return nil
}

func runLinkPath(sourceID, targetID string, maxDepth int) error {
	// Verify both ideas exist
	sourceIdea, err := ctx.Repository.GetByID(sourceID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Source idea not found: %s\n", sourceID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	_, err = ctx.Repository.GetByID(targetID)
	if err != nil {
		if _, printErr := cliutil.ErrorColor.Printf("❌ Target idea not found: %s\n", targetID); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print error message")
		}
		return nil
	}

	fmt.Println()
	if _, err := infoColor.Printf("🔍 Finding paths from %s to %s...\n", truncateID(sourceID), truncateID(targetID)); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	paths, err := ctx.Repository.FindRelationshipPath(sourceID, targetID, maxDepth)
	if err != nil {
		return fmt.Errorf("failed to find path: %w", err)
	}

	if len(paths) == 0 {
		if _, err := cliutil.WarningColor.Printf("❌ No path found between %s and %s\n", truncateID(sourceID), truncateID(targetID)); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		fmt.Println()
		fmt.Println("💡 Try linking ideas that might connect these two concepts")
		return nil
	}

	// Display all found paths
	for i, path := range paths {
		if _, err := cliutil.SuccessColor.Printf("Path %d (%d hops):\n", i+1, len(path)); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}

		// Start with source idea
		fmt.Printf("  [%s] %s\n", truncateID(sourceID), cliutil.TruncateText(sourceIdea.Content, 50))

		// Display each step in the path
		currentID := sourceID
		for _, rel := range path {
			// Determine next ID
			var nextID string
			if rel.SourceIdeaID == currentID {
				nextID = rel.TargetIdeaID
			} else {
				nextID = rel.SourceIdeaID
			}

			// Get next idea
			nextIdea, _ := ctx.Repository.GetByID(nextID)
			nextContent := ideaNotFoundMessage
			if nextIdea != nil {
				nextContent = cliutil.TruncateText(nextIdea.Content, 50)
			}

			fmt.Printf("    → %s →\n", rel.RelationshipType)
			fmt.Printf("  [%s] %s\n", truncateID(nextID), nextContent)

			currentID = nextID
		}
		fmt.Println()
	}

	fmt.Printf("Found %d path(s)\n", len(paths))
	return nil
}

// --- Helper Functions ---

func truncateID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
