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

package registryproxy

import (
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type proxyHandler struct {
	handler http.Handler
}

// New ...
func New(urls ...string) http.Handler {
	var registryURL string
	var err error
	if len(urls) > 1 {
		log.Errorf("the parm, urls should have only 0 or 1 elements")
		return nil
	}
	if len(urls) == 0 {
		registryURL, err = config.RegistryURL()
		if err != nil {
			log.Error(err)
			return nil
		}
	} else {
		registryURL = urls[0]
	}
	targetURL, err := url.Parse(registryURL)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &proxyHandler{
		handler: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				director(targetURL, req)
			},
			ModifyResponse: modifyResponse,
		},
	}

}

// Overwrite the http requests
func director(target *url.URL, req *http.Request) {
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

// Modify the http response
func modifyResponse(res *http.Response) error {
	matchMF, _, _ := util.MatchPushManifest(res.Request)
	if matchMF {
		return handlePutManifest(res)
	}
	matchBB, _ := util.MatchPutBlobURL(res.Request)
	if matchBB {
		return handlePutBlob(res)
	}
	matchPatchBlob, _ := util.MatchPatchBlobURL(res.Request)
	if matchPatchBlob {
		return handlePatchBlob(res)
	}
	return nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func handlePutManifest(res *http.Response) error {
	mfInfo := res.Request.Context().Value(util.MFInfokKey)
	mf, ok := mfInfo.(*util.MfInfo)
	if !ok {
		return errors.New("failed to convert manifest information context into MfInfo")
	}

	defer func() {
		_, err := mf.TagLock.Free()
		if err != nil {
			log.Errorf("Error to unlock in response handler, %v", err)
		}
		if err := mf.TagLock.Conn.Close(); err != nil {
			log.Errorf("Error to close redis connection in response handler, %v", err)
		}
	}()

	// 201
	if res.StatusCode == http.StatusCreated {
		af := &models.Artifact{
			PID:      mf.ProjectID,
			Repo:     mf.Repository,
			Tag:      mf.Tag,
			Digest:   mf.Digest,
			PushTime: time.Now(),
			Kind:     "Docker-Image",
		}

		// insert or update
		if !mf.Exist {
			_, err := dao.AddArtifact(af)
			if err != nil {
				log.Errorf("Error to add artifact, %v", err)
				return err
			}
		}
		if mf.DigestChanged {
			err := dao.UpdateArtifactDigest(af)
			if err != nil {
				log.Errorf("Error to add artifact, %v", err)
				return err
			}
		}

		if !mf.Exist || mf.DigestChanged {
			afnbs := []*models.ArtifactAndBlob{}
			self := &models.ArtifactAndBlob{
				DigestAF:   mf.Digest,
				DigestBlob: mf.Digest,
			}
			afnbs = append(afnbs, self)
			for _, d := range mf.Refrerence {
				afnb := &models.ArtifactAndBlob{
					DigestAF:   mf.Digest,
					DigestBlob: d.Digest.String(),
				}
				afnbs = append(afnbs, afnb)
			}
			if err := dao.AddArtifactNBlobs(afnbs); err != nil {
				log.Errorf("Error to add artifact and blobs in proxy response handler, %v", err)
				return err
			}
		}

	} else if res.StatusCode >= 300 || res.StatusCode <= 511 {
		if !mf.Exist {
			success := util.TryFreeQuota(mf.ProjectID, mf.Quota)
			if !success {
				return errors.New("Error to release resource booked for the manifest")
			}
		}
	}

	return nil
}

// handle put blob complete request
// 1, add blob into DB if success
// 2, roll back resource if failure.
func handlePutBlob(res *http.Response) error {
	bbInfo := res.Request.Context().Value(util.BBInfokKey)
	bb, ok := bbInfo.(*util.BlobInfo)
	if !ok {
		return errors.New("failed to convert blob information context into BBInfo")
	}

	defer func() {
		_, err := bb.DigestLock.Free()
		if err != nil {
			log.Errorf("Error to unlock in response handler, %v", err)
		}
		if err := bb.DigestLock.Conn.Close(); err != nil {
			log.Errorf("Error to close redis connection in response handler, %v", err)
		}
	}()

	if res.StatusCode != http.StatusCreated {
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
	} else if res.StatusCode >= 300 || res.StatusCode <= 511 {
		success := util.TryFreeQuota(bb.ProjectID, bb.Quota)
		if !success {
			return errors.New("Error to release resource booked for the manifest")
		}
	}
	return nil
}

// Do record bunk size on success, registry will return an 202 for PATCH success, and with an UUID.
func handlePatchBlob(res *http.Response) error {
	if res.StatusCode == http.StatusAccepted {
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

		uuid := res.Header.Get("Docker-Upload-UUID")
		cl, err := strconv.ParseInt(res.Request.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			return err
		}
		success, err := setBunkSize(con, uuid, cl)
		if err != nil {
			return err
		}
		if !success {
			//ToDo discuss what to do here.
		}
	}
	return nil
}

// set the temp size for uuid.
func setBunkSize(conn redis.Conn, uuid string, size int64) (bool, error) {
	size, err := util.GetBlobSize(conn, uuid)
	if err != nil {
		return false, err
	}
	setRes, err := redis.String(conn.Do("SET", uuid, size))
	if err != nil {
		return false, err
	}

	return setRes == "OK", nil
}

// ServeHTTP ...
func (ph proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ph.handler.ServeHTTP(rw, req)
}
