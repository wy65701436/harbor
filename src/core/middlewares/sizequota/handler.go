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
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/quota"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
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

type blobQuotaHandler struct {
	next     http.Handler
	blobInfo *util.BlobInfo
}

// New ...
func New(next http.Handler) http.Handler {
	return &blobQuotaHandler{
		next: next,
	}
}

// ServeHTTP ...
func (bqh *blobQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	matchPatchBlob, repository := util.MatchPatchBlobURL(req)
	if matchPatchBlob {
		bqh.blobInfo.Repository = repository
		bqh.handlePatchBlob(rw, req)
	}

	matchPutBlob, repository := util.MatchPutBlobURL(req)
	if matchPutBlob {
		bqh.blobInfo.Repository = repository
		bqh.handlePutBlobComplete(rw, req)
	}

	bqh.next.ServeHTTP(rw, req)
}

func (bqh *blobQuotaHandler) handlePatchBlob(rw http.ResponseWriter, req *http.Request) error {
	ct := req.Header.Get("Content-Type")
	if ct != "" && ct != "application/octet-stream" {
		return errors.New("unsupported content type")
	}

	bqh.blobInfo.UUID = getUUID(req.URL.Path)
	if bqh.blobInfo.UUID != "" {
		tempSize, err := strconv.ParseInt(req.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			return err
		}
		success, err := bqh.setBunkSize(tempSize)
		if err != nil {
			return err
		}
		if !success {
			// ToDo what to do here
		}
	}

	return nil
}

func (bqh *blobQuotaHandler) handlePutBlobComplete(rw http.ResponseWriter, req *http.Request) error {
	defer func() {
		if bqh.blobInfo.UUID != "" {
			_, err := bqh.removeUUID()
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
	bqh.blobInfo.Digest = dgst.String()
	bqh.blobInfo.UUID = getUUID(req.URL.Path)
	projectID, err := util.GetProjectID(strings.Split(bqh.blobInfo.Repository, "/")[0])
	if err != nil {
		return err
	}
	bqh.blobInfo.ProjectID = projectID

	digestLock, err := bqh.tryLockDigest()
	if err != nil {
		return err
	}
	bqh.blobInfo.DigestLock = digestLock

	blobExist, err := bqh.blobExist()
	if err != nil {
		return err
	}

	if !blobExist {
		size, err := bqh.getBlobSize()
		if err != nil {
			return err
		}
		quotaRes := &quota.ResourceList{
			quota.ResourceStorage: size,
		}
		err = util.TryRequireQuota(bqh.blobInfo.ProjectID, quotaRes)
		if err != nil {
			bqh.tryFreeDigest()
			log.Errorf("cannot get quota for the blob %v", err)
			return err
		}
		bqh.blobInfo.Quota = quotaRes
	}

	return nil
}

func getUUID(path string) string {
	strs := strings.Split(path, "/")
	return strs[len(strs)-1]
}

// check the existence of a blob in project
func (bqh *blobQuotaHandler) blobExist() (exist bool, err error) {
	return false, nil
}

func (bqh *blobQuotaHandler) removeUUID() (bool, error) {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return false, err
	}
	defer con.Close()

	exists, err := redis.Int(con.Do("EXISTS", bqh.blobInfo.UUID))
	if err != nil {
		return false, err
	}
	if exists == 1 {
		res, err := redis.Int(con.Do("DEL", bqh.blobInfo.UUID))
		if err != nil {
			return false, err
		}
		return res == 1, nil
	}
	return true, nil
}

// set the temp size for uuid.
func (bqh *blobQuotaHandler) setBunkSize(size int64) (bool, error) {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return false, err
	}

	defer con.Close()

	exists, err := redis.Int(con.Do("EXISTS", size))
	if err != nil {
		return false, err
	}
	if exists == 1 {
		curSize, err := redis.Int(con.Do("GET", bqh.blobInfo.UUID))
		if err != nil {
			return false, err
		}
		size += int64(curSize)
	}
	setRes, err := redis.String(con.Do("SET", bqh.blobInfo.UUID, size))
	if err != nil {
		return false, err
	}

	return setRes == "OK", nil

}

// get blob size for complete blob request
func (bqh *blobQuotaHandler) getBlobSize() (int64, error) {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return 0, err
	}

	defer con.Close()

	exists, err := redis.Int(con.Do("EXISTS", bqh.blobInfo.UUID))
	if err != nil {
		return 0, err
	}
	if exists == 1 {
		size, err := redis.Int64(con.Do("GET", bqh.blobInfo.UUID))
		if err != nil {
			return 0, err
		}
		return size, nil
	}

	return 0, errors.New("cannot get blob size")

}

// tryLockDigest locks blob with redis ...
func (bqh *blobQuotaHandler) tryLockDigest() (*common_redis.Mutex, error) {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return nil, err
	}
	digestLock := common_redis.New(con, bqh.blobInfo.Repository+":"+bqh.blobInfo.Digest, common_util.GenerateRandomString())
	success, err := digestLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock digest: %s, %s ", bqh.blobInfo.Repository, bqh.blobInfo.Digest)
	}
	return digestLock, nil
}

func (bqh *blobQuotaHandler) tryFreeDigest() {
	_, err := bqh.blobInfo.DigestLock.Free()
	if err != nil {
		log.Warningf("Error to unlock digest: %s,%s with error: %v ", bqh.blobInfo.Repository, bqh.blobInfo.Digest, err)
	}
}
