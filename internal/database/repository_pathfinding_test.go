//go:build integration
// +build integration

package database_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPathTestDB creates a temporary database for path-finding testing
func setupPathTestDB(t *testing.T) (*database.Repository, func()) {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "pathtest_*.db")
	require.NoError(t, err)
	tmpfile.Close()

	repo, err := database.NewRepository(tmpfile.Name())
	require.NoError(t, err)

	cleanup := func() {
		repo.Close()
		os.Remove(tmpfile.Name())
	}

	return repo, cleanup
}

// createTestIdea creates a test idea and returns its ID
func createTestIdea(t *testing.T, repo *database.Repository, content string) string {
	t.Helper()

	idea := models.NewIdea(content)
	err := repo.Create(idea)
	require.NoError(t, err)

	return idea.ID
}

// createTestRelationship creates a test relationship and returns its ID
func createTestRelationship(t *testing.T, repo *database.Repository,
	sourceID, targetID string, relType models.RelationshipType) string {
	t.Helper()

	rel, err := models.NewIdeaRelationship(sourceID, targetID, relType)
	require.NoError(t, err)

	err = repo.CreateRelationship(rel)
	require.NoError(t, err)

	return rel.ID
}

// TestFindRelationshipPath_DirectPath tests a simple single-hop path (A → B)
func TestFindRelationshipPath_DirectPath(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create ideas
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")

	// Create relationship: A depends_on B
	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)

	// Find path
	paths, err := repo.FindRelationshipPath(ideaA, ideaB, 3)

	require.NoError(t, err)
	assert.Len(t, paths, 1, "should find exactly 1 path")
	assert.Len(t, paths[0], 1, "path should have 1 hop")

	// Verify the relationship
	rel := paths[0][0]
	assert.Equal(t, ideaA, rel.SourceIdeaID)
	assert.Equal(t, ideaB, rel.TargetIdeaID)
	assert.Equal(t, models.DependsOn, rel.RelationshipType)
}

// TestFindRelationshipPath_MultiHop tests a multi-hop path (A → B → C → D)
func TestFindRelationshipPath_MultiHop(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create ideas
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")
	ideaD := createTestIdea(t, repo, "Idea D")

	// Create chain: A → B → C → D
	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)
	createTestRelationship(t, repo, ideaB, ideaC, models.DependsOn)
	createTestRelationship(t, repo, ideaC, ideaD, models.DependsOn)

	// Find path with sufficient depth
	paths, err := repo.FindRelationshipPath(ideaA, ideaD, 5)

	require.NoError(t, err)
	assert.Len(t, paths, 1, "should find exactly 1 path")
	assert.Len(t, paths[0], 3, "path should have 3 hops")

	// Verify each hop
	assert.Equal(t, ideaA, paths[0][0].SourceIdeaID)
	assert.Equal(t, ideaB, paths[0][0].TargetIdeaID)

	assert.Equal(t, ideaB, paths[0][1].SourceIdeaID)
	assert.Equal(t, ideaC, paths[0][1].TargetIdeaID)

	assert.Equal(t, ideaC, paths[0][2].SourceIdeaID)
	assert.Equal(t, ideaD, paths[0][2].TargetIdeaID)
}

// TestFindRelationshipPath_MultiplePaths tests finding multiple paths in a diamond structure:
//
//	  → B →
//	A       D
//	  → C →
func TestFindRelationshipPath_MultiplePaths(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create diamond structure
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")
	ideaD := createTestIdea(t, repo, "Idea D")

	// Create two paths
	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)
	createTestRelationship(t, repo, ideaA, ideaC, models.DependsOn)
	createTestRelationship(t, repo, ideaB, ideaD, models.DependsOn)
	createTestRelationship(t, repo, ideaC, ideaD, models.DependsOn)

	// Find all paths
	paths, err := repo.FindRelationshipPath(ideaA, ideaD, 5)

	require.NoError(t, err)
	assert.Len(t, paths, 2, "should find both paths")

	// Both paths should have 2 hops
	for i, path := range paths {
		assert.Len(t, path, 2, "path %d should have 2 hops", i)
	}

	// Verify paths go through different middle nodes
	middleNodes := make(map[string]bool)
	for _, path := range paths {
		// Second node in path (either B or C)
		middleNode := path[0].TargetIdeaID
		middleNodes[middleNode] = true
	}
	assert.Len(t, middleNodes, 2, "paths should go through different middle nodes")
	assert.Contains(t, middleNodes, ideaB)
	assert.Contains(t, middleNodes, ideaC)
}

// TestFindRelationshipPath_MaxDepthExceeded tests that paths exceeding max depth are not found
func TestFindRelationshipPath_MaxDepthExceeded(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create long chain: A → B → C → D → E
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")
	ideaD := createTestIdea(t, repo, "Idea D")
	ideaE := createTestIdea(t, repo, "Idea E")

	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)
	createTestRelationship(t, repo, ideaB, ideaC, models.DependsOn)
	createTestRelationship(t, repo, ideaC, ideaD, models.DependsOn)
	createTestRelationship(t, repo, ideaD, ideaE, models.DependsOn)

	// Try to find path with insufficient depth (maxDepth = 2)
	paths, err := repo.FindRelationshipPath(ideaA, ideaE, 2)

	require.NoError(t, err)
	assert.Empty(t, paths, "should not find path when depth limit is too low")

	// Now try with sufficient depth
	paths, err = repo.FindRelationshipPath(ideaA, ideaE, 5)

	require.NoError(t, err)
	assert.Len(t, paths, 1, "should find path with sufficient depth")
	assert.Len(t, paths[0], 4, "path should have 4 hops")
}

// TestFindRelationshipPath_Cycles tests path-finding in graphs with cycles
// Cycle: A → B → C → A
// The BFS treats relationships bidirectionally, so it can traverse in both directions
func TestFindRelationshipPath_Cycles(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create cycle: A → B → C → A
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")

	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)
	createTestRelationship(t, repo, ideaB, ideaC, models.DependsOn)
	createTestRelationship(t, repo, ideaC, ideaA, models.DependsOn)

	// Find path from A to C
	paths, err := repo.FindRelationshipPath(ideaA, ideaC, 5)

	require.NoError(t, err)
	// BFS finds multiple paths due to bidirectional traversal:
	// 1. Direct: A → C (via C → A backwards) - 1 hop
	// 2. Forward: A → B → C - 2 hops
	assert.GreaterOrEqual(t, len(paths), 1, "should find at least one path")

	// The shortest path should be found (1 hop via bidirectional traversal)
	foundShortPath := false
	foundLongPath := false
	for _, path := range paths {
		if len(path) == 1 {
			foundShortPath = true
		}
		if len(path) == 2 {
			foundLongPath = true
		}
	}

	assert.True(t, foundShortPath, "should find 1-hop path via bidirectional traversal")
	assert.True(t, foundLongPath, "should also find 2-hop path A→B→C")

	// Verify that the algorithm doesn't create infinite loops
	for _, path := range paths {
		assert.LessOrEqual(t, len(path), 3, "paths should not loop infinitely")
	}
}

// TestFindRelationshipPath_NoPath tests finding no path in disconnected graphs
func TestFindRelationshipPath_NoPath(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create two disconnected components
	// Component 1: A → B
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)

	// Component 2: C → D (no connection to A or B)
	ideaC := createTestIdea(t, repo, "Idea C")
	ideaD := createTestIdea(t, repo, "Idea D")
	createTestRelationship(t, repo, ideaC, ideaD, models.DependsOn)

	// Try to find path from A to D (impossible)
	paths, err := repo.FindRelationshipPath(ideaA, ideaD, 5)

	require.NoError(t, err)
	assert.Empty(t, paths, "should find no path between disconnected components")
}

// TestFindRelationshipPath_Bidirectional tests bidirectional relationship traversal
// The BFS should be able to traverse relationships in both directions
func TestFindRelationshipPath_Bidirectional(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create ideas
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")

	// Create chain: A → B → C
	createTestRelationship(t, repo, ideaA, ideaB, models.RelatedTo)
	createTestRelationship(t, repo, ideaB, ideaC, models.DependsOn)

	// Path from A to C should go through B
	paths, err := repo.FindRelationshipPath(ideaA, ideaC, 5)

	require.NoError(t, err)
	assert.Len(t, paths, 1)
	assert.Len(t, paths[0], 2)

	// Test reverse: C to A (should work because BFS traverses bidirectionally)
	paths, err = repo.FindRelationshipPath(ideaC, ideaA, 5)

	require.NoError(t, err)
	// The BFS should be able to traverse back through the relationships
	assert.NotEmpty(t, paths, "should find reverse path through bidirectional traversal")
}

// TestFindRelationshipPath_MixedTypes tests paths with mixed relationship types
func TestFindRelationshipPath_MixedTypes(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create chain with different relationship types
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")
	ideaD := createTestIdea(t, repo, "Idea D")

	// Mixed path: A depends_on B, B part_of C, C blocks D
	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)
	createTestRelationship(t, repo, ideaB, ideaC, models.PartOf)
	createTestRelationship(t, repo, ideaC, ideaD, models.Blocks)

	paths, err := repo.FindRelationshipPath(ideaA, ideaD, 5)

	require.NoError(t, err)
	assert.Len(t, paths, 1)
	assert.Len(t, paths[0], 3)

	// Verify relationship types are preserved
	assert.Equal(t, models.DependsOn, paths[0][0].RelationshipType)
	assert.Equal(t, models.PartOf, paths[0][1].RelationshipType)
	assert.Equal(t, models.Blocks, paths[0][2].RelationshipType)
}

// TestFindRelationshipPath_ComplexGraph tests path-finding in a complex graph
// with multiple paths of different lengths
/*
Create complex graph:
       → B → C →
     /           \
   A → D → E → F → G
     \           /
       → H → I →

Three paths from A to G with different lengths:
1. A → D → E → F → G (4 hops)
2. A → B → C → G (3 hops)
3. A → H → I → G (3 hops)
*/
func TestFindRelationshipPath_ComplexGraph(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	ideas := make(map[string]string)
	for _, name := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"} {
		ideas[name] = createTestIdea(t, repo, "Idea "+name)
	}

	// Path 1: A → D → E → F → G
	createTestRelationship(t, repo, ideas["A"], ideas["D"], models.DependsOn)
	createTestRelationship(t, repo, ideas["D"], ideas["E"], models.DependsOn)
	createTestRelationship(t, repo, ideas["E"], ideas["F"], models.DependsOn)
	createTestRelationship(t, repo, ideas["F"], ideas["G"], models.DependsOn)

	// Path 2: A → B → C → G
	createTestRelationship(t, repo, ideas["A"], ideas["B"], models.DependsOn)
	createTestRelationship(t, repo, ideas["B"], ideas["C"], models.DependsOn)
	createTestRelationship(t, repo, ideas["C"], ideas["G"], models.DependsOn)

	// Path 3: A → H → I → G
	createTestRelationship(t, repo, ideas["A"], ideas["H"], models.DependsOn)
	createTestRelationship(t, repo, ideas["H"], ideas["I"], models.DependsOn)
	createTestRelationship(t, repo, ideas["I"], ideas["G"], models.DependsOn)

	// Find all paths
	paths, err := repo.FindRelationshipPath(ideas["A"], ideas["G"], 10)

	require.NoError(t, err)
	assert.Len(t, paths, 3, "should find all three paths")

	// Verify path lengths
	pathLengths := make(map[int]int)
	for _, path := range paths {
		pathLengths[len(path)]++
	}

	assert.Equal(t, 1, pathLengths[4], "should have 1 path with 4 hops")
	assert.Equal(t, 2, pathLengths[3], "should have 2 paths with 3 hops")
}

// TestFindRelationshipPath_InvalidInputs tests error handling for invalid inputs
func TestFindRelationshipPath_InvalidInputs(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	validID := createTestIdea(t, repo, "Valid Idea")

	tests := []struct {
		name      string
		sourceID  string
		targetID  string
		maxDepth  int
		wantError bool
	}{
		{
			name:      "non-existent source",
			sourceID:  "non-existent-id",
			targetID:  validID,
			maxDepth:  3,
			wantError: true,
		},
		{
			name:      "non-existent target",
			sourceID:  validID,
			targetID:  "non-existent-id",
			maxDepth:  3,
			wantError: true,
		},
		{
			name:      "empty source",
			sourceID:  "",
			targetID:  validID,
			maxDepth:  3,
			wantError: true,
		},
		{
			name:      "empty target",
			sourceID:  validID,
			targetID:  "",
			maxDepth:  3,
			wantError: true,
		},
		{
			name:      "zero max depth",
			sourceID:  validID,
			targetID:  validID,
			maxDepth:  0,
			wantError: false, // Will use default depth
		},
		{
			name:      "negative max depth",
			sourceID:  validID,
			targetID:  validID,
			maxDepth:  -1,
			wantError: false, // Will use default depth
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := repo.FindRelationshipPath(tt.sourceID, tt.targetID, tt.maxDepth)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// For same source and target, should return empty paths
				if tt.sourceID == tt.targetID {
					assert.Empty(t, paths)
				}
			}
		})
	}
}

// TestFindRelationshipPath_SameSourceAndTarget tests finding path from an idea to itself
func TestFindRelationshipPath_SameSourceAndTarget(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	ideaA := createTestIdea(t, repo, "Idea A")

	// Find path from A to A (should return empty)
	paths, err := repo.FindRelationshipPath(ideaA, ideaA, 3)

	require.NoError(t, err)
	assert.Empty(t, paths, "path from idea to itself should be empty")
}

// TestFindRelationshipPath_Performance tests performance with a moderately large graph
func TestFindRelationshipPath_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create a linear chain graph (30 ideas, 29 relationships)
	// This avoids exponential path explosion while still testing performance
	const numIdeas = 30
	ideas := make([]string, numIdeas)

	for i := 0; i < numIdeas; i++ {
		ideas[i] = createTestIdea(t, repo, fmt.Sprintf("Idea %d", i))
	}

	// Create a simple chain (avoids path explosion from multiple paths)
	for i := 0; i < numIdeas-1; i++ {
		createTestRelationship(t, repo, ideas[i], ideas[i+1], models.DependsOn)
	}

	// Measure time to find paths
	start := time.Now()
	paths, err := repo.FindRelationshipPath(ideas[0], ideas[29], 30)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, paths, 1, "should find exactly 1 path in linear chain")
	assert.Len(t, paths[0], 29, "path should have 29 hops")

	// Should complete in reasonable time (< 500ms for linear chain)
	assert.Less(t, duration, 500*time.Millisecond,
		"path-finding should be fast for linear graphs")

	t.Logf("Found %d paths in %v", len(paths), duration)
}

// TestFindRelationshipPath_ShortestPathFirst tests that BFS returns shortest paths first
func TestFindRelationshipPath_ShortestPathFirst(t *testing.T) {
	repo, cleanup := setupPathTestDB(t)
	defer cleanup()

	// Create graph with one short path and one long path
	ideaA := createTestIdea(t, repo, "Idea A")
	ideaB := createTestIdea(t, repo, "Idea B")
	ideaC := createTestIdea(t, repo, "Idea C")
	ideaD := createTestIdea(t, repo, "Idea D")
	ideaE := createTestIdea(t, repo, "Idea E")

	// Short path: A → B (1 hop)
	createTestRelationship(t, repo, ideaA, ideaB, models.DependsOn)

	// Long path: A → C → D → E → B (4 hops)
	createTestRelationship(t, repo, ideaA, ideaC, models.DependsOn)
	createTestRelationship(t, repo, ideaC, ideaD, models.DependsOn)
	createTestRelationship(t, repo, ideaD, ideaE, models.DependsOn)
	createTestRelationship(t, repo, ideaE, ideaB, models.DependsOn)

	paths, err := repo.FindRelationshipPath(ideaA, ideaB, 10)

	require.NoError(t, err)
	assert.Len(t, paths, 2, "should find both paths")

	// BFS should return shortest path first
	assert.Len(t, paths[0], 1, "first path should be shortest (1 hop)")
	assert.Len(t, paths[1], 4, "second path should be longer (4 hops)")
}

// Benchmark tests

// BenchmarkFindRelationshipPath_Simple benchmarks simple path finding
func BenchmarkFindRelationshipPath_Simple(b *testing.B) {
	// Setup
	tmpfile, _ := os.CreateTemp("", "bench_*.db")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	repo, _ := database.NewRepository(tmpfile.Name())
	defer repo.Close()

	// Create simple chain
	ideaA := createBenchIdea(b, repo, "A")
	ideaB := createBenchIdea(b, repo, "B")
	ideaC := createBenchIdea(b, repo, "C")

	createBenchRelationship(b, repo, ideaA, ideaB, models.DependsOn)
	createBenchRelationship(b, repo, ideaB, ideaC, models.DependsOn)

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindRelationshipPath(ideaA, ideaC, 5)
	}
}

// BenchmarkFindRelationshipPath_Complex benchmarks complex graph path finding
func BenchmarkFindRelationshipPath_Complex(b *testing.B) {
	// Setup
	tmpfile, _ := os.CreateTemp("", "bench_*.db")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	repo, _ := database.NewRepository(tmpfile.Name())
	defer repo.Close()

	// Create complex graph with 50 ideas
	const numIdeas = 50
	ideas := make([]string, numIdeas)

	for i := 0; i < numIdeas; i++ {
		ideas[i] = createBenchIdea(b, repo, fmt.Sprintf("Idea %d", i))
	}

	// Create relationships
	for i := 0; i < numIdeas-2; i++ {
		createBenchRelationship(b, repo, ideas[i], ideas[i+1], models.DependsOn)
		createBenchRelationship(b, repo, ideas[i], ideas[i+2], models.RelatedTo)
	}

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindRelationshipPath(ideas[0], ideas[49], 15)
	}
}

// BenchmarkFindRelationshipPath_MultiplePaths benchmarks finding multiple paths
func BenchmarkFindRelationshipPath_MultiplePaths(b *testing.B) {
	// Setup
	tmpfile, _ := os.CreateTemp("", "bench_*.db")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	repo, _ := database.NewRepository(tmpfile.Name())
	defer repo.Close()

	// Create diamond structure with multiple parallel paths
	ideaA := createBenchIdea(b, repo, "A")
	ideaB1 := createBenchIdea(b, repo, "B1")
	ideaB2 := createBenchIdea(b, repo, "B2")
	ideaB3 := createBenchIdea(b, repo, "B3")
	ideaC := createBenchIdea(b, repo, "C")

	// Create 3 parallel paths
	createBenchRelationship(b, repo, ideaA, ideaB1, models.DependsOn)
	createBenchRelationship(b, repo, ideaA, ideaB2, models.DependsOn)
	createBenchRelationship(b, repo, ideaA, ideaB3, models.DependsOn)
	createBenchRelationship(b, repo, ideaB1, ideaC, models.DependsOn)
	createBenchRelationship(b, repo, ideaB2, ideaC, models.DependsOn)
	createBenchRelationship(b, repo, ideaB3, ideaC, models.DependsOn)

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindRelationshipPath(ideaA, ideaC, 5)
	}
}

// Helper functions for benchmarks

func createBenchIdea(b *testing.B, repo *database.Repository, content string) string {
	b.Helper()
	idea := models.NewIdea(content)
	_ = repo.Create(idea)
	return idea.ID
}

func createBenchRelationship(b *testing.B, repo *database.Repository,
	sourceID, targetID string, relType models.RelationshipType) {
	b.Helper()
	rel, _ := models.NewIdeaRelationship(sourceID, targetID, relType)
	_ = repo.CreateRelationship(rel)
}
