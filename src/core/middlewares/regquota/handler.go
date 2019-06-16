package regquota

import (
	"fmt"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

type regQuotaHandler struct {
	next http.Handler
}

func New(next http.Handler) (http.Handler, error) {
	return &regQuotaHandler{
		next: next,
	}, nil
}

//PATCH manifest ...
func (bh regQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, _, _ := util.MatchManifestURL(req)
	if match {
		//var imageSize int64
		//var imageDigest string
		//mediaType := req.Header.Get("Content-Type")
		//if req.Method == http.MethodPut && mediaType == "application/vnd.docker.distribution.manifest.v2+json" {
		//	data, err := ioutil.ReadAll(req.Body)
		//	if err != nil {
		//		log.Warningf("Error occured when to copy manifest body %v", err)
		//		http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occured when to decode manifest body %v", err)), http.StatusInternalServerError)
		//		return
		//	}
		//
		//	manifest, desc, err := distribution.UnmarshalManifest(mediaType, data)
		//	if err != nil {
		//		log.Warningf("Error occured when to Unmarshal Manifest %v", err)
		//		http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occured when to Unmarshal Manifest %v", err)), http.StatusInternalServerError)
		//		return
		//	}
		//	for _, ref := range manifest.References() {
		//		imageSize += ref.Size
		//	}
		//	imageDigest = desc.Digest.String()
		//	imageSize += desc.Size
		//	log.Infof("manifest digest... %s", imageDigest)
		//	log.Infof("manifest size... %v", imageSize)
		//	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		//}
		http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occured when to Unmarshal Manifest")), http.StatusInternalServerError)
		return
	}

	bh.next.ServeHTTP(rw, req)
}
