package docker

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/trufflesecurity/trufflehog/v3/pkg/cache/memory"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// LayerScanResult represents the cached result of scanning a layer
type LayerScanResult struct {
	Chunks    []*sources.Chunk
	Error     error
	Timestamp time.Time
}

// LayerCache provides caching for Docker layer scan results
type LayerCache struct {
	cache   *memory.Cache[*LayerScanResult]
	hits    int64 // accessed atomically
	misses  int64 // accessed atomically
	mu      sync.RWMutex
	maxSize int
	enabled bool
}

// NewLayerCache creates a new layer cache with the specified maximum size
func NewLayerCache(maxSize int) *LayerCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default cache size
	}
	
	return &LayerCache{
		cache:   memory.New[*LayerScanResult](
			memory.WithExpirationInterval[*LayerScanResult](1 * time.Hour),
			memory.WithPurgeInterval[*LayerScanResult](1 * time.Hour),
		),
		maxSize: maxSize,
		enabled: true,
	}
}

// Get retrieves a cached scan result for the given layer digest
func (lc *LayerCache) Get(digest v1.Hash) (*LayerScanResult, bool) {
	if !lc.enabled {
		return nil, false
	}

	result, found := lc.cache.Get(digest.String())
	if found {
		atomic.AddInt64(&lc.hits, 1)
		return result, true
	}

	atomic.AddInt64(&lc.misses, 1)
	return nil, false
}

// Set stores a scan result for the given layer digest
func (lc *LayerCache) Set(digest v1.Hash, result *LayerScanResult) {
	if !lc.enabled {
		return
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Check if we need to evict old entries to stay under the size limit
	if lc.cache.Count() >= lc.maxSize {
		// Simple eviction strategy: remove the oldest entries
		// This is a basic implementation - could be improved with LRU
		keys := lc.cache.Keys()
		if len(keys) > 0 {
			lc.cache.Delete(keys[0])
		}
	}

	result.Timestamp = time.Now()
	lc.cache.Set(digest.String(), result)
}

// GetStats returns cache statistics
func (lc *LayerCache) GetStats() (hits, misses int64, size int) {
	hits = atomic.LoadInt64(&lc.hits)
	misses = atomic.LoadInt64(&lc.misses)
	size = lc.cache.Count()
	return
}

// Clear removes all cached entries
func (lc *LayerCache) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.cache.Clear()
	atomic.StoreInt64(&lc.hits, 0)
	atomic.StoreInt64(&lc.misses, 0)
}

// SetEnabled enables or disables the cache
func (lc *LayerCache) SetEnabled(enabled bool) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.enabled = enabled
}

// IsEnabled returns whether the cache is enabled
func (lc *LayerCache) IsEnabled() bool {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.enabled
}