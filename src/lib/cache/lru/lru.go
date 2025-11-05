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

package lru

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
)

// EvictionReason indicates why an entry was evicted
type EvictionReason string

const (
	EvictionReasonSize    EvictionReason = "size_limit"
	EvictionReasonCount   EvictionReason = "count_limit"
	EvictionReasonExpired EvictionReason = "expired"
	EvictionReasonManual  EvictionReason = "manual"
)

// Config holds LRU cache configuration
type Config struct {
	MaxSize    int64                                   // Max total size in bytes
	MaxEntries int                                     // Max number of entries
	DefaultTTL time.Duration                           // Default TTL
	OnEvict    func(key string, reason EvictionReason) // Eviction callback
}

// entry represents a cache entry
type entry struct {
	key      string
	value    []byte
	size     int64
	expireAt time.Time
	listElem *list.Element
}

// Cache is an LRU cache with size limits optimized for Harbor metadata
type Cache struct {
	mu          sync.RWMutex
	config      Config
	items       map[string]*entry
	lruList     *list.List
	currentSize int64
	codec       cache.Codec
}

// New creates a new LRU cache
func New(config Config) (*Cache, error) {
	if config.MaxSize <= 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB default
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = 10000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 5 * time.Minute
	}

	c := &Cache{
		config:  config,
		items:   make(map[string]*entry),
		lruList: list.New(),
		codec:   cache.DefaultCodec(),
	}

	// Start cleanup goroutine for expired entries
	go c.cleanupExpired()

	return c, nil
}

// Fetch retrieves value from cache
func (c *Cache) Fetch(ctx context.Context, key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ent, ok := c.items[key]
	if !ok {
		return cache.ErrNotFound
	}

	// Check expiration
	if time.Now().After(ent.expireAt) {
		c.removeEntry(ent, EvictionReasonExpired)
		return cache.ErrNotFound
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(ent.listElem)

	// Decode value
	if err := c.codec.Decode(ent.value, value); err != nil {
		return fmt.Errorf("failed to decode cached value: %w", err)
	}

	return nil
}

// Save stores value in cache
func (c *Cache) Save(ctx context.Context, key string, value any, expiration ...time.Duration) error {
	data, err := c.codec.Encode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	size := int64(len(data))

	// Check if single entry exceeds max size
	if size > c.config.MaxSize {
		return fmt.Errorf("entry size %d exceeds max cache size %d", size, c.config.MaxSize)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate expiration
	ttl := c.config.DefaultTTL
	if len(expiration) > 0 {
		ttl = expiration[0]
	}
	expireAt := time.Now().Add(ttl)

	// Check if key already exists
	if existingEnt, ok := c.items[key]; ok {
		// Update existing entry
		c.currentSize -= existingEnt.size
		existingEnt.value = data
		existingEnt.size = size
		existingEnt.expireAt = expireAt
		c.currentSize += size
		c.lruList.MoveToFront(existingEnt.listElem)
	} else {
		// Evict entries if necessary to make room
		c.evictIfNeeded(size)

		// Add new entry
		ent := &entry{
			key:      key,
			value:    data,
			size:     size,
			expireAt: expireAt,
		}
		ent.listElem = c.lruList.PushFront(ent)
		c.items[key] = ent
		c.currentSize += size
	}

	return nil
}

// evictIfNeeded evicts LRU entries to make room for new entry
func (c *Cache) evictIfNeeded(newSize int64) {
	// Evict by size limit
	for c.currentSize+newSize > c.config.MaxSize && c.lruList.Len() > 0 {
		elem := c.lruList.Back()
		if elem != nil {
			ent := elem.Value.(*entry)
			c.removeEntry(ent, EvictionReasonSize)
		}
	}

	// Evict by count limit
	for len(c.items) >= c.config.MaxEntries && c.lruList.Len() > 0 {
		elem := c.lruList.Back()
		if elem != nil {
			ent := elem.Value.(*entry)
			c.removeEntry(ent, EvictionReasonCount)
		}
	}
}

// removeEntry removes an entry from cache
func (c *Cache) removeEntry(ent *entry, reason EvictionReason) {
	c.lruList.Remove(ent.listElem)
	delete(c.items, ent.key)
	c.currentSize -= ent.size

	if c.config.OnEvict != nil {
		c.config.OnEvict(ent.key, reason)
	}
}

// Delete removes key from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.removeEntry(ent, EvictionReasonManual)
	}
	return nil
}

// Contains checks if key exists and is not expired
func (c *Cache) Contains(ctx context.Context, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ent, ok := c.items[key]
	if !ok {
		return false
	}

	return time.Now().Before(ent.expireAt)
}

// Ping checks cache health
func (c *Cache) Ping(ctx context.Context) error {
	return nil
}

// Scan returns iterator over cache keys
func (c *Cache) Scan(ctx context.Context, match string) (cache.Iterator, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for key := range c.items {
		keys = append(keys, key)
	}

	return &scanIterator{keys: keys}, nil
}

// cleanupExpired periodically removes expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		var toRemove []*entry

		for _, ent := range c.items {
			if now.After(ent.expireAt) {
				toRemove = append(toRemove, ent)
			}
		}

		for _, ent := range toRemove {
			c.removeEntry(ent, EvictionReasonExpired)
		}
		c.mu.Unlock()
	}
}

// Stats returns cache statistics
func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Entries:     len(c.items),
		CurrentSize: c.currentSize,
		MaxSize:     c.config.MaxSize,
		MaxEntries:  c.config.MaxEntries,
	}
}

// Stats holds cache statistics
type Stats struct {
	Entries     int
	CurrentSize int64
	MaxSize     int64
	MaxEntries  int
}

type scanIterator struct {
	keys []string
	pos  int
}

func (i *scanIterator) Next(ctx context.Context) bool {
	i.pos++
	return i.pos <= len(i.keys)
}

func (i *scanIterator) Val() string {
	if i.pos > 0 && i.pos <= len(i.keys) {
		return i.keys[i.pos-1]
	}
	return ""
}
