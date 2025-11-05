// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tiered

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/lru"
	"github.com/goharbor/harbor/src/lib/log"
)

// TieredCache implements a two-tier caching strategy for Harbor metadata
// L1: In-memory LRU cache for hot metadata (manifests, project info, repository metadata)
// L2: Redis cache for shared metadata across Harbor instances
type TieredCache struct {
	l1     cache.Cache // In-memory LRU cache (fast, small, per-instance)
	l2     cache.Cache // Redis cache (shared, larger, cross-instance)
	config *Config

	// Statistics
	l1Hits    atomic.Uint64
	l2Hits    atomic.Uint64
	misses    atomic.Uint64
	totalReqs atomic.Uint64
}

// Config holds configuration for tiered cache
type Config struct {
	// L1 (Memory) Configuration - for hot metadata
	L1MaxSize    int64         // Max memory size in bytes (default: 100MB)
	L1MaxEntries int           // Max number of entries (default: 10000)
	L1DefaultTTL time.Duration // Default TTL for L1 (default: 2min)

	// L2 (Redis) Configuration - for shared metadata
	L2DefaultTTL time.Duration // Default TTL for L2 (default: 10min)
	L2Address    string        // Redis address

	// Behavior
	EnableL1     bool // Enable L1 cache (default: true)
	EnableL2     bool // Enable L2 cache (default: true)
	PromoteOnHit bool // Promote L2 hits to L1 (default: true)

	// Content-specific TTLs for different metadata types
	ManifestByDigestTTL time.Duration // Manifests by digest are immutable (default: 24h)
	ManifestByTagTTL    time.Duration // Manifests by tag can change (default: 5min)
	ProjectMetaTTL      time.Duration // Project metadata (default: 10min)
	RepositoryMetaTTL   time.Duration // Repository metadata (default: 10min)
	ArtifactMetaTTL     time.Duration // Artifact metadata (default: 15min)
	TagMetaTTL          time.Duration // Tag metadata (default: 5min)
	QueryResultTTL      time.Duration // Database query results (default: 1min)
}

// DefaultConfig returns default configuration optimized for metadata caching
func DefaultConfig() *Config {
	return &Config{
		L1MaxSize:           100 * 1024 * 1024, // 100MB - enough for ~10k manifests
		L1MaxEntries:        10000,             // ~10k metadata entries
		L1DefaultTTL:        2 * time.Minute,
		L2DefaultTTL:        10 * time.Minute,
		EnableL1:            true,
		EnableL2:            true,
		PromoteOnHit:        true,
		ManifestByDigestTTL: 24 * time.Hour,  // Immutable
		ManifestByTagTTL:    5 * time.Minute, // Mutable
		ProjectMetaTTL:      10 * time.Minute,
		RepositoryMetaTTL:   10 * time.Minute,
		ArtifactMetaTTL:     15 * time.Minute,
		TagMetaTTL:          5 * time.Minute,
		QueryResultTTL:      1 * time.Minute,
	}
}

// NewTieredCache creates a new tiered cache for Harbor metadata
func NewTieredCache(config *Config) (*TieredCache, error) {
	if config == nil {
		config = DefaultConfig()
	}

	tc := &TieredCache{
		config: config,
	}

	// Initialize L1 (in-memory LRU cache)
	if config.EnableL1 {
		l1, err := lru.New(lru.Config{
			MaxSize:    config.L1MaxSize,
			MaxEntries: config.L1MaxEntries,
			DefaultTTL: config.L1DefaultTTL,
			OnEvict: func(key string, reason lru.EvictionReason) {
				log.Debugf("L1 cache eviction: key=%s, reason=%s", key, reason)
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create L1 cache: %w", err)
		}
		tc.l1 = l1
		log.Infof("Tiered cache L1 initialized: maxSize=%dMB, maxEntries=%d",
			config.L1MaxSize/(1024*1024), config.L1MaxEntries)
	}

	// Initialize L2 (Redis cache)
	if config.EnableL2 && config.L2Address != "" {
		l2, err := cache.New(cache.Redis,
			cache.Address(config.L2Address),
			cache.Prefix("harbor:metadata:"),
		)
		if err != nil {
			log.Warningf("Failed to create L2 cache, will use L1 only: %v", err)
			config.EnableL2 = false
		} else {
			tc.l2 = l2
			log.Infof("Tiered cache L2 initialized: address=%s", config.L2Address)
		}
	}

	// Start metrics collection
	go tc.collectMetrics()

	return tc, nil
}

// Fetch retrieves metadata from cache (tries L1, then L2, then returns ErrNotFound)
func (tc *TieredCache) Fetch(ctx context.Context, key string, value any) error {
	tc.totalReqs.Add(1)

	// Try L1 first (in-memory, fastest)
	if tc.config.EnableL1 && tc.l1 != nil {
		err := tc.l1.Fetch(ctx, key, value)
		if err == nil {
			tc.l1Hits.Add(1)
			log.Debugf("L1 cache hit: %s", key)
			return nil
		}
		if err != cache.ErrNotFound {
			log.Errorf("L1 cache error for key %s: %v", key, err)
		}
	}

	// Try L2 on L1 miss (Redis, shared across instances)
	if tc.config.EnableL2 && tc.l2 != nil {
		err := tc.l2.Fetch(ctx, key, value)
		if err == nil {
			tc.l2Hits.Add(1)
			log.Debugf("L2 cache hit: %s", key)

			// Promote to L1 if enabled (async to avoid blocking)
			if tc.config.PromoteOnHit && tc.l1 != nil {
				go func() {
					if err := tc.l1.Save(context.Background(), key, value, tc.config.L1DefaultTTL); err != nil {
						log.Debugf("Failed to promote key %s to L1: %v", key, err)
					}
				}()
			}
			return nil
		}
		if err != cache.ErrNotFound {
			log.Errorf("L2 cache error for key %s: %v", key, err)
		}
	}

	// Cache miss - caller should fetch from database
	tc.misses.Add(1)
	return cache.ErrNotFound
}

// Save stores metadata in cache (writes to L1 and L2)
func (tc *TieredCache) Save(ctx context.Context, key string, value any, expiration ...time.Duration) error {
	ttl := tc.config.L1DefaultTTL
	if len(expiration) > 0 {
		ttl = expiration[0]
	}

	// Save to L1 (synchronous for immediate availability)
	if tc.config.EnableL1 && tc.l1 != nil {
		if err := tc.l1.Save(ctx, key, value, ttl); err != nil {
			log.Errorf("Failed to save to L1 cache: %v", err)
		}
	}

	// Save to L2 (async to avoid blocking)
	if tc.config.EnableL2 && tc.l2 != nil {
		l2TTL := tc.config.L2DefaultTTL
		if len(expiration) > 0 && expiration[0] > ttl {
			l2TTL = expiration[0]
		}

		go func() {
			if err := tc.l2.Save(context.Background(), key, value, l2TTL); err != nil {
				log.Errorf("Failed to save to L2 cache: %v", err)
			}
		}()
	}

	return nil
}

// Delete removes key from all cache tiers (for cache invalidation)
func (tc *TieredCache) Delete(ctx context.Context, key string) error {
	var errs []error

	if tc.config.EnableL1 && tc.l1 != nil {
		if err := tc.l1.Delete(ctx, key); err != nil {
			errs = append(errs, fmt.Errorf("L1 delete error: %w", err))
		}
	}

	if tc.config.EnableL2 && tc.l2 != nil {
		if err := tc.l2.Delete(ctx, key); err != nil {
			errs = append(errs, fmt.Errorf("L2 delete error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cache delete errors: %v", errs)
	}
	return nil
}

// Contains checks if key exists in any tier
func (tc *TieredCache) Contains(ctx context.Context, key string) bool {
	if tc.config.EnableL1 && tc.l1 != nil && tc.l1.Contains(ctx, key) {
		return true
	}
	if tc.config.EnableL2 && tc.l2 != nil && tc.l2.Contains(ctx, key) {
		return true
	}
	return false
}

// Ping checks health of all cache tiers
func (tc *TieredCache) Ping(ctx context.Context) error {
	if tc.config.EnableL1 && tc.l1 != nil {
		if err := tc.l1.Ping(ctx); err != nil {
			return fmt.Errorf("L1 ping failed: %w", err)
		}
	}
	if tc.config.EnableL2 && tc.l2 != nil {
		if err := tc.l2.Ping(ctx); err != nil {
			return fmt.Errorf("L2 ping failed: %w", err)
		}
	}
	return nil
}

// Scan scans keys (delegates to L2 for efficiency, as it's shared)
func (tc *TieredCache) Scan(ctx context.Context, match string) (cache.Iterator, error) {
	if tc.config.EnableL2 && tc.l2 != nil {
		return tc.l2.Scan(ctx, match)
	}
	if tc.config.EnableL1 && tc.l1 != nil {
		return tc.l1.Scan(ctx, match)
	}
	return nil, fmt.Errorf("no cache tier available for scan")
}

// Stats returns cache statistics
func (tc *TieredCache) Stats() CacheStats {
	total := tc.totalReqs.Load()
	l1Hits := tc.l1Hits.Load()
	l2Hits := tc.l2Hits.Load()
	misses := tc.misses.Load()

	var l1HitRate, l2HitRate, overallHitRate float64
	if total > 0 {
		l1HitRate = float64(l1Hits) / float64(total) * 100
		l2HitRate = float64(l2Hits) / float64(total) * 100
		overallHitRate = float64(l1Hits+l2Hits) / float64(total) * 100
	}

	return CacheStats{
		TotalRequests:  total,
		L1Hits:         l1Hits,
		L2Hits:         l2Hits,
		Misses:         misses,
		L1HitRate:      l1HitRate,
		L2HitRate:      l2HitRate,
		OverallHitRate: overallHitRate,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	TotalRequests  uint64
	L1Hits         uint64
	L2Hits         uint64
	Misses         uint64
	L1HitRate      float64
	L2HitRate      float64
	OverallHitRate float64
}

// collectMetrics periodically collects and reports metrics
func (tc *TieredCache) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := tc.Stats()

		if stats.TotalRequests > 0 {
			log.Infof("Tiered cache stats: total=%d, L1=%.2f%%, L2=%.2f%%, overall=%.2f%%, misses=%d",
				stats.TotalRequests, stats.L1HitRate, stats.L2HitRate, stats.OverallHitRate, stats.Misses)
		}
	}
}

// GetTTLForContentType returns appropriate TTL based on content type
func (tc *TieredCache) GetTTLForContentType(contentType ContentType, isImmutable bool) time.Duration {
	switch contentType {
	case ContentTypeManifest:
		if isImmutable {
			return tc.config.ManifestByDigestTTL
		}
		return tc.config.ManifestByTagTTL
	case ContentTypeProject:
		return tc.config.ProjectMetaTTL
	case ContentTypeRepository:
		return tc.config.RepositoryMetaTTL
	case ContentTypeArtifact:
		return tc.config.ArtifactMetaTTL
	case ContentTypeTag:
		return tc.config.TagMetaTTL
	case ContentTypeQueryResult:
		return tc.config.QueryResultTTL
	default:
		return tc.config.L1DefaultTTL
	}
}

// ContentType represents the type of cached content
type ContentType string

const (
	ContentTypeManifest    ContentType = "manifest"
	ContentTypeProject     ContentType = "project"
	ContentTypeRepository  ContentType = "repository"
	ContentTypeArtifact    ContentType = "artifact"
	ContentTypeTag         ContentType = "tag"
	ContentTypeQueryResult ContentType = "query"
)
