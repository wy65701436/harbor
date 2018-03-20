package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/ui/config"
)

// Record the docker deamon raw response.
var rec *httptest.ResponseRecorder

const (
	tokenUsername = "harbor-ui"
)

// NotaryEndpoint , exported for testing.
var NotaryEndpoint = config.InternalNotaryEndpoint()

type ContentTrustHandler struct {
	next http.Handler
}

func (cth ContentTrustHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(imageInfoCtxKey)
	if imgRaw == nil || !config.WithNotary() {
		cth.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(imageInfoCtxKey).(ImageInfo)
	if img.digest == "" {
		cth.next.ServeHTTP(rw, req)
		return
	}
	if !GetPolicyChecker().contentTrustEnabled(img.projectName) {
		cth.next.ServeHTTP(rw, req)
		return
	}
	match, err := matchNotaryDigest(img)
	if err != nil {
		http.Error(rw, marshalError("Failed in communication with Notary please check the log"), http.StatusInternalServerError)
		return
	}
	if !match {
		log.Debugf("digest mismatch, failing the response.")
		http.Error(rw, marshalError("The image is not signed in Notary."), http.StatusPreconditionFailed)
		return
	}
	cth.next.ServeHTTP(rw, req)
}

func matchNotaryDigest(img ImageInfo) (bool, error) {
	targets, err := notary.GetInternalTargets(NotaryEndpoint, tokenUsername, img.repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if isDigest(img.reference) {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if img.digest == d {
				return true, nil
			}
		} else {
			if t.Tag == img.reference {
				log.Debugf("found reference: %s in notary, try to match digest.", img.reference)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				if img.digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}

//A sha256 is a string with 64 characters.
func isDigest(ref string) bool {
	return strings.HasPrefix(ref, "sha256:") && len(ref) == 71
}
