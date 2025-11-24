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
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"

	"github.com/goharbor/harbor/src/lib/log"
	libredis "github.com/goharbor/harbor/src/lib/redis"
)

const (
	// executionRefreshQueueKey is the Redis Set key for storing execution IDs that need status refresh
	executionRefreshQueueKey = "cache:execution:refresh:queue"

	// defaultTotalInstances is the default number of core instances if not configured
	defaultTotalInstances = 1
)

// ExecutionQueue manages the queue of executions that need status refresh using Redis Set
type ExecutionQueue struct {
	client interface {
		SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
		SMembers(ctx context.Context, key string) *redis.StringSliceCmd
		SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
		SCard(ctx context.Context, key string) *redis.IntCmd
	}
}

// NewExecutionQueue creates a new ExecutionQueue instance
func NewExecutionQueue() (*ExecutionQueue, error) {
	client, err := libredis.GetHarborClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis client: %w", err)
	}

	return &ExecutionQueue{
		client: client,
	}, nil
}

// Add adds an execution to the refresh queue
// This is idempotent - adding the same execution multiple times has no effect
func (q *ExecutionQueue) Add(ctx context.Context, executionID int64, vendor string) error {
	member := fmt.Sprintf("%d:%s", executionID, vendor)
	return q.client.SAdd(ctx, executionRefreshQueueKey, member).Err()
}

// GetAll retrieves all executions from the refresh queue
func (q *ExecutionQueue) GetAll(ctx context.Context) ([]ExecutionItem, error) {
	members, err := q.client.SMembers(ctx, executionRefreshQueueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue members: %w", err)
	}

	items := make([]ExecutionItem, 0, len(members))
	for _, member := range members {
		item, err := parseExecutionItem(member)
		if err != nil {
			// Log error but continue processing other items
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// Remove removes an execution from the refresh queue
func (q *ExecutionQueue) Remove(ctx context.Context, executionID int64, vendor string) error {
	member := fmt.Sprintf("%d:%s", executionID, vendor)
	return q.client.SRem(ctx, executionRefreshQueueKey, member).Err()
}

// Size returns the number of items in the queue
func (q *ExecutionQueue) Size(ctx context.Context) (int64, error) {
	return q.client.SCard(ctx, executionRefreshQueueKey).Result()
}

// ExecutionItem represents an execution that needs status refresh
type ExecutionItem struct {
	ExecutionID int64
	Vendor      string
}

// parseExecutionItem parses a queue member string into an ExecutionItem
func parseExecutionItem(member string) (ExecutionItem, error) {
	parts := strings.SplitN(member, ":", 2)
	if len(parts) != 2 {
		return ExecutionItem{}, fmt.Errorf("invalid member format: %s", member)
	}

	executionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ExecutionItem{}, fmt.Errorf("invalid execution ID: %s", parts[0])
	}

	return ExecutionItem{
		ExecutionID: executionID,
		Vendor:      parts[1],
	}, nil
}

// InstanceCoordinator handles instance coordination for distributed execution processing
type InstanceCoordinator struct {
	totalInstances int
	myInstanceID   int
}

// NewInstanceCoordinator creates a new InstanceCoordinator
// It reads configuration from environment variables:
// - CORE_INSTANCE_TOTAL: total number of core instances (default: 1)
// - CORE_INSTANCE_ID: this instance's ID, 0-based (default: 0)
// - HOSTNAME: Kubernetes pod name (auto-detected if CORE_INSTANCE_ID not set)
func NewInstanceCoordinator() *InstanceCoordinator {
	totalInstances := defaultTotalInstances
	if val := os.Getenv("CORE_INSTANCE_TOTAL"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			totalInstances = parsed
		}
	}

	myInstanceID := 0
	if val := os.Getenv("CORE_INSTANCE_ID"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 && parsed < totalInstances {
			myInstanceID = parsed
		}
	} else {
		// Auto-detect from Kubernetes pod name if CORE_INSTANCE_ID not explicitly set
		myInstanceID = detectKubernetesInstanceID()
	}

	return &InstanceCoordinator{
		totalInstances: totalInstances,
		myInstanceID:   myInstanceID,
	}
}

// detectKubernetesInstanceID attempts to extract instance ID from Kubernetes pod name
// This enables zero-configuration deployment in Kubernetes StatefulSets
func detectKubernetesInstanceID() int {
	// Try HOSTNAME environment variable (Kubernetes sets this to pod name)
	if hostname := os.Getenv("HOSTNAME"); hostname != "" {
		if id := extractInstanceIDFromPodName(hostname); id >= 0 {
			log.Infof("Auto-detected Kubernetes instance ID %d from hostname: %s", id, hostname)
			return id
		}
	}

	// Fallback: try to read from /etc/hostname
	if data, err := os.ReadFile("/etc/hostname"); err == nil {
		hostname := strings.TrimSpace(string(data))
		if id := extractInstanceIDFromPodName(hostname); id >= 0 {
			log.Infof("Auto-detected Kubernetes instance ID %d from /etc/hostname: %s", id, hostname)
			return id
		}
	}

	log.Debug("Could not auto-detect Kubernetes instance ID, using default 0")
	return 0
}

// extractInstanceIDFromPodName extracts the numeric suffix from Kubernetes pod names
// Examples:
//
//	harbor-core-0 -> 0
//	harbor-core-3 -> 3
//	harbor-core-statefulset-2 -> 2
//	my-harbor-core-deployment-7b9c8d-5 -> 5
func extractInstanceIDFromPodName(podName string) int {
	if podName == "" {
		return 0
	}

	// Find the last dash and extract number after it
	parts := strings.Split(podName, "-")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if id, err := strconv.Atoi(lastPart); err == nil && id >= 0 {
			return id
		}
	}

	return 0
}

// ShouldProcess determines if this instance should process the given execution
// using consistent hashing based on execution ID
func (ic *InstanceCoordinator) ShouldProcess(executionID int64) bool {
	// If only one instance, always process
	if ic.totalInstances == 1 {
		return true
	}

	// Use modulo for consistent hashing
	assignedInstance := int(executionID % int64(ic.totalInstances))
	return assignedInstance == ic.myInstanceID
}

// GetInstanceInfo returns information about this instance
func (ic *InstanceCoordinator) GetInstanceInfo() (instanceID, totalInstances int) {
	return ic.myInstanceID, ic.totalInstances
}
