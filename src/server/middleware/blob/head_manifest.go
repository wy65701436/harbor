package blob

import (
	"fmt"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
	"net/http"
	"time"
)

// HeadManifestMiddleware intercept the head manifest request
func HeadManifestMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		if err := handleHead(req); err != nil {
			lib_http.SendError(rw, err)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

// handleManifestHead ...
func handleManifestHead(req *http.Request) error {
	none := lib.ArtifactInfo{}
	afInfo := lib.GetArtifactInfo(req.Context())
	if afInfo == none {
		return errors.New("cannot get the artifact information from request context").WithCode(errors.NotFoundCode)
	}

	arf, err := artifactController.GetByReference(req.Context(), afInfo.Repository, afInfo.Reference, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to get artifact info, %s, %s", afInfo.ProjectName, afInfo.Repository)
	}
	if arf == nil {
		return errors.New(nil).WithMessage("Artifact info not found, %s, %s", afInfo.ProjectName, afInfo.Repository).WithCode(errors.NotFoundCode)
	}

	bb, err := blob.Ctl.Get(req.Context(), arf.Digest)
	if err != nil {
		return err
	}

	switch bb.Status {
	case blob_models.StatusNone, blob_models.StatusDelete:
		if err := blob.Ctl.Touch(req.Context(), bb); err != nil {
			log.Errorf("failed to update blob: %s status to StatusNone, error:%v", blobInfo.Digest, err)
			return errors.Wrapf(err, fmt.Sprintf("the request id is: %s", req.Header.Get(requestid.HeaderXRequestID)))
		}
	case blob_models.StatusDeleting:
		now := time.Now().UTC()
		// if the deleting exceed 2 hours, marks the blob as StatusDeleteFailed and gives a 404, so client can push it again
		if now.Sub(bb.UpdateTime) > time.Duration(config.GetGCTimeWindow())*time.Hour {
			if err := blob.Ctl.Fail(req.Context(), bb); err != nil {
				log.Errorf("failed to update blob: %s status to StatusDeleteFailed, error:%v", blobInfo.Digest, err)
				return errors.Wrapf(err, fmt.Sprintf("the request id is: %s", req.Header.Get(requestid.HeaderXRequestID)))
			}
		}
		return errors.New(nil).WithMessage(fmt.Sprintf("the asking blob is delete failed, mark it as non existing, request id: %s", req.Header.Get(requestid.HeaderXRequestID))).WithCode(errors.NotFoundCode)
	case blob_models.StatusDeleteFailed:
		return errors.New(nil).WithMessage(fmt.Sprintf("the asking blob is in GC, mark it as non existing, request id: %s", req.Header.Get(requestid.HeaderXRequestID))).WithCode(errors.NotFoundCode)
	default:
		return errors.New(nil).WithMessage(fmt.Sprintf("wrong blob status, %s", bb.Status))
	}
	return nil
}
