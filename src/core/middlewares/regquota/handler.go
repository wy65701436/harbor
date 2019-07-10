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

package regquota

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	common_util "github.com/goharbor/review/harbor/src/common/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

type regQuotaHandler struct {
	next   http.Handler
	mfInfo *util.MfInfo
}

// New ...
func New(next http.Handler) http.Handler {
	return &regQuotaHandler{
		next: next,
	}
}

// ServeHTTP manifest ...
func (rqh *regQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, repository, tag := util.MatchPushManifest(req)
	if match {
		mfInfo := util.MfInfo{
			Repository: repository,
			Tag:        tag,
		}
		rqh.mfInfo = &mfInfo

		mediaType := req.Header.Get("Content-Type")
		if mediaType == schema1.MediaTypeManifest ||
			mediaType == schema1.MediaTypeSignedManifest ||
			mediaType == schema2.MediaTypeManifest {

			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Warningf("Error occurred when to copy manifest body %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to decode manifest body %v", err)), http.StatusInternalServerError)
				return
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

			manifest, desc, err := distribution.UnmarshalManifest(mediaType, data)
			if err != nil {
				log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to Unmarshal Manifest %v", err)), http.StatusInternalServerError)
				return
			}
			rqh.mfInfo.Refrerence = manifest.References()
			rqh.mfInfo.Digest = desc.Digest.String()
			rqh.mfInfo.Size = desc.Size
			log.Infof("manifest digest... %s", rqh.mfInfo.Digest)
			log.Infof("manifest size... %v", rqh.mfInfo.Size)

			projectID, err := rqh.getProjectID(strings.Split(repository, "/")[0])
			if err != nil {
				log.Warningf("Error occurred when to get project ID %v", err)
				return
			}
			rqh.mfInfo.ProjectID = projectID
			mfExist, err := rqh.mfExist()
			if !mfExist {
				tagLock, err := rqh.tryLockTag(req)
				if err != nil {
					log.Warningf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)
					http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)), http.StatusInternalServerError)
					return
				}
				rqh.mfInfo.TagLock = tagLock

				tagExist, err := rqh.tagExist()
				if err != nil {
					log.Warningf("Error occurred when to check Manifest existence %v", err)
					http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to check Manifest existence %v", err)), http.StatusInternalServerError)
					return
				}

				hasEnoughQuota := rqh.tryRequireQuota(desc.Size, tagExist)
				if !hasEnoughQuota {
					_, err := rqh.mfInfo.TagLock.Free()
					if err != nil {
						log.Warningf("Error to unlock tag: %s, with error: %v ", rqh.mfInfo.Tag, err)
					}
					log.Warningf("Cannot get quota for the manifest %v", err)
					http.Error(rw, util.MarshalError("StatusNotAcceptable", fmt.Sprintf("Cannot get quota for the manifest %v", err)), http.StatusNotAcceptable)
					return
				}
			}

			*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, mfInfo)))
		}

	}

	rqh.next.ServeHTTP(rw, req)
}

// tryLockTag locks tag with redis ...
func (rqh *regQuotaHandler) tryLockTag(request *http.Request) (*common_redis.Mutex, error) {
	redisUrl, err := config.RegistryURL()
	if err != nil {
		log.Warningf("Error occurred when to lock tag with redis %v", err)
		return nil, err
	}
	con, err := redis.DialURL(
		redisUrl,
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return nil, err
	}
	defer con.Close()

	tagLock := common_redis.New(con, rqh.mfInfo.Repository+":"+rqh.mfInfo.Tag, common_util.GenerateRandomString())
	lockSuccess, err := tagLock.Require()
	if !lockSuccess {
		return nil, err
	}
	return tagLock, nil
}

func (rqh *regQuotaHandler) getProjectID(name string) (int64, error) {
	project, err := dao.GetProjectByName(name)
	if err != nil {
		return 0, err
	}
	if project != nil {
		return project.ProjectID, nil
	}
	return 0, fmt.Errorf("project %s is not found", name)
}

func (rqh *regQuotaHandler) tagExist() (bool, error) {
	artifactQuery := &models.ArtifactQuery{
		ProjectID: rqh.mfInfo.ProjectID,
	}
	afs, err := dao.ListArtifacts(artifactQuery)
	if err != nil {
		log.Errorf("Error occurred when to get project ID %v", err)
		return false, err
	}

	for _, af := range afs {
		if af.Repo == rqh.mfInfo.Repository && af.Tag == rqh.mfInfo.Tag {
			return true, nil
		}
	}
	return false, nil
}

func (rqh *regQuotaHandler) mfExist() (bool, error) {
	artifactQuery := &models.ArtifactQuery{
		ProjectID: rqh.mfInfo.ProjectID,
	}
	afs, err := dao.ListArtifacts(artifactQuery)
	if err != nil {
		log.Errorf("Error occurred when to get project ID %v", err)
		return false, err
	}

	for _, af := range afs {
		if af.Digest == rqh.mfInfo.Digest {
			return true, nil
		}
	}
	return false, nil
}

func (rqh *regQuotaHandler) tryRequireQuota(quotaSize int64, tagExist bool) bool {
	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(rqh.mfInfo.ProjectID, 10))
	if err != nil {
		log.Errorf("Error occurred when to new quota manager %v", err)
		return false
	}
	var quotaRes quota.ResourceList
	if tagExist {
		quotaRes = quota.ResourceList{
			quota.ResourceStorage: quotaSize,
		}
	} else {
		quotaRes = quota.ResourceList{
			quota.ResourceStorage: quotaSize,
			quota.ResourceCount:   1,
		}
	}

	if err := quotaMgr.AddResources(quotaRes); err != nil {
		log.Errorf("Cannot get quota for the manifest %v", err)
		return false
	}
	return true
}
