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
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	htesting "github.com/goharbor/harbor/src/testing"
)

type executionQueueTestSuite struct {
	htesting.Suite
	queue *ExecutionQueue
	ctx   context.Context
}

func (suite *executionQueueTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"execution", "task"}
	suite.ctx = context.TODO()
}

func (suite *executionQueueTestSuite) SetupTest() {
	// Initialize cache for testing
	if err := cache.Initialize("memory", ""); err != nil {
		suite.T().Fatalf("failed to initialize cache: %v", err)
	}

	// Note: ExecutionQueue requires Redis client, so we'll need to mock or skip in unit tests
	// For integration tests, ensure Redis is available
	queue, err := NewExecutionQueue()
	if err != nil {
		suite.T().Skip("Redis not available, skipping queue tests")
	}
	suite.queue = queue
}

func (suite *executionQueueTestSuite) TearDownTest() {
	// Clean up queue
	if suite.queue != nil {
		items, _ := suite.queue.GetAll(suite.ctx)
		for _, item := range items {
			_ = suite.queue.Remove(suite.ctx, item.ExecutionID, item.Vendor)
		}
	}
}

func (suite *executionQueueTestSuite) TestAddAndGetAll() {
	// Add some executions to the queue
	err := suite.queue.Add(suite.ctx, 1, "GC")
	suite.NoError(err)

	err = suite.queue.Add(suite.ctx, 2, "REPLICATION")
	suite.NoError(err)

	err = suite.queue.Add(suite.ctx, 3, "SCAN")
	suite.NoError(err)

	// Get all items
	items, err := suite.queue.GetAll(suite.ctx)
	suite.NoError(err)
	suite.Len(items, 3)

	// Verify items
	executionIDs := make(map[int64]string)
	for _, item := range items {
		executionIDs[item.ExecutionID] = item.Vendor
	}

	suite.Equal("GC", executionIDs[1])
	suite.Equal("REPLICATION", executionIDs[2])
	suite.Equal("SCAN", executionIDs[3])
}

func (suite *executionQueueTestSuite) TestAddIdempotent() {
	// Add the same execution multiple times
	err := suite.queue.Add(suite.ctx, 1, "GC")
	suite.NoError(err)

	err = suite.queue.Add(suite.ctx, 1, "GC")
	suite.NoError(err)

	err = suite.queue.Add(suite.ctx, 1, "GC")
	suite.NoError(err)

	// Should only have one item
	items, err := suite.queue.GetAll(suite.ctx)
	suite.NoError(err)
	suite.Len(items, 1)
	suite.Equal(int64(1), items[0].ExecutionID)
	suite.Equal("GC", items[0].Vendor)
}

func (suite *executionQueueTestSuite) TestRemove() {
	// Add items
	err := suite.queue.Add(suite.ctx, 1, "GC")
	suite.NoError(err)
	err = suite.queue.Add(suite.ctx, 2, "REPLICATION")
	suite.NoError(err)

	// Remove one item
	err = suite.queue.Remove(suite.ctx, 1, "GC")
	suite.NoError(err)

	// Should only have one item left
	items, err := suite.queue.GetAll(suite.ctx)
	suite.NoError(err)
	suite.Len(items, 1)
	suite.Equal(int64(2), items[0].ExecutionID)
}

func (suite *executionQueueTestSuite) TestSize() {
	// Initially empty
	size, err := suite.queue.Size(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(0), size)

	// Add items
	_ = suite.queue.Add(suite.ctx, 1, "GC")
	_ = suite.queue.Add(suite.ctx, 2, "REPLICATION")
	_ = suite.queue.Add(suite.ctx, 3, "SCAN")

	size, err = suite.queue.Size(suite.ctx)
	suite.NoError(err)
	suite.Equal(int64(3), size)
}

func TestExecutionQueueSuite(t *testing.T) {
	suite.Run(t, &executionQueueTestSuite{})
}

// Test InstanceCoordinator
func TestInstanceCoordinator(t *testing.T) {
	// Test default values (single instance)
	coordinator := NewInstanceCoordinator()
	instanceID, totalInstances := coordinator.GetInstanceInfo()
	if instanceID != 0 || totalInstances != 1 {
		t.Errorf("Expected default (0, 1), got (%d, %d)", instanceID, totalInstances)
	}

	// Should process all executions in single instance mode
	if !coordinator.ShouldProcess(1) {
		t.Error("Single instance should process all executions")
	}
	if !coordinator.ShouldProcess(100) {
		t.Error("Single instance should process all executions")
	}
}

func TestInstanceCoordinatorMultiInstance(t *testing.T) {
	// Set environment variables for multi-instance
	os.Setenv("CORE_INSTANCE_TOTAL", "5")
	os.Setenv("CORE_INSTANCE_ID", "2")
	defer os.Unsetenv("CORE_INSTANCE_TOTAL")
	defer os.Unsetenv("CORE_INSTANCE_ID")

	coordinator := NewInstanceCoordinator()
	instanceID, totalInstances := coordinator.GetInstanceInfo()

	if instanceID != 2 || totalInstances != 5 {
		t.Errorf("Expected (2, 5), got (%d, %d)", instanceID, totalInstances)
	}

	// Test consistent hashing
	// Instance 2 should process executions where id % 5 == 2
	testCases := []struct {
		executionID   int64
		shouldProcess bool
	}{
		{2, true},   // 2 % 5 = 2
		{7, true},   // 7 % 5 = 2
		{12, true},  // 12 % 5 = 2
		{1, false},  // 1 % 5 = 1
		{3, false},  // 3 % 5 = 3
		{10, false}, // 10 % 5 = 0
	}

	for _, tc := range testCases {
		result := coordinator.ShouldProcess(tc.executionID)
		if result != tc.shouldProcess {
			t.Errorf("Execution %d: expected %v, got %v", tc.executionID, tc.shouldProcess, result)
		}
	}
}

func TestInstanceCoordinatorDistribution(t *testing.T) {
	// Test that work is evenly distributed across instances
	totalInstances := 5
	executionCount := 10000

	// Count how many executions each instance would process
	counts := make([]int, totalInstances)

	for instanceID := 0; instanceID < totalInstances; instanceID++ {
		os.Setenv("CORE_INSTANCE_TOTAL", "5")
		os.Setenv("CORE_INSTANCE_ID", string(rune('0'+instanceID)))

		coordinator := NewInstanceCoordinator()

		for execID := int64(0); execID < int64(executionCount); execID++ {
			if coordinator.ShouldProcess(execID) {
				counts[instanceID]++
			}
		}

		os.Unsetenv("CORE_INSTANCE_TOTAL")
		os.Unsetenv("CORE_INSTANCE_ID")
	}

	// Each instance should process approximately executionCount/totalInstances
	expectedPerInstance := executionCount / totalInstances

	for i, count := range counts {
		if count != expectedPerInstance {
			t.Errorf("Instance %d processed %d executions, expected %d", i, count, expectedPerInstance)
		}
	}

	// Verify total
	total := 0
	for _, count := range counts {
		total += count
	}
	if total != executionCount {
		t.Errorf("Total processed: %d, expected %d", total, executionCount)
	}
}

func TestParseExecutionItem(t *testing.T) {
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
		item, err := parseExecutionItem(tc.input)

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
