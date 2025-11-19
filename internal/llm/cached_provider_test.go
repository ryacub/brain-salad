package llm

import (
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

func TestCachedProvider_CacheHit(t *testing.T) {
	mockProvider := &MockProvider{
		name:      "mock",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.5,
			Recommendation: "Great idea!",
			Provider:       "mock",
			Duration:       100 * time.Millisecond,
			FromCache:      false,
		},
	}

	cachedProvider := NewCachedProvider(mockProvider)

	req := AnalysisRequest{
		IdeaContent: "build automation tool",
		Telos:       &models.Telos{},
	}

	// First call - should hit the mock provider
	result1, err := cachedProvider.Analyze(req)
	if err != nil {
		t.Fatalf("First Analyze() failed: %v", err)
	}
	if result1.FromCache {
		t.Error("First result should not be from cache")
	}
	if mockProvider.callCount != 1 {
		t.Errorf("Mock provider should be called once, got %d", mockProvider.callCount)
	}

	// Second call - should hit the cache
	result2, err := cachedProvider.Analyze(req)
	if err != nil {
		t.Fatalf("Second Analyze() failed: %v", err)
	}
	if !result2.FromCache {
		t.Error("Second result should be from cache")
	}
	if mockProvider.callCount != 1 {
		t.Errorf("Mock provider should still be called once, got %d", mockProvider.callCount)
	}

	// Verify results are the same
	if result2.FinalScore != result1.FinalScore {
		t.Errorf("Cached result FinalScore = %.2f, want %.2f", result2.FinalScore, result1.FinalScore)
	}
	if result2.Recommendation != result1.Recommendation {
		t.Errorf("Cached result Recommendation = %q, want %q", result2.Recommendation, result1.Recommendation)
	}
}

func TestCachedProvider_SimilarityMatch(t *testing.T) {
	mockProvider := &MockProvider{
		name:      "mock",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.5,
			Recommendation: "Great idea!",
			Provider:       "mock",
		},
	}

	cachedProvider := NewCachedProvider(mockProvider)

	// Store result for first idea
	req1 := AnalysisRequest{
		IdeaContent: "build automation tool",
		Telos:       &models.Telos{},
	}
	_, err := cachedProvider.Analyze(req1)
	if err != nil {
		t.Fatalf("First Analyze() failed: %v", err)
	}

	// Similar idea should hit cache (same tokens, reordered)
	req2 := AnalysisRequest{
		IdeaContent: "automation tool build",
		Telos:       &models.Telos{},
	}
	result2, err := cachedProvider.Analyze(req2)
	if err != nil {
		t.Fatalf("Second Analyze() failed: %v", err)
	}
	if !result2.FromCache {
		t.Error("Similar idea should hit cache")
	}
	if mockProvider.callCount != 1 {
		t.Errorf("Mock provider should be called once, got %d", mockProvider.callCount)
	}
}

func TestCachedProvider_CacheMiss(t *testing.T) {
	mockProvider := &MockProvider{
		name:      "mock",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.5,
			Recommendation: "Great idea!",
			Provider:       "mock",
		},
	}

	cachedProvider := NewCachedProvider(mockProvider)

	// Store result for first idea
	req1 := AnalysisRequest{
		IdeaContent: "build automation tool",
		Telos:       &models.Telos{},
	}
	_, err := cachedProvider.Analyze(req1)
	if err != nil {
		t.Fatalf("First Analyze() failed: %v", err)
	}

	// Very different idea should miss cache
	req2 := AnalysisRequest{
		IdeaContent: "create web dashboard",
		Telos:       &models.Telos{},
	}
	result2, err := cachedProvider.Analyze(req2)
	if err != nil {
		t.Fatalf("Second Analyze() failed: %v", err)
	}
	if result2.FromCache {
		t.Error("Different idea should not hit cache")
	}
	if mockProvider.callCount != 2 {
		t.Errorf("Mock provider should be called twice, got %d", mockProvider.callCount)
	}
}

func TestCachedProvider_GetStats(t *testing.T) {
	mockProvider := &MockProvider{
		name:      "mock",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.5,
			Recommendation: "Great idea!",
			Provider:       "mock",
		},
	}

	cachedProvider := NewCachedProvider(mockProvider)

	req := AnalysisRequest{
		IdeaContent: "build automation tool",
		Telos:       &models.Telos{},
	}

	// First call - cache miss
	_, _ = cachedProvider.Analyze(req)

	// Second call - cache hit
	_, _ = cachedProvider.Analyze(req)

	// Get stats
	stats := cachedProvider.GetCacheStats()

	if stats.Size != 1 {
		t.Errorf("Cache size = %d, want 1", stats.Size)
	}
	if stats.Hits != 1 {
		t.Errorf("Cache hits = %d, want 1", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Cache misses = %d, want 1", stats.Misses)
	}
	if stats.HitRate < 0.49 || stats.HitRate > 0.51 {
		t.Errorf("Cache hit rate = %.2f, want ~0.50", stats.HitRate)
	}
}

func TestCachedProvider_ClearCache(t *testing.T) {
	mockProvider := &MockProvider{
		name:      "mock",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.5,
			Recommendation: "Great idea!",
			Provider:       "mock",
		},
	}

	cachedProvider := NewCachedProvider(mockProvider)

	req := AnalysisRequest{
		IdeaContent: "build automation tool",
		Telos:       &models.Telos{},
	}

	// Store result
	_, _ = cachedProvider.Analyze(req)

	stats := cachedProvider.GetCacheStats()
	if stats.Size != 1 {
		t.Fatalf("Cache size before clear = %d, want 1", stats.Size)
	}

	// Clear cache
	cachedProvider.ClearCache()

	stats = cachedProvider.GetCacheStats()
	if stats.Size != 0 {
		t.Errorf("Cache size after clear = %d, want 0", stats.Size)
	}

	// Next call should miss cache
	result, _ := cachedProvider.Analyze(req)
	if result.FromCache {
		t.Error("Result should not be from cache after clear")
	}
	if mockProvider.callCount != 2 {
		t.Errorf("Mock provider should be called twice, got %d", mockProvider.callCount)
	}
}
