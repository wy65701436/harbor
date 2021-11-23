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

package cosign

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/accessory/model/base"
)

// Signature signature model
type Signature struct {
	base.Default
}

// Kind gives the reference type of cosign signature.
func (c *Signature) Kind() string {
	return model.RefHard
}

// GetType ...
func (c *Signature) GetType() string {
	return model.TypeCosignSignature
}

// IsHard ...
func (c *Signature) IsHard() bool {
	return true
}

// GetIcon special icon for cosign signature
func (c *Signature) GetIcon() string {
	return ""
}

// New returns cosign signature
func New(opts model.Options) (model.Accessory, error) {
	return &Signature{base.Default{
		Opts: &opts,
	}}, nil
}

func init() {
	model.Register(model.TypeCosignSignature, New)
}
