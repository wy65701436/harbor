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

package blobquota

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
	"net/http"
)

type blobQuotaHandler struct {
	blobInfo *util.BlobInfo
	next     http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &blobQuotaHandler{
		next: next,
	}
}

// ServeHTTP ...
func (bqh blobQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPut {
		match, repository := util.MatchPutBlobURL(req)
		if match {
			bb := &util.BlobInfo{}
			bqh.blobInfo = bb
			bqh.blobInfo.Repository = repository

			ct := req.Header.Get("Content-Type")
			bqh.blobInfo.ContentType = ct

			dgstStr := req.FormValue("digest")
			if dgstStr == "" {
				http.Error(rw, util.MarshalError("StatusInternalServerError", "blob digest missing"), http.StatusInternalServerError)
				return
			}
			dgst, err := digest.Parse(dgstStr)
			if err != nil {
				http.Error(rw, util.MarshalError("StatusInternalServerError", "blob digest parsing failed"), http.StatusInternalServerError)
				return
			}

			bqh.blobInfo.Digest = dgst.String()

			projectID, err := util.GetProjectID(repository)
			if err != nil {
				log.Errorf("error occurred when to get project id for blob, %s", repository)
				http.Error(rw, util.MarshalError("StatusInternalServerError", fmt.Sprintf("error occurred when to get project id for blob, %s", repository)), http.StatusInternalServerError)
				return
			}
			bqh.blobInfo.ProjectID = projectID

			// try to require 1 kb for blob in case the storage is out of limit.
			quotaRes := &quota.ResourceList{
				quota.ResourceStorage: 1,
			}
			if err := util.TryRequireQuota(projectID, quotaRes); err != nil {
				log.Errorf("error occurred when to require quota for blob, %s", repository)
				http.Error(rw, util.MarshalError("StatusInternalServerError", fmt.Sprintf("error occurred when to require quota for blob, %s", repository)), http.StatusInternalServerError)
				return
			}
			bqh.blobInfo.Quota = quotaRes
		}
		*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, bqh.blobInfo)))
	}
	bqh.next.ServeHTTP(rw, req)
}
