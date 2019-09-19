package redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutable"
)

// RedisCache ...
type RedisCache struct {
	pool *redis.Pool
}

// NewRedisCache ...
func NewRedisCache(pool *redis.Pool) RedisCache {
	return RedisCache{
		pool: pool,
	}
}

// Set ...
func (imrc RedisCache) Set(pid int64, c art.Candidate) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("SADD", immutableSetKey(pid), c.Digest); err != nil {
		return err
	}

	return nil
}

// Stat ...
func (imrc RedisCache) Stat(pid int64, digest string) (bool, error) {
	conn := imrc.pool.Get()
	defer conn.Close()

	member, err := redis.Bool(conn.Do("SISMEMBER", immutableSetKey(pid), digest))
	if err != nil {
		return false, err
	}

	if !member {
		return false, immutable.ErrTagUnknown
	}

	return true, nil
}

// Clear ...
func (imrc RedisCache) Clear(pid int64, c art.Candidate) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	key := immutableSetKey(pid)

	// Check membership to repository first
	member, err := redis.Bool(conn.Do("SISMEMBER", key, c.Digest))
	if err != nil {
		return err
	}

	if !member {
		return immutable.ErrTagUnknown
	}

	reply, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}

	if reply == 0 {
		return immutable.ErrTagUnknown
	}

	return nil
}

// Flush ...
func (imrc RedisCache) Flush(pid int64) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("DEL", immutableSetKey(pid))
	if err != nil {
		return err
	}

	if reply == 0 {
		return immutable.ErrTagUnknown
	}

	return nil
}

func immutableSetKey(pid int64) string {
	return fmt.Sprintf("IMMUTABLES::TAGS::%d", pid)
}
