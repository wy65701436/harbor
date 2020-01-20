package contenttrust

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/notary"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// NotaryEndpoint ...
var NotaryEndpoint = ""

// Middleware handle docker pull content trust check
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(rw, req)
		})
	}
}

func validate(req *http.Request) (bool, util.ImageInfo) {
	var img util.ImageInfo
	imgRaw := req.Context().Value(util.ImageInfoCtxKey)
	if imgRaw == nil || !config.WithNotary() {
		return false, img
	}
	img, _ = req.Context().Value(util.ImageInfoCtxKey).(util.ImageInfo)
	if img.Digest == "" {
		return false, img
	}
	if scannerPull, ok := util.ScannerPullFromContext(req.Context()); ok && scannerPull {
		return false, img
	}
	if !util.GetPolicyChecker().ContentTrustEnabled(img.ProjectName) {
		return false, img
	}
	return true, img
}

func matchNotaryDigest(mf middleware.ManifestInfo) (bool, error) {
	if NotaryEndpoint == "" {
		NotaryEndpoint = config.InternalNotaryEndpoint()
	}
	targets, err := notary.GetInternalTargets(NotaryEndpoint, util.TokenUsername, mf.Repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if mf.Digest != "" {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if mf.Digest == d {
				return true, nil
			}
		} else {
			if t.Tag == mf.Tag {
				log.Debugf("found reference: %s in notary, try to match digest.", mf.Tag)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				//ToDo get the digest
				if mf.Digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}
