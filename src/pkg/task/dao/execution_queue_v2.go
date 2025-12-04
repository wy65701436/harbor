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
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/goharbor/harbor/src/lib/log"
	libredis "github.com/goharbor/harbor/src/lib/redis"
)

const (
	// executionRefreshQueueKeyV2 is the Redis Set key for storing execution IDs that need status refresh
	executionRefreshQueueKeyV2 = "cache:execution:refresh:queue:v2"

	// executionRefreshProcessingKey is the Redis Set for executions currently being processed
	executionRefreshProcessingKey = "cache:execution:refresh:processing"

	// processingTimeout is how long an execution can be in "processing" state before it's considered stale
	processingTimeout = 5 * time.Minute
)

// ExecutionQueueV2 manages the queue using Redis Sets with atomic work claiming
// This approach eliminates the need for instance IDs and consistent hashing
type ExecutionQueueV2 struct {
	client *redis.Client
	nodeID string // Unique identifier for this node (for debugging)
}

// NewExecutionQueueV2 creates a new ExecutionQueueV2 instance
func NewExecutionQueueV2() (*ExecutionQueueV2, error) {
	client, err := libredis.GetHarborClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis client: %w", err)
	}

	// Generate a unique node ID for debugging (hostname + timestamp)
	nodeID := fmt.Sprintf("%s-%d", getHostname(), time.Now().Unix())

	return &ExecutionQueueV2{
		client: client,
		nodeID: nodeID,
	}, nil
}

// Add adds an execution to the refresh queue
func (q *ExecutionQueueV2) Add(ctx context.Context, executionID int64, vendor string) error {
	member := fmt.Sprintf("%d:%s", executionID, vendor)
	return q.client.SAdd(ctx, executionRefreshQueueKeyV2, member).Err()
}

// ClaimBatch atomically claims a batch of executions for processing
// This uses Redis SPOP to atomically remove items from the queue
// Returns the claimed items and any error
func (q *ExecutionQueueV2) ClaimBatch(ctx context.Context, batchSize int) ([]ExecutionItemV2, error) {
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// Use Lua script for atomic batch claiming
	// This ensures no two instances claim the same execution
	script := redis.NewScript(`
		local queue_key = KEYS[1]
		local processing_key = KEYS[2]
		local batch_size = tonumber(ARGV[1])
		local node_id = ARGV[2]
		local timestamp = ARGV[3]
		
		-- Pop up to batch_size items from queue
		local items = redis.call('SPOP', queue_key, batch_size)
		
		if #items == 0 then
			return {}
		end
		
		-- Add to processing set with metadata
		for i, item in ipairs(items) do
			local processing_value = item .. ":" .. node_id .. ":" .. timestamp
			redis.call('SADD', processing_key, processing_value)
			-- Set expiry on processing set to auto-cleanup stale items
			redis.call('EXPIRE', processing_key, 600)
		end
		
		return items
	`)

	timestamp := time.Now().Unix()
	result, err := script.Run(ctx, q.client,
		[]string{executionRefreshQueueKeyV2, executionRefreshProcessingKey},
		batchSize, q.nodeID, timestamp).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to claim batch: %w", err)
	}

	// Parse results
	members, ok := result.([]interface{})
	if !ok || len(members) == 0 {
		return nil, nil
	}

	items := make([]ExecutionItemV2, 0, len(members))
	for _, m := range members {
		memberStr, ok := m.(string)
		if !ok {
			continue
		}

		item, err := parseExecutionItemV2(memberStr)
		if err != nil {
			log.Warningf("failed to parse execution item %s: %v", memberStr, err)
			continue
		}
		items = append(items, item)
	}

	log.Debugf("Node %s claimed %d executions for processing", q.nodeID, len(items))
	return items, nil
}

// MarkComplete removes an execution from the processing set after successful processing
func (q *ExecutionQueueV2) MarkComplete(ctx context.Context, executionID int64, vendor string) error {
	// Remove all entries for this execution from processing set
	// (there might be entries with different node IDs if processing was retried)
	pattern := fmt.Sprintf("%d:%s:*", executionID, vendor)

	// Use SSCAN to find matching members
	iter := q.client.SScan(ctx, executionRefreshProcessingKey, 0, pattern, 100).Iterator()

	var toRemove []string
	for iter.Next(ctx) {
		toRemove = append(toRemove, iter.Val())
	}

	if len(toRemove) > 0 {
		return q.client.SRem(ctx, executionRefreshProcessingKey, toInterface(toRemove)...).Err()
	}

	return nil
}

// MarkFailed returns an execution to the queue if processing failed
func (q *ExecutionQueueV2) MarkFailed(ctx context.Context, executionID int64, vendor string) error {
	member := fmt.Sprintf("%d:%s", executionID, vendor)

	// Remove from processing
	pattern := fmt.Sprintf("%d:%s:*", executionID, vendor)
	iter := q.client.SScan(ctx, executionRefreshProcessingKey, 0, pattern, 100).Iterator()

	var toRemove []string
	for iter.Next(ctx) {
		toRemove = append(toRemove, iter.Val())
	}

	if len(toRemove) > 0 {
		if err := q.client.SRem(ctx, executionRefreshProcessingKey, toInterface(toRemove)...).Err(); err != nil {
			log.Warningf("failed to remove from processing set: %v", err)
		}
	}

	// Add back to queue for retry
	return q.client.SAdd(ctx, executionRefreshQueueKeyV2, member).Err()
}

// RecoverStaleProcessing moves stale items from processing back to queue
// This should be called periodically to handle crashed instances
func (q *ExecutionQueueV2) RecoverStaleProcessing(ctx context.Context) (int, error) {
	// Get all processing items
	members, err := q.client.SMembers(ctx, executionRefreshProcessingKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get processing items: %w", err)
	}

	now := time.Now().Unix()
	recovered := 0

	for _, member := range members {
		// Parse: "executionID:vendor:nodeID:timestamp"
		parts := strings.Split(member, ":")
		if len(parts) < 4 {
			continue
		}

		timestamp, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
		if err != nil {
			continue
		}

		// If processing for more than timeout, recover it
		if now-timestamp > int64(processingTimeout.Seconds()) {
			// Extract execution ID and vendor
			executionID := parts[0]
			vendor := parts[1]

			// Move back to queue
			queueMember := fmt.Sprintf("%s:%s", executionID, vendor)
			if err := q.client.SAdd(ctx, executionRefreshQueueKeyV2, queueMember).Err(); err != nil {
				log.Warningf("failed to recover stale execution %s: %v", queueMember, err)
				continue
			}

			// Remove from processing
			if err := q.client.SRem(ctx, executionRefreshProcessingKey, member).Err(); err != nil {
				log.Warningf("failed to remove stale processing entry %s: %v", member, err)
			}

			recovered++
			log.Infof("Recovered stale execution %s (was processing by %s for %d seconds)",
				queueMember, parts[2], now-timestamp)
		}
	}

	return recovered, nil
}

// GetQueueSize returns the number of items waiting in the queue
func (q *ExecutionQueueV2) GetQueueSize(ctx context.Context) (int64, error) {
	return q.client.SCard(ctx, executionRefreshQueueKeyV2).Result()
}

// GetProcessingSize returns the number of items currently being processed
func (q *ExecutionQueueV2) GetProcessingSize(ctx context.Context) (int64, error) {
	return q.client.SCard(ctx, executionRefreshProcessingKey).Result()
}

// ExecutionItemV2 represents an execution that needs status refresh
type ExecutionItemV2 struct {
	ExecutionID int64
	Vendor      string
}

// parseExecutionItemV2 parses a queue member string into an ExecutionItemV2
func parseExecutionItemV2(member string) (ExecutionItemV2, error) {
	parts := strings.SplitN(member, ":", 2)
	if len(parts) != 2 {
		return ExecutionItemV2{}, fmt.Errorf("invalid member format: %s", member)
	}

	executionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ExecutionItemV2{}, fmt.Errorf("invalid execution ID: %s", parts[0])
	}

	return ExecutionItemV2{
		ExecutionID: executionID,
		Vendor:      parts[1],
	}, nil
}

// Helper functions
func getHostname() string {
	if hostname := os.Getenv("HOSTNAME"); hostname != "" {
		return hostname
	}
	if data, err := os.ReadFile("/etc/hostname"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "unknown"
}

func toInterface(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}
