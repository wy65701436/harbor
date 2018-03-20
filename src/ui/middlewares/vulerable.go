package middlewares

import (
	"fmt"
	"net/http"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

type vulnerableHandler struct {
	next http.Handler
}

func (vh vulnerableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(imageInfoCtxKey)
	if imgRaw == nil || !config.WithClair() {
		vh.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(imageInfoCtxKey).(ImageInfo)
	if img.digest == "" {
		vh.next.ServeHTTP(rw, req)
		return
	}
	projectVulnerableEnabled, projectVulnerableSeverity := GetPolicyChecker().vulnerablePolicy(img.projectName)
	if !projectVulnerableEnabled {
		vh.next.ServeHTTP(rw, req)
		return
	}
	overview, err := dao.GetImgScanOverview(img.digest)
	if err != nil {
		log.Errorf("failed to get ImgScanOverview with repo: %s, reference: %s, digest: %s. Error: %v", img.repository, img.reference, img.digest, err)
		http.Error(rw, marshalError("Failed to get ImgScanOverview."), http.StatusPreconditionFailed)
		return
	}
	// severity is 0 means that the image fails to scan or not scanned successfully.
	if overview == nil || overview.Sev == 0 {
		log.Debugf("cannot get the image scan overview info, failing the response.")
		http.Error(rw, marshalError("Cannot get the image severity."), http.StatusPreconditionFailed)
		return
	}
	imageSev := overview.Sev
	if imageSev >= int(projectVulnerableSeverity) {
		log.Debugf("the image severity: %q is higher then project setting: %q, failing the response.", models.Severity(imageSev), projectVulnerableSeverity)
		http.Error(rw, marshalError(fmt.Sprintf("The severity of vulnerability of the image: %q is equal or higher than the threshold in project setting: %q.", models.Severity(imageSev), projectVulnerableSeverity)), http.StatusPreconditionFailed)
		return
	}
	vh.next.ServeHTTP(rw, req)
}
