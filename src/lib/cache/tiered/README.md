# Tiered Cache for Harbor Metadata

## Overview

The tiered cache provides a two-tier caching strategy optimized for Harbor metadata (manifests, projects, repositories, artifacts, tags). It significantly improves performance for high-concurrency read operations by reducing database queries and Redis network calls.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client Request                        │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│          L1: In-Memory LRU Cache (Per Instance)         │
│  • Hot metadata (manifests, artifact metadata)          │
│  • Size: 100MB (configurable)                           │
│  • Entries: 10,000 (configurable)                       │
│  • TTL: 2 minutes (content-dependent)                   │
│  • Target Hit Rate: 60-70%                              │
└──────────────────────┬──────────────────────────────────┘
                       │ L1 Miss
                       ▼
┌─────────────────────────────────────────────────────────┐
│          L2: Redis Cache (Shared Across Instances)      │
│  • Warm metadata (shared across all Harbor cores)       │
│  • Size: Unlimited (Redis capacity)                     │
│  • TTL: 10 minutes (content-dependent)                  │
│  • Target Hit Rate: 25-30%                              │
└──────────────────────┬──────────────────────────────────┘
                       │ L2 Miss
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Database + Storage (Source of Truth)        │
│  • PostgreSQL for metadata                              │
│  • S3/Filesystem for blobs                              │
└─────────────────────────────────────────────────────────┘
```

## Features

### 1. **Content-Aware TTL**
Different metadata types have different TTLs based on their mutability:

| Content Type | TTL | Reason |
|--------------|-----|--------|
| Manifest (by digest) | 24 hours | Immutable, content-addressable |
| Manifest (by tag) | 5 minutes | Mutable, tags can be updated |
| Artifact metadata | 15 minutes | Semi-stable |
| Project metadata | 10 minutes | Rarely changes |
| Repository metadata | 10 minutes | Rarely changes |
| Tag metadata | 5 minutes | Can change frequently |
| Query results | 1 minute | Volatile |

### 2. **LRU Eviction**
L1 cache uses LRU (Least Recently Used) eviction with both size and count limits:
- Evicts least recently used entries when size limit is reached
- Evicts least recently used entries when entry count limit is reached
- Automatically removes expired entries

### 3. **Promotion on Hit**
When data is found in L2 (Redis), it's automatically promoted to L1 (memory) for faster subsequent access.

### 4. **Write-Through Optional**
By default, writes go to L1 immediately and L2 asynchronously. Can be configured for synchronous write-through.

### 5. **Automatic Cleanup**
Expired entries are automatically cleaned up every minute to prevent memory bloat.

## Configuration

### Environment Variables

```bash
# L1 (Memory) Configuration
HARBOR_CACHE_L1_MAX_SIZE_MB=100              # Max L1 size in MB (default: 100)
HARBOR_CACHE_L1_MAX_ENTRIES=10000            # Max L1 entries (default: 10000)
HARBOR_CACHE_L1_DEFAULT_TTL_MINUTES=2        # Default L1 TTL (default: 2)
HARBOR_CACHE_L1_ENABLED=true                 # Enable L1 cache (default: true)

# L2 (Redis) Configuration
HARBOR_CACHE_L2_DEFAULT_TTL_MINUTES=10       # Default L2 TTL (default: 10)
HARBOR_CACHE_L2_ENABLED=true                 # Enable L2 cache (default: true)

# Content-Specific TTLs
HARBOR_CACHE_MANIFEST_DIGEST_TTL_HOURS=24    # Manifest by digest TTL (default: 24)
HARBOR_CACHE_MANIFEST_TAG_TTL_MINUTES=5      # Manifest by tag TTL (default: 5)
HARBOR_CACHE_PROJECT_META_TTL_MINUTES=10     # Project metadata TTL (default: 10)
HARBOR_CACHE_ARTIFACT_META_TTL_MINUTES=15    # Artifact metadata TTL (default: 15)
```

### Programmatic Configuration

```go
import "github.com/goharbor/harbor/src/lib/cache/tiered"

config := &tiered.Config{
    L1MaxSize:           200 * 1024 * 1024, // 200MB
    L1MaxEntries:        20000,
    L1DefaultTTL:        2 * time.Minute,
    L2DefaultTTL:        10 * time.Minute,
    L2Address:           "redis://localhost:6379",
    EnableL1:            true,
    EnableL2:            true,
    PromoteOnHit:        true,
    ManifestByDigestTTL: 24 * time.Hour,
    ManifestByTagTTL:    5 * time.Minute,
    ProjectMetaTTL:      10 * time.Minute,
    RepositoryMetaTTL:   10 * time.Minute,
    ArtifactMetaTTL:     15 * time.Minute,
    TagMetaTTL:          5 * time.Minute,
    QueryResultTTL:      1 * time.Minute,
}

cache, err := tiered.NewTieredCache(config)
```

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/goharbor/harbor/src/lib/cache/tiered"
)

// Initialize (typically done once at startup)
cache, err := tiered.InitializeTieredCache()
if err != nil {
    log.Fatalf("Failed to initialize tiered cache: %v", err)
}

// Fetch from cache
var manifest []byte
err = cache.Fetch(ctx, "manifest:sha256:abc123", &manifest)
if err == cache.ErrNotFound {
    // Cache miss - fetch from database
    manifest = fetchFromDatabase()
    // Save to cache
    cache.Save(ctx, "manifest:sha256:abc123", manifest, 24*time.Hour)
}

// Delete from cache (cache invalidation)
cache.Delete(ctx, "manifest:sha256:abc123")
```

### Integration with Existing Managers

The tiered cache is automatically used by:
- `pkg/cached/manifest/redis.Manager` - Manifest caching
- `pkg/cached/artifact/redis.Manager` - Artifact metadata caching

No code changes needed in controllers - they automatically benefit from tiered caching.

## Performance Characteristics

### Expected Performance Improvements

| Operation | Before | After | Improvement |
|-----------|--------|-------|-------------|
| Manifest GET (L1 hit) | 50ms | 0.5ms | **100x faster** |
| Manifest GET (L2 hit) | 50ms | 5ms | **10x faster** |
| Artifact GET (L1 hit) | 30ms | 0.3ms | **100x faster** |
| Artifact GET (L2 hit) | 30ms | 3ms | **10x faster** |
| Database queries | 1000/sec | 200/sec | **80% reduction** |
| Redis queries | 800/sec | 300/sec | **60% reduction** |

### Cache Hit Rates (Expected)

- **L1 Hit Rate**: 60-70% for hot metadata
- **L2 Hit Rate**: 25-30% for warm metadata
- **Overall Hit Rate**: 85-95%
- **Database Hit Rate**: 5-15% for cold/new metadata

### Memory Usage

- **Per Harbor Core Instance**: ~100-200MB for L1 cache
- **Redis**: Depends on workload, typically 1-5GB for L2 cache
- **Total Overhead**: Minimal compared to performance gains

## Monitoring

### Cache Statistics

The tiered cache automatically logs statistics every 30 seconds:

```
Tiered cache stats: total=10000, L1=65.50%, L2=28.30%, overall=93.80%, misses=620
```

### Metrics (Prometheus)

The following metrics are exported:

- `harbor_cache_hits_total{tier="l1"}` - L1 cache hits
- `harbor_cache_hits_total{tier="l2"}` - L2 cache hits
- `harbor_cache_misses_total` - Cache misses
- `harbor_cache_size_bytes{tier="l1"}` - L1 cache size
- `harbor_cache_evictions_total{tier="l1",reason="size_limit"}` - L1 evictions by reason

## Best Practices

### 1. **Size L1 Appropriately**
- For small deployments: 50-100MB
- For medium deployments: 100-200MB
- For large deployments: 200-500MB

### 2. **Monitor Hit Rates**
- L1 hit rate < 50%: Consider increasing L1 size
- Overall hit rate < 80%: Consider increasing TTLs
- High eviction rate: Increase L1 size or max entries

### 3. **Tune TTLs**
- Immutable content (digests): Long TTL (hours/days)
- Mutable content (tags): Short TTL (minutes)
- Metadata: Medium TTL (5-15 minutes)

### 4. **Cache Invalidation**
Always invalidate cache when:
- Deleting artifacts/manifests
- Updating tags
- Modifying project/repository metadata

## Troubleshooting

### High Memory Usage
- Reduce `L1_MAX_SIZE_MB`
- Reduce `L1_MAX_ENTRIES`
- Check for memory leaks in application

### Low Hit Rate
- Increase L1 size
- Increase TTLs
- Check if cache is being invalidated too frequently
- Verify Redis is accessible

### Cache Inconsistency
- Ensure cache invalidation is called on updates/deletes
- Check Redis connectivity
- Verify TTLs are appropriate

## Migration from Simple Cache

The tiered cache is backward compatible with the existing cache interface. To migrate:

1. Update initialization code to use `tiered.InitializeTieredCache()`
2. Existing code using `cache.Cache` interface works unchanged
3. Optionally add content-specific TTLs using `GetTTLForContentType()`

No breaking changes required!

## Future Enhancements

- [ ] Negative caching for 404 responses
- [ ] Cache warming for popular images
- [ ] Adaptive TTLs based on access patterns
- [ ] Compression for large cached values
- [ ] Cache prefetching based on prediction

