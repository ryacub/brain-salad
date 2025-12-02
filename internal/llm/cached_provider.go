package llm

import (
	"github.com/ryacub/telos-idea-matrix/internal/llm/cache"
)

// CachedProvider wraps a Provider with caching capabilities.
// It checks the cache before calling the underlying provider and stores results after.
type CachedProvider struct {
	provider Provider
	cache    *cache.Cache
}

// NewCachedProvider creates a new cached provider that wraps the given provider.
func NewCachedProvider(provider Provider) *CachedProvider {
	return &CachedProvider{
		provider: provider,
		cache:    cache.NewCache(),
	}
}

// Name returns the provider name with cache indicator.
func (cp *CachedProvider) Name() string {
	return cp.provider.Name() + "_cached"
}

// IsAvailable delegates to the underlying provider.
func (cp *CachedProvider) IsAvailable() bool {
	return cp.provider.IsAvailable()
}

// Analyze checks the cache before calling the underlying provider.
// If a cache hit occurs, returns the cached result with FromCache=true.
// Otherwise, calls the provider, stores the result in cache, and returns it.
func (cp *CachedProvider) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	// Try to get from cache first
	if cachedValue, found := cp.cache.Get(req.IdeaContent); found {
		// Type assert to AnalysisResult
		if result, ok := cachedValue.(*AnalysisResult); ok {
			// Mark as coming from cache
			cachedResult := *result
			cachedResult.FromCache = true
			return &cachedResult, nil
		}
	}

	// Cache miss - call underlying provider
	result, err := cp.provider.Analyze(req)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	cp.cache.Store(req.IdeaContent, result)

	return result, nil
}

// GetCache returns the underlying cache for statistics and management.
func (cp *CachedProvider) GetCache() *cache.Cache {
	return cp.cache
}

// GetCacheStats returns cache statistics.
func (cp *CachedProvider) GetCacheStats() cache.CacheStats {
	return cp.cache.GetStats()
}

// ClearCache clears all cached entries.
func (cp *CachedProvider) ClearCache() {
	cp.cache.Clear()
}
