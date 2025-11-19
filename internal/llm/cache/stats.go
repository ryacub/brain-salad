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
