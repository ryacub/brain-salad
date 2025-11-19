package models_test

import (
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRelationshipType_IsValid tests the IsValid method for all relationship types
func TestRelationshipType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		relType  models.RelationshipType
		expected bool
	}{
		{"valid depends_on", models.DependsOn, true},
		{"valid related_to", models.RelatedTo, true},
		{"valid part_of", models.PartOf, true},
		{"valid parent", models.Parent, true},
		{"valid child", models.Child, true},
		{"valid duplicate", models.Duplicate, true},
		{"valid blocks", models.Blocks, true},
		{"valid blocked_by", models.BlockedBy, true},
		{"valid similar_to", models.SimilarTo, true},
		{"invalid type", models.RelationshipType("invalid"), false},
		{"empty string", models.RelationshipType(""), false},
		{"random string", models.RelationshipType("foo_bar"), false},
		{"uppercase", models.RelationshipType("DEPENDS_ON"), false},
		{"mixed case", models.RelationshipType("DependsOn"), false},
		{"with spaces", models.RelationshipType("depends on"), false},
		{"special characters", models.RelationshipType("depends@on"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.relType.IsValid())
		})
	}
}

// TestRelationshipType_String tests the String method
func TestRelationshipType_String(t *testing.T) {
	tests := []struct {
		name     string
		relType  models.RelationshipType
		expected string
	}{
		{"depends_on", models.DependsOn, "depends_on"},
		{"related_to", models.RelatedTo, "related_to"},
		{"part_of", models.PartOf, "part_of"},
		{"parent", models.Parent, "parent"},
		{"child", models.Child, "child"},
		{"duplicate", models.Duplicate, "duplicate"},
		{"blocks", models.Blocks, "blocks"},
		{"blocked_by", models.BlockedBy, "blocked_by"},
		{"similar_to", models.SimilarTo, "similar_to"},
		{"invalid type", models.RelationshipType("invalid"), "invalid"},
		{"empty string", models.RelationshipType(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.relType.String())
		})
	}
}

// TestParseRelationshipType tests parsing strings into relationship types
func TestParseRelationshipType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  models.RelationshipType
		wantError bool
	}{
		{"valid depends_on", "depends_on", models.DependsOn, false},
		{"valid related_to", "related_to", models.RelatedTo, false},
		{"valid part_of", "part_of", models.PartOf, false},
		{"valid parent", "parent", models.Parent, false},
		{"valid child", "child", models.Child, false},
		{"valid duplicate", "duplicate", models.Duplicate, false},
		{"valid blocks", "blocks", models.Blocks, false},
		{"valid blocked_by", "blocked_by", models.BlockedBy, false},
		{"valid similar_to", "similar_to", models.SimilarTo, false},
		{"invalid type", "foo", models.RelationshipType(""), true},
		{"empty string", "", models.RelationshipType(""), true},
		{"uppercase", "DEPENDS_ON", models.RelationshipType(""), true},
		{"mixed case", "DependsOn", models.RelationshipType(""), true},
		{"with spaces", "depends on", models.RelationshipType(""), true},
		{"special characters", "depends@on", models.RelationshipType(""), true},
		{"numeric", "123", models.RelationshipType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := models.ParseRelationshipType(tt.input)

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid relationship type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestRelationshipType_GetInverse tests inverse relationship mappings
func TestRelationshipType_GetInverse(t *testing.T) {
	tests := []struct {
		name        string
		relType     models.RelationshipType
		expectedInv models.RelationshipType
		hasInverse  bool
	}{
		{"parent to child", models.Parent, models.Child, true},
		{"child to parent", models.Child, models.Parent, true},
		{"depends_on to blocked_by", models.DependsOn, models.BlockedBy, true},
		{"blocked_by to depends_on", models.BlockedBy, models.DependsOn, true},
		{"blocks to blocked_by", models.Blocks, models.BlockedBy, true},
		{"related_to has no inverse", models.RelatedTo, "", false},
		{"similar_to has no inverse", models.SimilarTo, "", false},
		{"duplicate has no inverse", models.Duplicate, "", false},
		{"part_of has no inverse", models.PartOf, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, hasInv := tt.relType.GetInverse()

			assert.Equal(t, tt.hasInverse, hasInv)
			if tt.hasInverse {
				assert.Equal(t, tt.expectedInv, inv)
			} else {
				assert.Equal(t, models.RelationshipType(""), inv)
			}
		})
	}
}

// TestRelationshipType_GetInverse_Bidirectional verifies inverse relationships are bidirectional
func TestRelationshipType_GetInverse_Bidirectional(t *testing.T) {
	tests := []struct {
		typeA models.RelationshipType
		typeB models.RelationshipType
	}{
		{models.Parent, models.Child},
		{models.DependsOn, models.BlockedBy},
	}

	for _, tt := range tests {
		t.Run(tt.typeA.String()+"<->"+tt.typeB.String(), func(t *testing.T) {
			// typeA -> typeB
			invA, hasInvA := tt.typeA.GetInverse()
			require.True(t, hasInvA)
			assert.Equal(t, tt.typeB, invA)

			// typeB -> typeA (bidirectional check)
			invB, hasInvB := tt.typeB.GetInverse()
			require.True(t, hasInvB)
			assert.Equal(t, tt.typeA, invB)
		})
	}
}

// TestRelationshipType_IsSymmetric tests symmetric relationship detection
func TestRelationshipType_IsSymmetric(t *testing.T) {
	tests := []struct {
		name        string
		relType     models.RelationshipType
		isSymmetric bool
	}{
		{"related_to is symmetric", models.RelatedTo, true},
		{"similar_to is symmetric", models.SimilarTo, true},
		{"duplicate is symmetric", models.Duplicate, true},
		{"depends_on is NOT symmetric", models.DependsOn, false},
		{"parent is NOT symmetric", models.Parent, false},
		{"child is NOT symmetric", models.Child, false},
		{"blocks is NOT symmetric", models.Blocks, false},
		{"blocked_by is NOT symmetric", models.BlockedBy, false},
		{"part_of is NOT symmetric", models.PartOf, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isSymmetric, tt.relType.IsSymmetric())
		})
	}
}

// TestAllRelationshipTypes tests that all relationship types are returned
func TestAllRelationshipTypes(t *testing.T) {
	types := models.AllRelationshipTypes()

	assert.Len(t, types, 9, "should have exactly 9 relationship types")

	expectedTypes := map[models.RelationshipType]bool{
		models.DependsOn: false,
		models.RelatedTo: false,
		models.PartOf:    false,
		models.Parent:    false,
		models.Child:     false,
		models.Duplicate: false,
		models.Blocks:    false,
		models.BlockedBy: false,
		models.SimilarTo: false,
	}

	// Mark each type as found
	for _, rt := range types {
		expectedTypes[rt] = true
	}

	// Verify all expected types were found
	for rt, found := range expectedTypes {
		assert.True(t, found, "relationship type %s should be in AllRelationshipTypes()", rt)
	}
}

// TestAllRelationshipTypes_NoDuplicates ensures no duplicate types are returned
func TestAllRelationshipTypes_NoDuplicates(t *testing.T) {
	types := models.AllRelationshipTypes()
	seen := make(map[models.RelationshipType]bool)

	for _, rt := range types {
		assert.False(t, seen[rt], "duplicate relationship type found: %s", rt)
		seen[rt] = true
	}
}

// TestAllRelationshipTypes_AllValid ensures all returned types are valid
func TestAllRelationshipTypes_AllValid(t *testing.T) {
	types := models.AllRelationshipTypes()

	for _, rt := range types {
		assert.True(t, rt.IsValid(), "relationship type %s should be valid", rt)
	}
}

// TestNewIdeaRelationship_Success tests successful relationship creation
func TestNewIdeaRelationship_Success(t *testing.T) {
	sourceID := "idea-123"
	targetID := "idea-456"
	relType := models.DependsOn

	rel, err := models.NewIdeaRelationship(sourceID, targetID, relType)

	require.NoError(t, err)
	require.NotNil(t, rel)

	// Verify all fields are populated correctly
	assert.NotEmpty(t, rel.ID, "ID should not be empty")
	assert.Equal(t, sourceID, rel.SourceIdeaID)
	assert.Equal(t, targetID, rel.TargetIdeaID)
	assert.Equal(t, relType, rel.RelationshipType)

	// Verify timestamp
	assert.False(t, rel.CreatedAt.IsZero(), "CreatedAt should not be zero")
	assert.True(t, time.Since(rel.CreatedAt) < time.Second, "CreatedAt should be recent")

	// Verify timestamp is in UTC
	assert.Equal(t, time.UTC, rel.CreatedAt.Location(), "CreatedAt should be in UTC")
}

// TestNewIdeaRelationship_AllTypes tests creation with all relationship types
func TestNewIdeaRelationship_AllTypes(t *testing.T) {
	sourceID := "idea-source"
	targetID := "idea-target"

	for _, relType := range models.AllRelationshipTypes() {
		t.Run(relType.String(), func(t *testing.T) {
			rel, err := models.NewIdeaRelationship(sourceID, targetID, relType)

			require.NoError(t, err)
			require.NotNil(t, rel)
			assert.Equal(t, relType, rel.RelationshipType)
		})
	}
}

// TestNewIdeaRelationship_ValidationErrors tests error cases
func TestNewIdeaRelationship_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		sourceID  string
		targetID  string
		relType   models.RelationshipType
		errorMsg  string
	}{
		{
			name:     "self-referential relationship",
			sourceID: "idea-same",
			targetID: "idea-same",
			relType:  models.DependsOn,
			errorMsg: "cannot create relationship from idea to itself",
		},
		{
			name:     "empty source ID",
			sourceID: "",
			targetID: "idea-456",
			relType:  models.DependsOn,
			errorMsg: "source idea ID cannot be empty",
		},
		{
			name:     "empty target ID",
			sourceID: "idea-123",
			targetID: "",
			relType:  models.DependsOn,
			errorMsg: "target idea ID cannot be empty",
		},
		{
			name:     "invalid relationship type",
			sourceID: "idea-123",
			targetID: "idea-456",
			relType:  models.RelationshipType("invalid"),
			errorMsg: "invalid relationship type",
		},
		{
			name:     "empty relationship type",
			sourceID: "idea-123",
			targetID: "idea-456",
			relType:  models.RelationshipType(""),
			errorMsg: "invalid relationship type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := models.NewIdeaRelationship(tt.sourceID, tt.targetID, tt.relType)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
			assert.Nil(t, rel)
		})
	}
}

// TestNewIdeaRelationship_UniqueIDs tests that each relationship gets a unique ID
func TestNewIdeaRelationship_UniqueIDs(t *testing.T) {
	sourceID := "idea-123"
	targetID := "idea-456"
	relType := models.DependsOn

	// Create multiple relationships
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		rel, err := models.NewIdeaRelationship(sourceID, targetID, relType)
		require.NoError(t, err)
		require.NotNil(t, rel)

		// Verify this ID hasn't been seen before
		assert.False(t, ids[rel.ID], "duplicate ID found: %s", rel.ID)
		ids[rel.ID] = true
	}

	assert.Len(t, ids, 100, "should have 100 unique IDs")
}

// TestIdeaRelationship_Validate tests the Validate method
func TestIdeaRelationship_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *models.IdeaRelationship
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid relationship",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "rel-123",
					SourceIdeaID:     "idea-123",
					TargetIdeaID:     "idea-456",
					RelationshipType: models.DependsOn,
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: false,
		},
		{
			name: "self-referential relationship",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "rel-123",
					SourceIdeaID:     "idea-same",
					TargetIdeaID:     "idea-same",
					RelationshipType: models.DependsOn,
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: true,
			errorMsg:  "cannot create relationship from idea to itself",
		},
		{
			name: "empty relationship ID",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "",
					SourceIdeaID:     "idea-123",
					TargetIdeaID:     "idea-456",
					RelationshipType: models.DependsOn,
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: true,
			errorMsg:  "relationship ID cannot be empty",
		},
		{
			name: "empty source ID",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "rel-123",
					SourceIdeaID:     "",
					TargetIdeaID:     "idea-456",
					RelationshipType: models.DependsOn,
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: true,
			errorMsg:  "source idea ID cannot be empty",
		},
		{
			name: "empty target ID",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "rel-123",
					SourceIdeaID:     "idea-123",
					TargetIdeaID:     "",
					RelationshipType: models.DependsOn,
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: true,
			errorMsg:  "target idea ID cannot be empty",
		},
		{
			name: "invalid relationship type",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "rel-123",
					SourceIdeaID:     "idea-123",
					TargetIdeaID:     "idea-456",
					RelationshipType: models.RelationshipType("invalid"),
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: true,
			errorMsg:  "invalid relationship type",
		},
		{
			name: "empty relationship type",
			setup: func() *models.IdeaRelationship {
				return &models.IdeaRelationship{
					ID:               "rel-123",
					SourceIdeaID:     "idea-123",
					TargetIdeaID:     "idea-456",
					RelationshipType: models.RelationshipType(""),
					CreatedAt:        time.Now().UTC(),
				}
			},
			wantError: true,
			errorMsg:  "invalid relationship type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := tt.setup()
			err := rel.Validate()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestIdeaRelationship_Validate_AllTypes tests validation with all relationship types
func TestIdeaRelationship_Validate_AllTypes(t *testing.T) {
	for _, relType := range models.AllRelationshipTypes() {
		t.Run(relType.String(), func(t *testing.T) {
			rel := &models.IdeaRelationship{
				ID:               "rel-123",
				SourceIdeaID:     "idea-123",
				TargetIdeaID:     "idea-456",
				RelationshipType: relType,
				CreatedAt:        time.Now().UTC(),
			}

			err := rel.Validate()
			assert.NoError(t, err)
		})
	}
}

// TestIdeaRelationship_EmptyStruct tests validation of empty struct
func TestIdeaRelationship_EmptyStruct(t *testing.T) {
	rel := &models.IdeaRelationship{}
	err := rel.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "relationship ID cannot be empty")
}

// TestNewIdeaRelationship_CreatedAtTimestamp tests timestamp properties
func TestNewIdeaRelationship_CreatedAtTimestamp(t *testing.T) {
	before := time.Now().UTC()
	rel, err := models.NewIdeaRelationship("idea-123", "idea-456", models.DependsOn)
	after := time.Now().UTC()

	require.NoError(t, err)
	require.NotNil(t, rel)

	// Timestamp should be between before and after
	assert.False(t, rel.CreatedAt.Before(before), "CreatedAt should not be before test start")
	assert.False(t, rel.CreatedAt.After(after), "CreatedAt should not be after test end")

	// Timestamp should be in UTC
	assert.Equal(t, time.UTC, rel.CreatedAt.Location())

	// Timestamp should not be zero
	assert.False(t, rel.CreatedAt.IsZero())
}

// TestSymmetricRelationships_Properties verifies properties of symmetric relationships
func TestSymmetricRelationships_Properties(t *testing.T) {
	symmetricTypes := []models.RelationshipType{
		models.RelatedTo,
		models.SimilarTo,
		models.Duplicate,
	}

	for _, relType := range symmetricTypes {
		t.Run(relType.String(), func(t *testing.T) {
			// Symmetric relationships should be marked as symmetric
			assert.True(t, relType.IsSymmetric(), "%s should be symmetric", relType)

			// Symmetric relationships should not have an inverse
			_, hasInverse := relType.GetInverse()
			assert.False(t, hasInverse, "%s should not have an inverse", relType)
		})
	}
}

// TestAsymmetricRelationships_Properties verifies properties of asymmetric relationships
func TestAsymmetricRelationships_Properties(t *testing.T) {
	asymmetricTypes := []models.RelationshipType{
		models.DependsOn,
		models.Parent,
		models.Child,
		models.Blocks,
		models.BlockedBy,
		models.PartOf,
	}

	for _, relType := range asymmetricTypes {
		t.Run(relType.String(), func(t *testing.T) {
			// Asymmetric relationships should not be marked as symmetric
			assert.False(t, relType.IsSymmetric(), "%s should not be symmetric", relType)
		})
	}
}

// BenchmarkNewIdeaRelationship benchmarks relationship creation
func BenchmarkNewIdeaRelationship(b *testing.B) {
	sourceID := "idea-123"
	targetID := "idea-456"
	relType := models.DependsOn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = models.NewIdeaRelationship(sourceID, targetID, relType)
	}
}

// BenchmarkRelationshipType_IsValid benchmarks type validation
func BenchmarkRelationshipType_IsValid(b *testing.B) {
	rt := models.DependsOn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rt.IsValid()
	}
}

// BenchmarkRelationshipType_GetInverse benchmarks inverse lookup
func BenchmarkRelationshipType_GetInverse(b *testing.B) {
	rt := models.Parent

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rt.GetInverse()
	}
}

// BenchmarkRelationshipType_IsSymmetric benchmarks symmetric check
func BenchmarkRelationshipType_IsSymmetric(b *testing.B) {
	rt := models.RelatedTo

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rt.IsSymmetric()
	}
}

// BenchmarkParseRelationshipType benchmarks parsing
func BenchmarkParseRelationshipType(b *testing.B) {
	input := "depends_on"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = models.ParseRelationshipType(input)
	}
}

// BenchmarkIdeaRelationship_Validate benchmarks validation
func BenchmarkIdeaRelationship_Validate(b *testing.B) {
	rel := &models.IdeaRelationship{
		ID:               "rel-123",
		SourceIdeaID:     "idea-123",
		TargetIdeaID:     "idea-456",
		RelationshipType: models.DependsOn,
		CreatedAt:        time.Now().UTC(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rel.Validate()
	}
}

// BenchmarkAllRelationshipTypes benchmarks getting all types
func BenchmarkAllRelationshipTypes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = models.AllRelationshipTypes()
	}
}
