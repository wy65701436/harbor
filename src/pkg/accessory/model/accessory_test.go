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

package model

import (
	"testing"

	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory/model"
	"github.com/stretchr/testify/suite"
)

type AccessoryTestSuite struct {
	suite.Suite
}

func (suite *AccessoryTestSuite) SetupSuite() {
	Register("mock", func(opts Options) (Accessory, error) {
		return &accessorytesting.Accessory{}, nil
	})
}

func (suite *AccessoryTestSuite) TestNew() {
	{
		c, err := New("")
		suite.Nil(c)
		suite.Error(err)
	}

	{
		c, err := New("mocks")
		suite.Nil(c)
		suite.Error(err)
	}

	{
		c, err := New("mock")
		suite.NotNil(c)
		suite.Nil(err)
	}
}

func TestAccessoryTestSuite(t *testing.T) {
	suite.Run(t, new(AccessoryTestSuite))
}
