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

package policy

import (
	"github.com/goharbor/harbor/src/replication/model"
)

// Controller controls the replication policies
type Controller interface {
	// Create new rule
	Create(*model.Policy) (int64, error)
	// List the policies, returns the total count, rule list and error
	List(...*model.PolicyQuery) (int64, []*model.Policy, error)
	// Get rule with specified ID
	Get(int64) (*model.Policy, error)
	// Get rule by the name
	GetByName(string) (*model.Policy, error)
	// Update the specified rule
	Update(policy *model.Policy) error
	// Remove the specified rule
	Remove(int64) error
}
