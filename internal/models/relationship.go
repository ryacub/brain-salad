package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RelationshipType defines the type of relationship between two ideas
type RelationshipType string

const (
	// DependsOn indicates source idea depends on target idea being completed first
	DependsOn RelationshipType = "depends_on"

	// RelatedTo indicates ideas are related but not dependent
	RelatedTo RelationshipType = "related_to"

	// PartOf indicates source idea is part of target idea
	PartOf RelationshipType = "part_of"

	// Parent indicates source idea is parent of target idea
	Parent RelationshipType = "parent"

	// Child indicates source idea is child of target idea
	Child RelationshipType = "child"

	// Duplicate indicates ideas are duplicates
	Duplicate RelationshipType = "duplicate"

	// Blocks indicates source idea blocks target idea
	Blocks RelationshipType = "blocks"

	// BlockedBy indicates source idea is blocked by target idea
	BlockedBy RelationshipType = "blocked_by"

	// SimilarTo indicates ideas are similar in content or goal
	SimilarTo RelationshipType = "similar_to"
)

// AllRelationshipTypes returns all valid relationship types
func AllRelationshipTypes() []RelationshipType {
	return []RelationshipType{
		DependsOn, RelatedTo, PartOf, Parent, Child,
		Duplicate, Blocks, BlockedBy, SimilarTo,
	}
}

// IsValid checks if the relationship type is valid
func (rt RelationshipType) IsValid() bool {
	for _, valid := range AllRelationshipTypes() {
		if rt == valid {
			return true
		}
	}
	return false
}

// String returns the string representation
func (rt RelationshipType) String() string {
	return string(rt)
}

// ParseRelationshipType parses a string into a RelationshipType
func ParseRelationshipType(s string) (RelationshipType, error) {
	rt := RelationshipType(s)
	if !rt.IsValid() {
		return "", fmt.Errorf("invalid relationship type: %s", s)
	}
	return rt, nil
}

// GetInverse returns the inverse relationship type
// Example: Parent -> Child, DependsOn -> BlockedBy
func (rt RelationshipType) GetInverse() (RelationshipType, bool) {
	inverseMap := map[RelationshipType]RelationshipType{
		Parent:    Child,
		Child:     Parent,
		DependsOn: BlockedBy,
		BlockedBy: DependsOn,
		Blocks:    BlockedBy,
	}

	if inverse, ok := inverseMap[rt]; ok {
		return inverse, true
	}
	return "", false
}

// IsSymmetric returns true if the relationship is symmetric
// Example: RelatedTo, SimilarTo, Duplicate are symmetric
func (rt RelationshipType) IsSymmetric() bool {
	symmetric := []RelationshipType{RelatedTo, SimilarTo, Duplicate}
	for _, s := range symmetric {
		if rt == s {
			return true
		}
	}
	return false
}

// IdeaRelationship represents a relationship between two ideas
type IdeaRelationship struct {
	ID               string           `json:"id" db:"id"`
	SourceIdeaID     string           `json:"source_idea_id" db:"source_idea_id"`
	TargetIdeaID     string           `json:"target_idea_id" db:"target_idea_id"`
	RelationshipType RelationshipType `json:"relationship_type" db:"relationship_type"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
}

// Validate checks if the relationship is valid
func (r *IdeaRelationship) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("relationship ID cannot be empty")
	}
	if r.SourceIdeaID == "" {
		return fmt.Errorf("source idea ID cannot be empty")
	}
	if r.TargetIdeaID == "" {
		return fmt.Errorf("target idea ID cannot be empty")
	}
	if r.SourceIdeaID == r.TargetIdeaID {
		return fmt.Errorf("cannot create relationship from idea to itself")
	}
	if !r.RelationshipType.IsValid() {
		return fmt.Errorf("invalid relationship type: %s", r.RelationshipType)
	}
	return nil
}

// NewIdeaRelationship creates a new relationship with validation
func NewIdeaRelationship(sourceID, targetID string, relType RelationshipType) (*IdeaRelationship, error) {
	rel := &IdeaRelationship{
		ID:               uuid.New().String(),
		SourceIdeaID:     sourceID,
		TargetIdeaID:     targetID,
		RelationshipType: relType,
		CreatedAt:        time.Now().UTC(),
	}

	if err := rel.Validate(); err != nil {
		return nil, err
	}

	return rel, nil
}
