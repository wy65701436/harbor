package gc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaxFails(t *testing.T) {
	rep := &GarbageCollector{}
	assert.Equal(t, uint(3), rep.MaxFails())
}

func TestShouldRetry(t *testing.T) {
	rep := &GarbageCollector{}
	assert.True(t, rep.ShouldRetry())
}

func TestValidate(t *testing.T) {
	rep := &GarbageCollector{}
	assert.Nil(t, rep.Validate(nil))
}
