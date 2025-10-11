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

package rest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/errors"
)

// RestDriverGetTestSuite tests the Get method in REST driver
type RestDriverGetTestSuite struct {
	suite.Suite
	ctx    context.Context
	driver *Driver
}

func (suite *RestDriverGetTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.driver = &Driver{
		configRESTURL: "http://localhost:8080/api/v2.0/configurations",
		client:        nil, // We'll test the unsupported behavior
	}
}

// TestGetMethodReturnsUnsupported tests that Get method returns ErrUnsupported
func (suite *RestDriverGetTestSuite) TestGetMethodReturnsUnsupported() {
	key := common.SkipAuditLogDatabase

	result, err := suite.driver.Get(suite.ctx, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)
}

// TestGetMethodWithDifferentKeys tests Get method with various keys
func (suite *RestDriverGetTestSuite) TestGetMethodWithDifferentKeys() {
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
func (suite *RestDriverGetTestSuite) TestGetMethodWithNilContext() {
	key := common.SkipAuditLogDatabase

	result, err := suite.driver.Get(nil, key)

	suite.Require().Error(err)
	suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
	suite.Nil(result)
}

// TestGetMethodConsistency tests that Get method consistently returns ErrUnsupported
func (suite *RestDriverGetTestSuite) TestGetMethodConsistency() {
	key := common.SkipAuditLogDatabase

	// Call multiple times to ensure consistency
	for i := 0; i < 5; i++ {
		result, err := suite.driver.Get(suite.ctx, key)

		suite.Require().Error(err)
		suite.True(errors.IsErr(err, errors.MethodNotAllowedCode))
		suite.Nil(result)
	}
}

// Run the test suite
func TestRestDriverGetTestSuite(t *testing.T) {
	suite.Run(t, new(RestDriverGetTestSuite))
}

// TestRestDriverGetStandalone provides additional standalone tests
func TestRestDriverGetStandalone(t *testing.T) {
	driver := &Driver{}
	ctx := context.Background()

	// Test that Get method is properly implemented and returns expected error
	result, err := driver.Get(ctx, "any_key")

	assert.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.MethodNotAllowedCode))
	assert.Nil(t, result)
}

// TestRestDriverGetDocumentation tests that the Get method behavior matches the TODO comment
func TestRestDriverGetDocumentation(t *testing.T) {
	driver := &Driver{}
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
