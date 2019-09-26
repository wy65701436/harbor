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

package immutable

import (
	"fmt"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"net/http"
)

type immutableHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &immutableHandler{
		next: next,
	}
}

// ServeHTTP ...
func (rh immutableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if match, _, _ := util.MatchPushManifest(req); !match {
		return
	}
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		var err error
		info, err = util.ParseManifestInfoFromPath(req)
		if err != nil {
			return
		}
	}

	isImmutableTag, err := rule.NewRuleMatcher(info.ProjectID).Match(art.Candidate{
		Repository: info.Repository,
		Tag:        info.Tag,
	})
	if err != nil {
		return
	}

	if isImmutableTag {
		http.Error(rw, util.MarshalError("DENIED",
			fmt.Sprintf("The tag:%s:%s is immutable, cannot be overwrite.", info.Repository, info.Tag)), http.StatusForbidden)
		return
	}

	rh.next.ServeHTTP(rw, req)
}
