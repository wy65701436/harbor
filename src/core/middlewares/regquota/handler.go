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
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
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
		mfInfo := &util.MfInfo{
			Repository: repository,
			Tag:        tag,
		}
		rqh.mfInfo = mfInfo

		mediaType := req.Header.Get("Content-Type")
		if mediaType == schema1.MediaTypeManifest ||
			mediaType == schema1.MediaTypeSignedManifest ||
			mediaType == schema2.MediaTypeManifest {

			tagLock, err := rqh.tryLockTag()
			if err != nil {
				log.Warningf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)), http.StatusInternalServerError)
				return
			}
			rqh.mfInfo.TagLock = tagLock

			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				rqh.tryFreeTag()
				log.Warningf("Error occurred when to copy manifest body %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to decode manifest body %v", err)), http.StatusInternalServerError)
				return
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

			manifest, desc, err := distribution.UnmarshalManifest(mediaType, data)
			if err != nil {
				rqh.tryFreeTag()
				log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to Unmarshal Manifest %v", err)), http.StatusInternalServerError)
				return
			}
			rqh.mfInfo.Refrerence = manifest.References()
			rqh.mfInfo.Digest = desc.Digest.String()
			rqh.mfInfo.Size = desc.Size
			log.Infof("manifest url... %s", req.URL.Path)
			log.Infof("manifest content type... %s", mediaType)
			log.Infof("manifest digest... %s", rqh.mfInfo.Digest)
			log.Infof("manifest size... %v", rqh.mfInfo.Size)
			log.Infof("manifest References... %v", rqh.mfInfo.Refrerence)

			projectID, err := rqh.getProjectID(strings.Split(repository, "/")[0])
			if err != nil {
				log.Warningf("Error occurred when to get project ID %v", err)
				return
			}
			rqh.mfInfo.ProjectID = projectID

			exist, af, err := rqh.imageExist()
			if err != nil {
				rqh.tryFreeTag()
				log.Warningf("Error occurred when to check Manifest existence %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to check Manifest existence %v", err)), http.StatusInternalServerError)
				return
			}
			rqh.mfInfo.Exist = exist
			if exist {
				if af.Digest != rqh.mfInfo.Digest {
					rqh.mfInfo.DigestChanged = true
				}
			}
			err = rqh.tryRequireQuota()
			if err != nil {
				rqh.tryFreeTag()
				log.Errorf("Cannot get quota for the manifest %v", err)
				http.Error(rw, util.MarshalError("StatusNotAcceptable", fmt.Sprintf("Cannot get quota for the manifest %v", err)), http.StatusNotAcceptable)
				return
			}

			*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, mfInfo)))
		}

	}

	rqh.next.ServeHTTP(rw, req)
}

// tryLockTag locks tag with redis ...
func (rqh *regQuotaHandler) tryLockTag() (*common_redis.Mutex, error) {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return nil, err
	}
	tagLock := common_redis.New(con, rqh.mfInfo.Repository+":"+rqh.mfInfo.Tag, common_util.GenerateRandomString())
	success, err := tagLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock tag: %s ", rqh.mfInfo.Repository+":"+rqh.mfInfo.Tag)
	}
	return tagLock, nil
}

func (rqh *regQuotaHandler) tryFreeTag() {
	_, err := rqh.mfInfo.TagLock.Free()
	if err != nil {
		log.Warningf("Error to unlock tag: %s, with error: %v ", rqh.mfInfo.Tag, err)
	}
}

// check the existence of a artifact, if exist, the method will return the artifact model
func (rqh *regQuotaHandler) imageExist() (exist bool, af *models.Artifact, err error) {
	artifactQuery := &models.ArtifactQuery{
		PID:  rqh.mfInfo.ProjectID,
		Repo: rqh.mfInfo.Repository,
		Tag:  rqh.mfInfo.Tag,
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

func (rqh *regQuotaHandler) tryRequireQuota() error {
	var quotaRes quota.ResourceList
	quotaRes = quota.ResourceList{
		quota.ResourceStorage: 0,
		quota.ResourceCount:   0,
	}

	if !rqh.mfInfo.Exist {
		// For manifest, we need to request 1 count and size for storage.
		quotaRes = quota.ResourceList{
			quota.ResourceStorage: rqh.mfInfo.Size,
			quota.ResourceCount:   1,
		}
	} else if rqh.mfInfo.DigestChanged {
		quotaRes = quota.ResourceList{
			quota.ResourceStorage: rqh.mfInfo.Size,
			quota.ResourceCount:   0,
		}
	}

	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(rqh.mfInfo.ProjectID, 10))
	if err != nil {
		log.Errorf("Error occurred when to new quota manager %v", err)
		return err
	}

	if err := quotaMgr.AddResources(quotaRes); err != nil {
		log.Errorf("Cannot get quota for the manifest %v", err)
		return err
	}
	return nil
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
