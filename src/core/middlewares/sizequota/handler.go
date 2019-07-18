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

package sizequota

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/quota"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type sizeQuotaHandler struct {
	next     http.Handler
	blobInfo *util.BlobInfo
	mfInfo   *util.MfInfo
}

// New ...
func New(next http.Handler) http.Handler {
	return &sizeQuotaHandler{
		next: next,
	}
}

// ServeHTTP ...
func (sqh *sizeQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	matchPutBlob, repository := util.MatchPutBlobURL(req)
	if matchPutBlob {
		bb := &util.BlobInfo{}
		sqh.blobInfo = bb
		sqh.blobInfo.Repository = repository

		if err := sqh.handlePutBlobComplete(rw, req); err != nil {
			log.Warningf("Error occurred when to handle put blob %v", err)
			http.Error(rw, util.MarshalError("InternalError", "Error occurred when to handle put blob"), http.StatusInternalServerError)
			return
		}
	}

	matchPushMF, repository, tag := util.MatchPushManifest(req)
	if matchPushMF {
		bb := &util.BlobInfo{}
		sqh.blobInfo = bb
		mfInfo := &util.MfInfo{}
		sqh.mfInfo = mfInfo
		sqh.blobInfo.Repository = repository
		sqh.mfInfo.Repository = repository
		sqh.mfInfo.Tag = tag

		if err := sqh.handlePutManifest(rw, req); err != nil {
			log.Warningf("Error occurred when to handle put manifest %v", err)
			http.Error(rw, util.MarshalError("InternalError", "Error occurred when to handle put manifest"), http.StatusInternalServerError)
			return
		}
	}

	sqh.next.ServeHTTP(rw, req)
}

func (sqh *sizeQuotaHandler) handlePutManifest(rw http.ResponseWriter, req *http.Request) error {
	mediaType := req.Header.Get("Content-Type")
	if mediaType == schema1.MediaTypeManifest ||
		mediaType == schema1.MediaTypeSignedManifest ||
		mediaType == schema2.MediaTypeManifest {

		log.Info("^^^^^^^^^^^^^^^^^")
		con, err := util.GetRegRedisCon()
		if err != nil {
			log.Infof("failed to get registry redis connection, %v", err)
			return err
		}
		log.Info("^^^^^^^^^^^^^^^^^")

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Warningf("Error occurred when to copy manifest body %v", err)
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		manifest, desc, err := distribution.UnmarshalManifest(mediaType, data)
		if err != nil {
			log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
			return err
		}
		projectID, err := util.GetProjectID(strings.Split(sqh.mfInfo.Repository, "/")[0])
		if err != nil {
			log.Warningf("Error occurred when to get project ID %v", err)
			return err
		}
		log.Info("^^^^^^^^^^^^^^^^^")

		sqh.mfInfo.ProjectID = projectID
		sqh.mfInfo.Refrerence = manifest.References()
		sqh.mfInfo.Digest = desc.Digest.String()
		sqh.blobInfo.ProjectID = projectID
		sqh.blobInfo.Digest = desc.Digest.String()
		sqh.blobInfo.Size = desc.Size
		sqh.blobInfo.ContentType = mediaType

		if err := sqh.requireQuota(con); err != nil {
			return err
		}
		log.Info("^^^^^^^^^^^^^^^^^")

		*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, sqh.mfInfo)))
		*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, sqh.blobInfo)))
	}

	return fmt.Errorf("unsupported content type for manifest: %s", mediaType)
}

func (sqh *sizeQuotaHandler) handlePutBlobComplete(rw http.ResponseWriter, req *http.Request) error {
	con, err := util.GetRegRedisCon()
	if err != nil {
		return err
	}

	defer func() {
		if sqh.blobInfo.UUID != "" {
			_, err := sqh.removeUUID(con)
			if err != nil {
				log.Warningf("error occurred when remove UUID for blob, %v", err)
			}
		}
	}()

	dgstStr := req.FormValue("digest")
	if dgstStr == "" {
		return errors.New("blob digest missing")
	}
	dgst, err := digest.Parse(dgstStr)
	if err != nil {
		return errors.New("blob digest parsing failed")
	}

	sqh.blobInfo.Digest = dgst.String()
	sqh.blobInfo.UUID = getUUID(req.URL.Path)
	size, err := util.GetBlobSize(con, sqh.blobInfo.UUID)
	if err != nil {
		return err
	}
	sqh.blobInfo.Size = size
	if err := sqh.requireQuota(con); err != nil {
		return err
	}
	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, sqh.blobInfo)))
	return nil

}

func (sqh *sizeQuotaHandler) requireQuota(conn redis.Conn) error {
	log.Info("111 ^^^^^^^^^^^^^^^^^^")
	projectID, err := util.GetProjectID(strings.Split(sqh.blobInfo.Repository, "/")[0])
	if err != nil {
		return err
	}
	sqh.blobInfo.ProjectID = projectID

	log.Info("111 ^^^^^^^^^^^^^^^^^^")
	digestLock, err := sqh.tryLockBlob(conn)
	if err != nil {
		log.Infof("failed to lock digest in redis, %v", err)
		return err
	}
	sqh.blobInfo.DigestLock = digestLock

	log.Info("111 ^^^^^^^^^^^^^^^^^^")
	blobExist, err := dao.HasBlobInProject(sqh.blobInfo.ProjectID, sqh.blobInfo.Digest)
	if err != nil {
		sqh.tryFreeBlob()
		return err
	}
	log.Info("111 ^^^^^^^^^^^^^^^^^^")

	if !blobExist {
		quotaRes := &quota.ResourceList{
			quota.ResourceStorage: sqh.blobInfo.Size,
		}
		log.Info("111 ^^^^^^^^^^^^^^^^^^")
		err = util.TryRequireQuota(sqh.blobInfo.ProjectID, quotaRes)
		if err != nil {
			log.Infof("project id, %d, size %d", sqh.blobInfo.ProjectID, sqh.blobInfo.Size)
			sqh.tryFreeBlob()
			log.Errorf("cannot get quota for the blob %v", err)
			return err
		}
		sqh.blobInfo.Quota = quotaRes
	}

	return nil
}

func (sqh *sizeQuotaHandler) removeUUID(conn redis.Conn) (bool, error) {
	exists, err := redis.Int(conn.Do("EXISTS", sqh.blobInfo.UUID))
	if err != nil {
		return false, err
	}
	if exists == 1 {
		res, err := redis.Int(conn.Do("DEL", sqh.blobInfo.UUID))
		if err != nil {
			return false, err
		}
		return res == 1, nil
	}
	return true, nil
}

// tryLockBlob locks blob with redis ...
func (sqh *sizeQuotaHandler) tryLockBlob(conn redis.Conn) (*common_redis.Mutex, error) {
	log.Info("tryLockBlob ............")
	digestLock := common_redis.New(conn, sqh.blobInfo.Repository+":"+sqh.blobInfo.Digest, common_util.GenerateRandomString())
	success, err := digestLock.Require()
	log.Infof("tryLockBlob ............ %s", success)
	log.Infof("tryLockBlob ............ %v", err)
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock digest: %s, %s ", sqh.blobInfo.Repository, sqh.blobInfo.Digest)
	}
	return digestLock, nil
}

func (sqh *sizeQuotaHandler) tryFreeBlob() {
	_, err := sqh.blobInfo.DigestLock.Free()
	if err != nil {
		log.Warningf("Error to unlock digest: %s,%s with error: %v ", sqh.blobInfo.Repository, sqh.blobInfo.Digest, err)
	}
}

// put blob path: /v2/<name>/blobs/uploads/<uuid>
func getUUID(path string) string {
	strs := strings.Split(path, "/")
	return strs[len(strs)-1]
}
