// Package cache provides LRU caching with similarity-based matching for LLM analysis results.
package cache

import (
	"container/list"
	"sync"
	"time"
)

const (
	DefaultSimilarityThreshold = 0.85
	DefaultTTL                 = 24 * time.Hour
	DefaultMaxSize             = 1000
)

type CacheEntry struct {
	Key            string
	NormalizedText string
	Result         any // Generic result type to avoid import cycles
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

func (c *Cache) Store(ideaContent string, result any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	normalized := NormalizeText(ideaContent)

	// If entry already exists, remove it from LRU list first
	if existingEntry, exists := c.entries[normalized]; exists {
		c.lru.Remove(existingEntry.element)
	}

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

func (c *Cache) Get(ideaContent string) (any, bool) {
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
