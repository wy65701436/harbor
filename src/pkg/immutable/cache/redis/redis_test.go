package redis

import (
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/art"
	immu_cache "github.com/goharbor/harbor/src/pkg/immutable/cache"
	"github.com/stretchr/testify/require"
	"testing"
)

// RedisRunnerTestSuite tests functions of redis runner
type RedisRunnerTestSuite struct {
	suite.Suite
	pool *redis.Pool
}

// TestRedisRunnerTestSuite is entry of go test
func TestRedisRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(RedisRunnerTestSuite))
}

// SetupSuite prepares test suite
func (suite *RedisRunnerTestSuite) SetupSuite() {
	suite.pool = tests.GiveMeRedisPool()
}

// TestRedisCache tests the redis job wrapper
func (suite *RedisRunnerTestSuite) TestRedisCache() {
	cache := NewRedisCache(suite.pool)
	immc := immu_cache.IMCandidate{
		Candidate: art.Candidate{
			Repository: "test_set_repo",
			Tag:        "test_set_tag",
		},
		Immutable: true,
	}
	err := cache.Set(1, immc)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)

	isImmu, err := cache.Stat(1, immc.Repository, immc.Tag)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)
	require.True(suite.T(), isImmu)

	err = cache.Clear(1, immc)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)

	isImmuAfterDel, err := cache.Stat(1, immc.Repository, immc.Tag)
	require.Error(suite.T(), err, "error expected, %s", err)
	require.False(suite.T(), isImmuAfterDel)

	var immc_mul []immu_cache.IMCandidate
	immc_mul1 := immu_cache.IMCandidate{
		Candidate: art.Candidate{
			Repository: "immc_mul1_repo",
			Tag:        "immc_mul1_tag",
		},
		Immutable: true,
	}
	immc_mul2 := immu_cache.IMCandidate{
		Candidate: art.Candidate{
			Repository: "immc_mul2_repo",
			Tag:        "immc_mul2_tag",
		},
		Immutable: false,
	}
	immc_mul = append(immc_mul, immc_mul1)
	immc_mul = append(immc_mul, immc_mul2)

	err = cache.SetMultiple(1, immc_mul)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)
	_, err = cache.Stat(1, immc_mul1.Repository, immc_mul1.Tag)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)

	err = cache.Flush(1)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)

	_, err = cache.Stat(1, immc_mul1.Repository, immc_mul1.Tag)
	require.Error(suite.T(), err, "error expected, %s", err)
	_, err = cache.Stat(1, immc_mul2.Repository, immc_mul2.Tag)
	require.Error(suite.T(), err, "error expected, %s", err)
}
