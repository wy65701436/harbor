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

package inmemory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/errors"
)

// InMemoryDriverGetTestSuite tests the Get method in InMemory driver
type InMemoryDriverGetTestSuite struct {
	suite.Suite
	ctx    context.Context
	driver *Driver
}

func (suite *InMemoryDriverGetTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.driver = &Driver{
		cfgMap: map[string]any{
			common.SkipAuditLogDatabase:    true,
			common.AuditLogForwardEndpoint: "syslog://localhost:514",
			"test_key":                     "test_value",
		},
	}
}

// TestGetMethodReturnsUnsupported tests that Get method returns ErrUnsupported
func (suite *InMemoryDriverGetTestSuite) TestGetMethodReturnsUnsupported() {
	key := common.SkipAuditLogDatabase

	result, err := suite.driver.Get(suite.ctx, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)
}

// TestGetMethodWithDifferentKeys tests Get method with various keys
func (suite *InMemoryDriverGetTestSuite) TestGetMethodWithDifferentKeys() {
	testCases := []struct {
		name string
		key  string
	}{
		{
			name: "skip_audit_log_database",
			key:  common.SkipAuditLogDatabase,
		},
		{
			name: "audit_log_forward_endpoint",
			key:  common.AuditLogForwardEndpoint,
		},
		{
			name: "pull_audit_log_disable",
			key:  common.PullAuditLogDisable,
		},
		{
			name: "existing_test_key",
			key:  "test_key",
		},
		{
			name: "empty_key",
			key:  "",
		},
		{
			name: "non_existent_key",
			key:  "non_existent_config",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := suite.driver.Get(suite.ctx, tc.key)

			suite.Require().Error(err)
			suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
			suite.Nil(result)
		})
	}
}

// TestGetMethodWithNilContext tests Get method with nil context
func (suite *InMemoryDriverGetTestSuite) TestGetMethodWithNilContext() {
	key := common.SkipAuditLogDatabase

	result, err := suite.driver.Get(nil, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)
}

// TestGetMethodWithEmptyDriver tests Get method with empty driver
func (suite *InMemoryDriverGetTestSuite) TestGetMethodWithEmptyDriver() {
	emptyDriver := &Driver{cfgMap: map[string]any{}}
	key := common.SkipAuditLogDatabase

	result, err := emptyDriver.Get(suite.ctx, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)
}

// TestGetMethodWithNilCfgMap tests Get method with nil cfgMap
func (suite *InMemoryDriverGetTestSuite) TestGetMethodWithNilCfgMap() {
	nilMapDriver := &Driver{cfgMap: nil}
	key := common.SkipAuditLogDatabase

	result, err := nilMapDriver.Get(suite.ctx, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)
}

// TestGetMethodConsistency tests that Get method consistently returns ErrUnsupported
func (suite *InMemoryDriverGetTestSuite) TestGetMethodConsistency() {
	key := common.SkipAuditLogDatabase

	// Call multiple times to ensure consistency
	for i := 0; i < 5; i++ {
		result, err := suite.driver.Get(suite.ctx, key)

		suite.Require().Error(err)
		suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
		suite.Nil(result)
	}
}

// TestGetMethodDoesNotModifyState tests that Get method doesn't modify driver state
func (suite *InMemoryDriverGetTestSuite) TestGetMethodDoesNotModifyState() {
	originalMap := make(map[string]any)
	for k, v := range suite.driver.cfgMap {
		originalMap[k] = v
	}

	key := common.SkipAuditLogDatabase
	result, err := suite.driver.Get(suite.ctx, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)

	// Verify that the cfgMap hasn't been modified
	suite.Equal(originalMap, suite.driver.cfgMap)
}

// Run the test suite
func TestInMemoryDriverGetTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryDriverGetTestSuite))
}

// TestInMemoryDriverGetStandalone provides additional standalone tests
func TestInMemoryDriverGetStandalone(t *testing.T) {
	driver := &Driver{
		cfgMap: map[string]any{
			common.SkipAuditLogDatabase: false,
		},
	}
	ctx := context.Background()

	// Test that Get method is properly implemented and returns expected error
	result, err := driver.Get(ctx, common.SkipAuditLogDatabase)

	assert.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.MethodNotAllowedCode))
	assert.Nil(t, result)
}

// TestInMemoryDriverGetDocumentation tests that the Get method behavior matches the TODO comment
func TestInMemoryDriverGetDocumentation(t *testing.T) {
	driver := &Driver{
		cfgMap: map[string]any{
			common.SkipAuditLogDatabase:    true,
			common.AuditLogForwardEndpoint: "syslog://test:514",
		},
	}
	ctx := context.Background()

	// The Get method has a TODO comment, indicating it's not fully implemented
	// and should return ErrUnsupported for now
	result, err := driver.Get(ctx, common.SkipAuditLogDatabase)

	assert.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.MethodNotAllowedCode))
	assert.Nil(t, result)

	// Verify the error type is correct
	assert.True(t, errors.IsErr(err, errors.MethodNotAllowedCode))
}

// TestNewInMemoryManager tests that the NewInMemoryManager function works correctly
func TestNewInMemoryManager(t *testing.T) {
	manager := NewInMemoryManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.Store)

	// Test that the manager can be used (even though Get will return ErrUnsupported)
	ctx := context.Background()
	_, err := manager.GetItemFromDriver(ctx, common.SkipAuditLogDatabase)

	// Should return ErrUnsupported since the underlying driver's Get method returns this
	assert.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.MethodNotAllowedCode))
}
