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
	"os"
	"testing"
)

func TestExtractInstanceIDFromPodName(t *testing.T) {
	testCases := []struct {
		podName    string
		expectedID int
	}{
		// StatefulSet pod names
		{"harbor-core-0", 0},
		{"harbor-core-1", 1},
		{"harbor-core-5", 5},
		{"harbor-core-99", 99},

		// Different naming patterns
		{"my-harbor-core-3", 3},
		{"harbor-core-statefulset-7", 7},

		// Deployment pod names (with random suffix)
		{"harbor-core-deployment-7b9c8d-5", 5},
		{"harbor-core-abc123-10", 10},

		// Edge cases
		{"harbor-core", 0},     // No number
		{"", 0},                // Empty string
		{"no-numbers-here", 0}, // No numeric suffix
		{"harbor-0-core-3", 3}, // Multiple numbers (uses last)
	}

	for _, tc := range testCases {
		result := extractInstanceIDFromPodName(tc.podName)
		if result != tc.expectedID {
			t.Errorf("Pod name '%s': expected ID %d, got %d", tc.podName, tc.expectedID, result)
		}
	}
}

func TestKubernetesAutoDetection(t *testing.T) {
	// Save original environment
	originalHostname := os.Getenv("HOSTNAME")
	originalTotal := os.Getenv("CORE_INSTANCE_TOTAL")
	originalID := os.Getenv("CORE_INSTANCE_ID")

	defer func() {
		// Restore original environment
		if originalHostname != "" {
			os.Setenv("HOSTNAME", originalHostname)
		} else {
			os.Unsetenv("HOSTNAME")
		}
		if originalTotal != "" {
			os.Setenv("CORE_INSTANCE_TOTAL", originalTotal)
		} else {
			os.Unsetenv("CORE_INSTANCE_TOTAL")
		}
		if originalID != "" {
			os.Setenv("CORE_INSTANCE_ID", originalID)
		} else {
			os.Unsetenv("CORE_INSTANCE_ID")
		}
	}()

	// Test auto-detection from HOSTNAME
	os.Unsetenv("CORE_INSTANCE_ID") // Ensure manual ID is not set
	os.Setenv("CORE_INSTANCE_TOTAL", "5")
	os.Setenv("HOSTNAME", "harbor-core-3")

	coordinator := NewInstanceCoordinator()
	instanceID, totalInstances := coordinator.GetInstanceInfo()

	if instanceID != 3 {
		t.Errorf("Expected auto-detected instance ID 3, got %d", instanceID)
	}
	if totalInstances != 5 {
		t.Errorf("Expected total instances 5, got %d", totalInstances)
	}

	// Verify it processes correct executions
	testCases := []struct {
		executionID   int64
		shouldProcess bool
	}{
		{3, true},  // 3 % 5 = 3
		{8, true},  // 8 % 5 = 3
		{13, true}, // 13 % 5 = 3
		{0, false}, // 0 % 5 = 0
		{1, false}, // 1 % 5 = 1
		{5, false}, // 5 % 5 = 0
	}

	for _, tc := range testCases {
		result := coordinator.ShouldProcess(tc.executionID)
		if result != tc.shouldProcess {
			t.Errorf("Execution %d: expected %v, got %v", tc.executionID, tc.shouldProcess, result)
		}
	}
}

func TestManualConfigurationOverridesAutoDetection(t *testing.T) {
	// Save original environment
	originalHostname := os.Getenv("HOSTNAME")
	originalTotal := os.Getenv("CORE_INSTANCE_TOTAL")
	originalID := os.Getenv("CORE_INSTANCE_ID")

	defer func() {
		if originalHostname != "" {
			os.Setenv("HOSTNAME", originalHostname)
		} else {
			os.Unsetenv("HOSTNAME")
		}
		if originalTotal != "" {
			os.Setenv("CORE_INSTANCE_TOTAL", originalTotal)
		} else {
			os.Unsetenv("CORE_INSTANCE_TOTAL")
		}
		if originalID != "" {
			os.Setenv("CORE_INSTANCE_ID", originalID)
		} else {
			os.Unsetenv("CORE_INSTANCE_ID")
		}
	}()

	// Set both HOSTNAME (auto-detect) and CORE_INSTANCE_ID (manual)
	os.Setenv("HOSTNAME", "harbor-core-3")
	os.Setenv("CORE_INSTANCE_TOTAL", "5")
	os.Setenv("CORE_INSTANCE_ID", "2") // Manual override

	coordinator := NewInstanceCoordinator()
	instanceID, _ := coordinator.GetInstanceInfo()

	// Manual configuration should take precedence
	if instanceID != 2 {
		t.Errorf("Expected manual instance ID 2 to override auto-detected 3, got %d", instanceID)
	}
}

func TestKubernetesAutoDetectionWithDifferentPodNames(t *testing.T) {
	testCases := []struct {
		hostname    string
		expectedID  int
		description string
	}{
		{"harbor-core-0", 0, "StatefulSet pod 0"},
		{"harbor-core-4", 4, "StatefulSet pod 4"},
		{"my-harbor-core-2", 2, "Custom prefix"},
		{"harbor-core-deployment-abc123-7", 7, "Deployment with random suffix"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			os.Unsetenv("CORE_INSTANCE_ID")
			os.Setenv("CORE_INSTANCE_TOTAL", "5")
			os.Setenv("HOSTNAME", tc.hostname)

			coordinator := NewInstanceCoordinator()
			instanceID, _ := coordinator.GetInstanceInfo()

			if instanceID != tc.expectedID {
				t.Errorf("%s: expected ID %d, got %d", tc.description, tc.expectedID, instanceID)
			}
		})
	}
}
