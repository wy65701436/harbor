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
	"github.com/goharbor/harbor/src/core/config"
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
}

// New ...
func New(next http.Handler) http.Handler {
	return &sizeQuotaHandler{
		next: next,
	}
}

// ServeHTTP ...
func (sqh *sizeQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	bb := &util.BlobInfo{}
	sqh.blobInfo = bb

	matchPutBlob, repository := util.MatchPatchBlobURL(req)
	if matchPutBlob {
		sqh.blobInfo.Repository = repository
		sqh.handlePutBlobComplete(rw, req)
	}

	//matchMF, repository, _ := util.MatchPushManifest(req)
	//if matchMF {
	//	sqh.blobInfo.Repository = repository
	//	sqh.handlePutManifest(rw, req)
	//}

	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, bb)))

	sqh.next.ServeHTTP(rw, req)
}

func (sqh *sizeQuotaHandler) handlePutManifest(rw http.ResponseWriter, req *http.Request) error {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(util.DialConnectionTimeout),
		redis.DialReadTimeout(util.DialReadTimeout),
		redis.DialWriteTimeout(util.DialWriteTimeout),
	)
	if err != nil {
		return err
	}
	defer con.Close()

	mediaType := req.Header.Get("Content-Type")
	if mediaType == schema1.MediaTypeManifest ||
		mediaType == schema1.MediaTypeSignedManifest ||
		mediaType == schema2.MediaTypeManifest {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Warningf("Error occurred when to copy manifest body %v", err)
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		_, desc, err := distribution.UnmarshalManifest(mediaType, data)
		if err != nil {
			log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
			return err
		}
		sqh.blobInfo.Digest = desc.Digest.String()
		sqh.blobInfo.Size = desc.Size
		return sqh.requireQuota(con)
	}

	return errors.New("unsupported content type")
}

func (sqh *sizeQuotaHandler) handlePutBlobComplete(rw http.ResponseWriter, req *http.Request) error {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(util.DialConnectionTimeout),
		redis.DialReadTimeout(util.DialReadTimeout),
		redis.DialWriteTimeout(util.DialWriteTimeout),
	)
	if err != nil {
		return err
	}

	defer func() {
		//if sqh.blobInfo.UUID != "" {
		//	_, err := sqh.removeUUID(con)
		//	if err != nil {
		//		log.Warningf("error occurred when remove UUID for blob, %v", err)
		//	}
		//}
		con.Close()
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
	log.Info("111111111111111")
	log.Info(sqh.blobInfo.UUID)
	log.Info("111111111111111")
	size, err := util.GetBlobSize(con, sqh.blobInfo.UUID)
	if err != nil {
		return err
	}
	sqh.blobInfo.Size = size
	return sqh.requireQuota(con)

}

func (sqh *sizeQuotaHandler) requireQuota(conn redis.Conn) error {
	projectID, err := util.GetProjectID(strings.Split(sqh.blobInfo.Repository, "/")[0])
	if err != nil {
		return err
	}
	sqh.blobInfo.ProjectID = projectID

	digestLock, err := sqh.tryLockDigest(conn)
	if err != nil {
		return err
	}
	sqh.blobInfo.DigestLock = digestLock

	blobExist, err := sqh.blobExist()
	if err != nil {
		sqh.tryFreeDigest()
		return err
	}

	if !blobExist {
		quotaRes := &quota.ResourceList{
			quota.ResourceStorage: sqh.blobInfo.Size,
		}
		err = util.TryRequireQuota(sqh.blobInfo.ProjectID, quotaRes)
		if err != nil {
			sqh.tryFreeDigest()
			log.Errorf("cannot get quota for the blob %v", err)
			return err
		}
		sqh.blobInfo.Quota = quotaRes
	}

	return nil
}

// check the existence of a blob in project
func (sqh *sizeQuotaHandler) blobExist() (exist bool, err error) {
	return dao.HasBlobInProject(sqh.blobInfo.ProjectID, sqh.blobInfo.Digest)
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

// tryLockDigest locks blob with redis ...
func (sqh *sizeQuotaHandler) tryLockDigest(conn redis.Conn) (*common_redis.Mutex, error) {
	digestLock := common_redis.New(conn, sqh.blobInfo.Repository+":"+sqh.blobInfo.Digest, common_util.GenerateRandomString())
	success, err := digestLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock digest: %s, %s ", sqh.blobInfo.Repository, sqh.blobInfo.Digest)
	}
	return digestLock, nil
}

func (sqh *sizeQuotaHandler) tryFreeDigest() {
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
