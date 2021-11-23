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

import "time"

// Option function to set the options of the cache
type Option func(*Options)

// Options options of the cache
type Options struct {
	ID                int64
	ArtifactID        int64
	SubjectArtifactID int64
	Size              int64
	Digest            string
	CreationTime      time.Time
}

func newOptions(opt ...Option) Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}

// ID ...
func ID(id int64) Option {
	return func(o *Options) {
		o.ID = id
	}
}

// ArtifactID ...
func ArtifactID(artifactID int64) Option {
	return func(o *Options) {
		o.ArtifactID = artifactID
	}
}

// SubjectArtifactID ...
func SubjectArtifactID(subArtID int64) Option {
	return func(o *Options) {
		o.SubjectArtifactID = subArtID
	}
}

// Size ...
func Size(size int64) Option {
	return func(o *Options) {
		o.Size = size
	}
}

// Digest ...
func Digest(digest string) Option {
	return func(o *Options) {
		o.Digest = digest
	}
}

// CreationTime ...
func CreationTime(creationTime time.Time) Option {
	return func(o *Options) {
		o.CreationTime = creationTime
	}
}
