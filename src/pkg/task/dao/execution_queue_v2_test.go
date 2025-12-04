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

package dao

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	htesting "github.com/goharbor/harbor/src/testing"
)

type executionQueueV2TestSuite struct {
	htesting.Suite
	queue *ExecutionQueueV2
	ctx   context.Context
}

func (suite *executionQueueV2TestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"execution", "task"}
	suite.ctx = context.TODO()
}

func (suite *executionQueueV2TestSuite) SetupTest() {
	// Initialize cache for testing
	if err := cache.Initialize("memory", ""); err != nil {
		suite.T().Fatalf("failed to initialize cache: %v", err)
	}

	// Note: ExecutionQueueV2 requires Redis client
	queue, err := NewExecutionQueueV2()
	if err != nil {
		suite.T().Skip("Redis not available, skipping V2 queue tests")
	}
	suite.queue = queue

	// Clean up any existing data
	suite.cleanupQueue()
}

func (suite *executionQueueV2TestSuite) TearDownTest() {
	suite.cleanupQueue()
}

func (suite *executionQueueV2TestSuite) cleanupQueue() {
	if suite.queue != nil {
		// Clean queue
		suite.queue.client.Del(suite.ctx, executionRefreshQueueKeyV2)
		// Clean processing set
		suite.queue.client.Del(suite.ctx, executionRefreshProcessingKey)
	}
}

func (suite *executionQueueV2TestSuite) TestAddAndClaimBatch() {
	// Add executions to queue
	err := suite.queue.Add(suite.ctx, 1, "GC")
	suite.NoError(err)

	err = suite.queue.Add(suite.ctx, 2, "REPLICATION")
	suite.NoError(err)

	err = suite.queue.Add(suite.ctx, 3, "SCAN")
	suite.NoError(err)

	// Check queue size
	size, err := suite.queue.GetQueueSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(3), size)

	// Claim a batch
	items, err := suite.queue.ClaimBatch(suite.ctx, 2)
	suite.NoError(err)
	suite.Len(items, 2)

	// Queue should have 1 item left
	size, err = suite.queue.GetQueueSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(1), size)

	// Processing should have 2 items
	procSize, err := suite.queue.GetProcessingSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(2), procSize)
}

func (suite *executionQueueV2TestSuite) TestAtomicClaiming() {
	// Add 100 items
	for i := 1; i <= 100; i++ {
		err := suite.queue.Add(suite.ctx, int64(i), "TEST")
		suite.NoError(err)
	}

	// Simulate 5 instances claiming concurrently
	var wg sync.WaitGroup
	claimed := make([][]ExecutionItemV2, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			items, err := suite.queue.ClaimBatch(suite.ctx, 20)
			if err == nil {
				claimed[idx] = items
			}
		}(i)
	}

	wg.Wait()

	// Verify no duplicates
	seen := make(map[int64]bool)
	totalClaimed := 0

	for _, items := range claimed {
		for _, item := range items {
			suite.False(seen[item.ExecutionID], "Duplicate execution ID: %d", item.ExecutionID)
			seen[item.ExecutionID] = true
			totalClaimed++
		}
	}

	suite.Equal(100, totalClaimed, "All items should be claimed exactly once")
}

func (suite *executionQueueV2TestSuite) TestMarkComplete() {
	// Add and claim
	suite.queue.Add(suite.ctx, 1, "GC")
	items, _ := suite.queue.ClaimBatch(suite.ctx, 1)
	suite.Len(items, 1)

	// Mark complete
	err := suite.queue.MarkComplete(suite.ctx, items[0].ExecutionID, items[0].Vendor)
	suite.NoError(err)

	// Processing should be empty
	size, err := suite.queue.GetProcessingSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(0), size)
}

func (suite *executionQueueV2TestSuite) TestMarkFailed() {
	// Add and claim
	suite.queue.Add(suite.ctx, 1, "GC")
	items, _ := suite.queue.ClaimBatch(suite.ctx, 1)
	suite.Len(items, 1)

	// Queue should be empty
	size, _ := suite.queue.GetQueueSize(suite.ctx)
	suite.Equal(int64(0), size)

	// Mark failed (returns to queue)
	err := suite.queue.MarkFailed(suite.ctx, items[0].ExecutionID, items[0].Vendor)
	suite.NoError(err)

	// Should be back in queue
	size, err = suite.queue.GetQueueSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(1), size)

	// Processing should be empty
	procSize, err := suite.queue.GetProcessingSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(0), procSize)
}

func (suite *executionQueueV2TestSuite) TestStaleRecovery() {
	// Add and claim
	suite.queue.Add(suite.ctx, 1, "GC")
	items, _ := suite.queue.ClaimBatch(suite.ctx, 1)
	suite.Len(items, 1)

	// Manually set timestamp to old value (simulate stale processing)
	oldTimestamp := time.Now().Add(-10 * time.Minute).Unix()
	staleEntry := fmt.Sprintf("%d:%s:%s:%d",
		items[0].ExecutionID, items[0].Vendor, suite.queue.nodeID, oldTimestamp)

	// Remove current entry and add stale one
	suite.queue.client.Del(suite.ctx, executionRefreshProcessingKey)
	suite.queue.client.SAdd(suite.ctx, executionRefreshProcessingKey, staleEntry)

	// Run recovery
	recovered, err := suite.queue.RecoverStaleProcessing(suite.ctx)
	suite.NoError(err)
	suite.Equal(1, recovered)

	// Should be back in queue
	size, err := suite.queue.GetQueueSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(1), size)

	// Processing should be empty
	procSize, err := suite.queue.GetProcessingSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(0), procSize)
}

func (suite *executionQueueV2TestSuite) TestIdempotentAdd() {
	// Add same item multiple times
	for i := 0; i < 5; i++ {
		err := suite.queue.Add(suite.ctx, 1, "GC")
		suite.NoError(err)
	}

	// Should only have one item
	size, err := suite.queue.GetQueueSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(1), size)
}

func (suite *executionQueueV2TestSuite) TestEmptyBatchClaim() {
	// Claim from empty queue
	items, err := suite.queue.ClaimBatch(suite.ctx, 10)
	suite.NoError(err)
	suite.Len(items, 0)
}

func (suite *executionQueueV2TestSuite) TestLargeBatch() {
	// Add 1000 items
	for i := 1; i <= 1000; i++ {
		suite.queue.Add(suite.ctx, int64(i), "TEST")
	}

	// Claim large batch
	items, err := suite.queue.ClaimBatch(suite.ctx, 500)
	suite.NoError(err)
	suite.Len(items, 500)

	// Verify queue size
	size, err := suite.queue.GetQueueSize(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(500), size)
}

func TestExecutionQueueV2Suite(t *testing.T) {
	suite.Run(t, &executionQueueV2TestSuite{})
}

func TestParseExecutionItemV2(t *testing.T) {
	testCases := []struct {
		input          string
		expectError    bool
		expectedID     int64
		expectedVendor string
	}{
		{"123:GC", false, 123, "GC"},
		{"456:REPLICATION", false, 456, "REPLICATION"},
		{"789:SCAN_ALL", false, 789, "SCAN_ALL"},
		{"invalid", true, 0, ""},
		{"abc:GC", true, 0, ""},
		{"", true, 0, ""},
	}

	for _, tc := range testCases {
		item, err := parseExecutionItemV2(tc.input)

		if tc.expectError {
			if err == nil {
				t.Errorf("Expected error for input '%s', got none", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
			}
			if item.ExecutionID != tc.expectedID {
				t.Errorf("Expected ID %d, got %d", tc.expectedID, item.ExecutionID)
			}
			if item.Vendor != tc.expectedVendor {
				t.Errorf("Expected vendor '%s', got '%s'", tc.expectedVendor, item.Vendor)
			}
		}
	}
}
