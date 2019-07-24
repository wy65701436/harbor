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
	"context"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
	"net/http"
	"time"
)

// PutManifestInterceptor ...
type PutBlobInterceptor struct {
	blobInfo *util.BlobInfo
}

// NewMountBlobInterceptor ...
func NewPutBlobInterceptor(blobInfo *util.BlobInfo) *PutBlobInterceptor {
	return &PutBlobInterceptor{
		blobInfo: blobInfo,
	}
}

func (pbi *PutBlobInterceptor) handleRequest(req *http.Request) error {
	// the redis connection will be closed in the put response.
	con, err := util.GetRegRedisCon()
	if err != nil {
		return err
	}

	defer func() {
		if pbi.blobInfo.UUID != "" {
			_, err := rmBlobUploadUUID(con, pbi.blobInfo.UUID)
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

	pbi.blobInfo.Digest = dgst.String()
	pbi.blobInfo.UUID = getUUID(req.URL.Path)
	size, err := util.GetBlobSize(con, pbi.blobInfo.UUID)
	if err != nil {
		return err
	}
	pbi.blobInfo.Size = size
	if err := requireQuota(con, pbi.blobInfo); err != nil {
		return err
	}
	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, pbi.blobInfo)))
	return nil
}

func (pbi *PutBlobInterceptor) handleResponse(rw util.CustmoResponseWriter, req *http.Request) error {
	bbInfo := req.Context().Value(util.BBInfokKey)
	bb, ok := bbInfo.(*util.BlobInfo)
	if !ok {
		return errors.New("failed to convert blob information context into BBInfo")
	}
	defer func() {
		_, err := bb.DigestLock.Free()
		if err != nil {
			log.Errorf("Error to unlock blob digest:%s in response handler, %v", bb.Digest, err)
		}
		if err := bb.DigestLock.Conn.Close(); err != nil {
			log.Errorf("Error to close redis connection in put blob response handler, %v", err)
		}
	}()

	if rw.Status() == http.StatusCreated {
		if !bb.Exist {
			blob := &models.Blob{
				Digest:       bb.Digest,
				ContentType:  bb.ContentType,
				Size:         bb.Size,
				CreationTime: time.Now(),
			}
			_, err := dao.AddBlob(blob)
			if err != nil {
				return err
			}
		}
	} else if rw.Status() >= 300 || rw.Status() <= 511 {
		if !bb.Exist {
			success := util.TryFreeQuota(bb.ProjectID, bb.Quota)
			if !success {
				return fmt.Errorf("Error to release resource booked for the blob, %d, digest: %s ", bb.ProjectID, bb.Digest)
			}
		}
	}
	return nil
}
