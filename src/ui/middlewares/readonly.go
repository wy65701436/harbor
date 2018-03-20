package middlewares

import (
	"net/http"
	"regexp"

	"github.com/vmware/harbor/src/ui/config"
)

//POST /v2/<name>/blobs/uploads/
const pushURLPattern = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)blobs/uploads/`

type ReadonlyHandler struct {
	next http.Handler
}

func (rh ReadonlyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if config.ReadOnly() {
		dockerPushFlag := matchDockerPush(req)
		if dockerPushFlag {
			http.Error(rw, "Docker Push is not allowed in read only mode.", http.StatusServiceUnavailable)
			return
		}
	}
	rh.next.ServeHTTP(rw, req)
}

// matchDockerPush checks if the request looks like a request to push manifest.
func matchDockerPush(req *http.Request) bool {
	if req.Method != http.MethodPost {
		return false
	}
	re := regexp.MustCompile(pushURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	return len(s) == 2
}
