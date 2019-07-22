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

package countquota

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

type countQuotaHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &countQuotaHandler{
		next: next,
	}
}

// ServeHTTP manifest ...
func (cqh *countQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, repository, tag := util.MatchPushManifest(req)
	if match {
		mfInfo := req.Context().Value(util.MFInfokKey)
		mf, ok := mfInfo.(*util.MfInfo)
		if !ok {
			http.Error(rw, util.MarshalError("InternalServerError", "Failed to get manifest infor from context"), http.StatusInternalServerError)
			return
		}

		tagLock, err := cqh.tryLockTag(mf)
		if err != nil {
			log.Warningf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)
			http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)), http.StatusInternalServerError)
			return
		}
		mf.TagLock = tagLock

		imageExist, af, err := cqh.imageExist(mf)
		if err != nil {
			cqh.tryFreeTag(mf)
			log.Warningf("Error occurred when to check Manifest existence by repo and tag name %v", err)
			http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to check Manifest existence %v", err)), http.StatusInternalServerError)
			return
		}
		mf.Exist = imageExist
		if imageExist {
			if af.Digest != mf.Digest {
				mf.DigestChanged = true
			}
		} else {
			quotaRes := &quota.ResourceList{
				quota.ResourceCount: 1,
			}
			err := util.TryRequireQuota(mf.ProjectID, quotaRes)
			if err != nil {
				cqh.tryFreeTag(mf)
				log.Errorf("Cannot get quota for the manifest %v", err)
				if err == util.ErrRequireQuota {
					http.Error(rw, util.MarshalError("StatusNotAcceptable", "Your request is reject as not enough quota."), http.StatusNotAcceptable)
					return
				}
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to require quota for the manifest %v", err)), http.StatusInternalServerError)
				return
			}
			mf.Quota = quotaRes
		}
		*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, mfInfo)))
	}

	cqh.next.ServeHTTP(rw, req)
}

// tryLockTag locks tag with redis ...
func (cqh *countQuotaHandler) tryLockTag(mfInfo *util.MfInfo) (*common_redis.Mutex, error) {
	con, err := util.GetRegRedisCon()
	if err != nil {
		return nil, err
	}
	tagLock := common_redis.New(con, "Quota::manifest-lock::"+mfInfo.Repository+":"+mfInfo.Tag, common_util.GenerateRandomString())
	success, err := tagLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock tag: %s ", mfInfo.Repository+":"+mfInfo.Tag)
	}
	return tagLock, nil
}

func (cqh *countQuotaHandler) tryFreeTag(mfInfo *util.MfInfo) {
	_, err := mfInfo.TagLock.Free()
	if err != nil {
		log.Warningf("Error to unlock tag: %s, with error: %v ", mfInfo.Tag, err)
	}
}

// check the existence of a artifact, if exist, the method will return the artifact model
func (cqh *countQuotaHandler) imageExist(mfInfo *util.MfInfo) (exist bool, af *models.Artifact, err error) {
	artifactQuery := &models.ArtifactQuery{
		PID:  mfInfo.ProjectID,
		Repo: mfInfo.Repository,
		Tag:  mfInfo.Tag,
	}
	afs, err := dao.ListArtifacts(artifactQuery)
	if err != nil {
		log.Errorf("Error occurred when to get project ID %v", err)
		return false, nil, err
	}
	if len(afs) > 0 {
		return true, afs[0], nil
	}
	return false, nil, nil
}
