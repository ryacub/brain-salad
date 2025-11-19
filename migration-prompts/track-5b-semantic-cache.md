# Track 5B: Semantic Cache System

**Phase**: 5 - LLM Integration
**Estimated Time**: 10-12 hours
**Dependencies**: 5A (needs types.go for llm.AnalysisResult)
**Can Run in Parallel**: Yes (after 5A creates types.go)

---

## Mission

You are implementing a semantic similarity-based cache system for LLM responses in the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has semantic caching in `src/llm_cache.rs`
- Cache should match similar ideas using Jaccard similarity (>0.85 threshold)
- 24-hour TTL, max 1000 entries per type, LRU eviction
- Cache hits should return in <5ms

## Reference Implementation

Review `/home/user/brain-salad/src/llm_cache.rs` for:
- Similarity threshold (0.85)
- TTL (24 hours)
- Max cache size (1000)
- Hit count tracking
- Normalization logic

## Your Task

Implement semantic cache system using strict TDD methodology.

**IMPORTANT**: Wait for 5A to complete `types.go` (should be done within 2 hours of 5A starting). Once types are available, you can proceed in parallel.

## Directory Structure

Create files in `go/internal/llm/cache/`:
- `cache.go` - Core cache implementation
- `similarity.go` - Text similarity algorithms
- `stats.go` - Cache statistics
- `cache_test.go` - Comprehensive tests
- `similarity_test.go` - Similarity algorithm tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/llm/cache/cache_test.go`:
- `TestCache_NewCache()`
- `TestCache_StoreAndRetrieve()`
- `TestCache_SimilarityMatching_AboveThreshold()`
- `TestCache_SimilarityMatching_BelowThreshold()`
- `TestCache_TTL_Expiration()`
- `TestCache_MaxSize_LRU_Eviction()`
- `TestCache_HitCount_Tracking()`
- `TestCache_ConcurrentAccess()`
- `TestCache_Stats()`

Create `go/internal/llm/cache/similarity_test.go`:
- `TestJaccardSimilarity()`
- `TestNormalizeText()`
- `TestTokenize()`

Run: `go test ./internal/llm/cache -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/llm/cache/similarity.go`:

```go
package cache

import (
    "regexp"
    "strings"
)

var (
    // Common English stopwords to filter out
    stopwords = map[string]bool{
        "a": true, "an": true, "and": true, "are": true, "as": true,
        "at": true, "be": true, "by": true, "for": true, "from": true,
        "has": true, "he": true, "in": true, "is": true, "it": true,
        "its": true, "of": true, "on": true, "that": true, "the": true,
        "to": true, "was": true, "will": true, "with": true,
    }

    nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s]+`)
    multipleSpaces  = regexp.MustCompile(`\s+`)
)

// NormalizeText canonicalizes text for similarity comparison
func NormalizeText(text string) string {
    text = strings.ToLower(text)
    text = nonAlphanumeric.ReplaceAllString(text, " ")
    text = multipleSpaces.ReplaceAllString(text, " ")
    text = strings.TrimSpace(text)
    return text
}

// Tokenize splits text into words and removes stopwords
func Tokenize(text string) []string {
    normalized := NormalizeText(text)
    words := strings.Fields(normalized)

    filtered := make([]string, 0, len(words))
    for _, word := range words {
        if !stopwords[word] && len(word) > 1 {
            filtered = append(filtered, word)
        }
    }
    return filtered
}

// JaccardSimilarity computes Jaccard similarity between two texts
func JaccardSimilarity(text1, text2 string) float64 {
    tokens1 := Tokenize(text1)
    tokens2 := Tokenize(text2)

    if len(tokens1) == 0 && len(tokens2) == 0 {
        return 1.0
    }
    if len(tokens1) == 0 || len(tokens2) == 0 {
        return 0.0
    }

    set1 := make(map[string]bool)
    for _, token := range tokens1 {
        set1[token] = true
    }

    set2 := make(map[string]bool)
    for _, token := range tokens2 {
        set2[token] = true
    }

    intersection := 0
    for token := range set1 {
        if set2[token] {
            intersection++
        }
    }

    union := len(set1) + len(set2) - intersection
    if union == 0 {
        return 0.0
    }

    return float64(intersection) / float64(union)
}
```

#### B. Implement `go/internal/llm/cache/cache.go`:

```go
package cache

import (
    "container/list"
    "sync"
    "time"

    "github.com/rayyacub/telos-idea-matrix/internal/llm"
)

const (
    DefaultSimilarityThreshold = 0.85
    DefaultTTL                 = 24 * time.Hour
    DefaultMaxSize             = 1000
)

type CacheEntry struct {
    Key            string
    NormalizedText string
    Result         *llm.AnalysisResult
    CachedAt       time.Time
    HitCount       int64
    LastSimilarity float64
    element        *list.Element
}

type Cache struct {
    entries             map[string]*CacheEntry
    lru                 *list.List
    mu                  sync.RWMutex
    similarityThreshold float64
    ttl                 time.Duration
    maxSize             int
    hits                int64
    misses              int64
}

func NewCache() *Cache {
    return &Cache{
        entries:             make(map[string]*CacheEntry),
        lru:                 list.New(),
        similarityThreshold: DefaultSimilarityThreshold,
        ttl:                 DefaultTTL,
        maxSize:             DefaultMaxSize,
    }
}

func (c *Cache) Store(ideaContent string, result *llm.AnalysisResult) {
    c.mu.Lock()
    defer c.mu.Unlock()

    normalized := NormalizeText(ideaContent)
    entry := &CacheEntry{
        Key:            normalized,
        NormalizedText: normalized,
        Result:         result,
        CachedAt:       time.Now(),
        HitCount:       0,
    }

    entry.element = c.lru.PushFront(entry)
    c.entries[normalized] = entry

    if c.lru.Len() > c.maxSize {
        c.evictOldest()
    }
}

func (c *Cache) Get(ideaContent string) (*llm.AnalysisResult, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    normalized := NormalizeText(ideaContent)

    // Try exact match first
    if entry, exists := c.entries[normalized]; exists {
        if !c.isExpired(entry) {
            entry.HitCount++
            entry.LastSimilarity = 1.0
            c.lru.MoveToFront(entry.element)
            c.hits++
            return entry.Result, true
        }
        c.removeEntry(entry)
    }

    // Try similarity match
    bestMatch := c.findSimilarEntry(normalized)
    if bestMatch != nil {
        bestMatch.HitCount++
        c.lru.MoveToFront(bestMatch.element)
        c.hits++
        return bestMatch.Result, true
    }

    c.misses++
    return nil, false
}

func (c *Cache) findSimilarEntry(normalized string) *CacheEntry {
    var bestMatch *CacheEntry
    var bestSimilarity float64

    for _, entry := range c.entries {
        if c.isExpired(entry) {
            continue
        }

        similarity := JaccardSimilarity(normalized, entry.NormalizedText)
        if similarity >= c.similarityThreshold && similarity > bestSimilarity {
            bestSimilarity = similarity
            bestMatch = entry
        }
    }

    if bestMatch != nil {
        bestMatch.LastSimilarity = bestSimilarity
    }

    return bestMatch
}

func (c *Cache) isExpired(entry *CacheEntry) bool {
    return time.Since(entry.CachedAt) > c.ttl
}

func (c *Cache) evictOldest() {
    element := c.lru.Back()
    if element != nil {
        entry := element.Value.(*CacheEntry)
        c.removeEntry(entry)
    }
}

func (c *Cache) removeEntry(entry *CacheEntry) {
    c.lru.Remove(entry.element)
    delete(c.entries, entry.Key)
}

func (c *Cache) Size() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.entries)
}

func (c *Cache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.entries = make(map[string]*CacheEntry)
    c.lru = list.New()
    c.hits = 0
    c.misses = 0
}
```

#### C. Implement `go/internal/llm/cache/stats.go`:

```go
package cache

type CacheStats struct {
    Size        int
    Hits        int64
    Misses      int64
    HitRate     float64
    AvgHitCount float64
}

func (c *Cache) GetStats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    total := c.hits + c.misses
    hitRate := 0.0
    if total > 0 {
        hitRate = float64(c.hits) / float64(total)
    }

    totalHitCount := int64(0)
    for _, entry := range c.entries {
        totalHitCount += entry.HitCount
    }

    avgHitCount := 0.0
    if len(c.entries) > 0 {
        avgHitCount = float64(totalHitCount) / float64(len(c.entries))
    }

    return CacheStats{
        Size:        len(c.entries),
        Hits:        c.hits,
        Misses:      c.misses,
        HitRate:     hitRate,
        AvgHitCount: avgHitCount,
    }
}
```

Run: `go test ./internal/llm/cache -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add cache persistence (save/load from disk)
- Optimize similarity calculation (skip low-similarity candidates early)
- Add cache warming (preload common queries)
- Extract configuration (threshold, TTL, max size)

## Integration

1. Wire into Ollama provider (check cache before calling Ollama, store after)
2. Add cache stats to metrics endpoint
3. Add `tm cache stats` CLI command
4. Add `tm cache clear` CLI command

## Success Criteria

- ✅ All tests pass with >90% coverage
- ✅ Cache hits return in <5ms
- ✅ Similarity matching accuracy >90%
- ✅ Proper LRU eviction
- ✅ Thread-safe (verified with -race flag)
- ✅ Hit rate >60% in production use

## Validation

```bash
# Unit tests
go test ./internal/llm/cache -v -cover -race

# Performance test
go test ./internal/llm/cache -v -bench=. -benchmem

# Integration test
go run ./cmd/cli/main.go analyze --ai "Build automation tool"
go run ./cmd/cli/main.go analyze --ai "Create automation tool"
# Second call should be instant (cache hit)
```

## Deliverables

- `go/internal/llm/cache/cache.go`
- `go/internal/llm/cache/similarity.go`
- `go/internal/llm/cache/stats.go`
- `go/internal/llm/cache/cache_test.go`
- `go/internal/llm/cache/similarity_test.go`
- Integration into Ollama provider
