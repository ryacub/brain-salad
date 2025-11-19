package cache

import (
	"sync"
	"testing"
	"time"
)

// TestResult is a simple test type to avoid import cycles
type TestResult struct {
	FinalScore     float64
	Recommendation string
	Provider       string
}

// Helper function to get and type assert
func getTestResult(cache *Cache, ideaContent string) (*TestResult, bool) {
	val, found := cache.Get(ideaContent)
	if !found {
		return nil, false
	}
	result, ok := val.(*TestResult)
	if !ok {
		return nil, false
	}
	return result, true
}

func TestCache_NewCache(t *testing.T) {
	cache := NewCache()
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.Size() != 0 {
		t.Errorf("NewCache() size = %d, want 0", cache.Size())
	}
	if cache.similarityThreshold != DefaultSimilarityThreshold {
		t.Errorf("NewCache() threshold = %.2f, want %.2f", cache.similarityThreshold, DefaultSimilarityThreshold)
	}
	if cache.ttl != DefaultTTL {
		t.Errorf("NewCache() ttl = %v, want %v", cache.ttl, DefaultTTL)
	}
	if cache.maxSize != DefaultMaxSize {
		t.Errorf("NewCache() maxSize = %d, want %d", cache.maxSize, DefaultMaxSize)
	}
}

func TestCache_StoreAndRetrieve(t *testing.T) {
	cache := NewCache()
	ideaContent := "build automation tool"
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	// Store the result
	cache.Store(ideaContent, result)
	if cache.Size() != 1 {
		t.Errorf("After Store(), size = %d, want 1", cache.Size())
	}

	// Retrieve exact match
	retrieved, found := getTestResult(cache, ideaContent)
	if !found {
		t.Fatal("Get() didn't find stored entry")
	}
	if retrieved.FinalScore != result.FinalScore {
		t.Errorf("Retrieved FinalScore = %.2f, want %.2f", retrieved.FinalScore, result.FinalScore)
	}
	if retrieved.Recommendation != result.Recommendation {
		t.Errorf("Retrieved Recommendation = %q, want %q", retrieved.Recommendation, result.Recommendation)
	}
}

func TestCache_SimilarityMatching_AboveThreshold(t *testing.T) {
	cache := NewCache()
	originalIdea := "build automation tool"
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great automation idea!",
		Provider:       "test",
	}

	cache.Store(originalIdea, result)

	// Similar idea should match (same tokens, just reordered)
	similarIdea := "automation tool build"
	retrieved, found := getTestResult(cache, similarIdea)
	if !found {
		t.Error("Get() didn't find similar entry above threshold")
	}
	if retrieved.FinalScore != result.FinalScore {
		t.Errorf("Retrieved FinalScore = %.2f, want %.2f", retrieved.FinalScore, result.FinalScore)
	}
}

func TestCache_SimilarityMatching_BelowThreshold(t *testing.T) {
	cache := NewCache()
	originalIdea := "build automation tool"
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	cache.Store(originalIdea, result)

	// Very different idea should not match
	differentIdea := "create web dashboard"
	retrieved, found := getTestResult(cache, differentIdea)
	if found {
		t.Error("Get() found entry for very different idea (should be below threshold)")
	}
	if retrieved != nil {
		t.Error("Get() returned non-nil result for cache miss")
	}
}

func TestCache_TTL_Expiration(t *testing.T) {
	cache := NewCache()
	cache.ttl = 100 * time.Millisecond // Short TTL for testing

	ideaContent := "build automation tool"
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	cache.Store(ideaContent, result)

	// Should find immediately
	_, found := getTestResult(cache, ideaContent)
	if !found {
		t.Fatal("Get() didn't find just-stored entry")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not find after TTL
	_, found = cache.Get(ideaContent)
	if found {
		t.Error("Get() found expired entry")
	}
}

func TestCache_MaxSize_LRU_Eviction(t *testing.T) {
	cache := NewCache()
	cache.maxSize = 3 // Small max size for testing

	results := []*TestResult{
		{FinalScore: 1.0, Recommendation: "First", Provider: "test"},
		{FinalScore: 2.0, Recommendation: "Second", Provider: "test"},
		{FinalScore: 3.0, Recommendation: "Third", Provider: "test"},
		{FinalScore: 4.0, Recommendation: "Fourth", Provider: "test"},
	}

	// Store 3 entries (fill to max) - use distinct content to avoid similarity matching
	cache.Store("build automation tool", results[0])
	cache.Store("create web dashboard", results[1])
	cache.Store("implement user authentication", results[2])

	if cache.Size() != 3 {
		t.Fatalf("After storing 3 entries, size = %d, want 3", cache.Size())
	}

	// Store 4th entry (should evict oldest)
	cache.Store("design database schema", results[3])

	if cache.Size() != 3 {
		t.Errorf("After storing 4th entry, size = %d, want 3", cache.Size())
	}

	// First entry should be evicted
	_, found := getTestResult(cache, "build automation tool")
	if found {
		t.Error("Get() found evicted entry (build automation tool)")
	}

	// Other entries should still be there
	ideas := []string{
		"create web dashboard",
		"implement user authentication",
		"design database schema",
	}
	for _, idea := range ideas {
		_, found := getTestResult(cache, idea)
		if !found {
			t.Errorf("Get(%q) didn't find entry after LRU eviction", idea)
		}
	}
}

func TestCache_HitCount_Tracking(t *testing.T) {
	cache := NewCache()
	ideaContent := "build automation tool"
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	cache.Store(ideaContent, result)

	// Get the entry multiple times
	for i := 0; i < 5; i++ {
		cache.Get(ideaContent)
	}

	// Check hit count in cache entry
	cache.mu.RLock()
	normalized := NormalizeText(ideaContent)
	entry, exists := cache.entries[normalized]
	cache.mu.RUnlock()

	if !exists {
		t.Fatal("Cache entry not found")
	}
	if entry.HitCount != 5 {
		t.Errorf("HitCount = %d, want 5", entry.HitCount)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := NewCache()
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Store("idea concurrent test", result)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Get("idea concurrent test")
			}
		}(i)
	}

	wg.Wait()

	// If we got here without deadlock or race conditions, test passes
	if cache.Size() == 0 {
		t.Error("Cache is empty after concurrent operations")
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache()
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	cache.Store("idea 1", result)
	cache.Store("idea 2", result)

	// 3 hits
	cache.Get("idea 1")
	cache.Get("idea 1")
	cache.Get("idea 2")

	// 2 misses
	cache.Get("nonexistent 1")
	cache.Get("nonexistent 2")

	stats := cache.GetStats()

	if stats.Size != 2 {
		t.Errorf("Stats.Size = %d, want 2", stats.Size)
	}
	if stats.Hits != 3 {
		t.Errorf("Stats.Hits = %d, want 3", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Stats.Misses = %d, want 2", stats.Misses)
	}

	expectedHitRate := 3.0 / 5.0 // 3 hits out of 5 total requests
	if stats.HitRate < expectedHitRate-0.01 || stats.HitRate > expectedHitRate+0.01 {
		t.Errorf("Stats.HitRate = %.2f, want %.2f", stats.HitRate, expectedHitRate)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache()
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	cache.Store("idea 1", result)
	cache.Store("idea 2", result)
	cache.Get("idea 1")

	if cache.Size() != 2 {
		t.Fatalf("Before Clear(), size = %d, want 2", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("After Clear(), size = %d, want 0", cache.Size())
	}

	stats := cache.GetStats()
	if stats.Hits != 0 {
		t.Errorf("After Clear(), Hits = %d, want 0", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("After Clear(), Misses = %d, want 0", stats.Misses)
	}
}

// Benchmark tests
func BenchmarkCache_Get_ExactMatch(b *testing.B) {
	cache := NewCache()
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}
	cache.Store("build automation tool", result)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("build automation tool")
	}
}

func BenchmarkCache_Get_SimilarityMatch(b *testing.B) {
	cache := NewCache()
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}
	cache.Store("build automation tool", result)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("automation build tool")
	}
}

func BenchmarkCache_Store(b *testing.B) {
	cache := NewCache()
	result := &TestResult{
		FinalScore:     8.5,
		Recommendation: "Great idea!",
		Provider:       "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Store("build automation tool", result)
	}
}

func BenchmarkJaccardSimilarity(b *testing.B) {
	text1 := "build automation tool for continuous integration"
	text2 := "create automation system for continuous deployment"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JaccardSimilarity(text1, text2)
	}
}
