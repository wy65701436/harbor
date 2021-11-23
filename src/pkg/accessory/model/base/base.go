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

package base

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"time"
)

var _ model.Accessory = (*Default)(nil)

// Default default model with TypeNone and RefNone
type Default struct {
	Opts *model.Options
}

// Kind ...
func (a *Default) Kind() string {
	return model.RefNone
}

// IsSoft ...
func (a *Default) IsSoft() bool {
	return false
}

// IsHard ...
func (a *Default) IsHard() bool {
	return false
}

// Type ...
func (a *Default) GetID() int64 {
	return a.Opts.ID
}

// GetType ...
func (a *Default) GetType() string {
	return model.TypeNone
}

// GetSize ...
func (a *Default) GetSize() int64 {
	return a.Opts.Size
}

// GetArtifactID ...
func (a *Default) GetArtifactID() int64 {
	return a.Opts.ArtifactID
}

// GetSubjectArtID ...
func (a *Default) GetSubjectArtID() int64 {
	return a.Opts.SubjectArtifactID
}

// GetDigest ...
func (a *Default) GetDigest() string {
	return a.Opts.Digest
}

// GetCreateTime ...
func (a *Default) GetCreateTime() time.Time {
	return a.Opts.CreationTime
}

// GetIcon ...
func (a *Default) GetIcon() string {
	return ""
}

// New returns base
func New(opts model.Options) (model.Accessory, error) {
	return &Default{Opts: &opts}, nil
}

func init() {
	model.Register(model.TypeNone, New)
}
