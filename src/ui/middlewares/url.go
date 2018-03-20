package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"
	uiutils "github.com/vmware/harbor/src/ui/utils"
)

const (
	RegistryProxyPrefix = "/registryproxy"
	manifestURLPattern  = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`
)

type UrlHandler struct {
	next http.Handler
}

func (uh UrlHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("in url handler, path: %s", req.URL.Path)
	req.URL.Path = strings.TrimPrefix(req.URL.Path, RegistryProxyPrefix)
	flag, repository, reference := matchPullManifest(req)
	if flag {
		components := strings.SplitN(repository, "/", 2)
		if len(components) < 2 {
			http.Error(rw, marshalError(fmt.Sprintf("Bad repository name: %s", repository)), http.StatusBadRequest)
			return
		}

		client, err := uiutils.NewRepositoryClientForUI(tokenUsername, repository)
		if err != nil {
			log.Errorf("Error creating repository Client: %v", err)
			http.Error(rw, marshalError(fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}
		digest, _, err := client.ManifestExist(reference)
		if err != nil {
			log.Errorf("Failed to get digest for reference: %s, error: %v", reference, err)
			http.Error(rw, marshalError(fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}

		img := ImageInfo{
			repository:  repository,
			reference:   reference,
			projectName: components[0],
			digest:      digest,
		}

		log.Debugf("image info of the request: %#v", img)
		ctx := context.WithValue(req.Context(), imageInfoCtxKey, img)
		req = req.WithContext(ctx)
	}
	uh.next.ServeHTTP(rw, req)
}

// matchPullManifest checks if the request looks like a request to pull manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func matchPullManifest(req *http.Request) (bool, string, string) {
	//TODO: add user agent check.
	if req.Method != http.MethodGet {
		return false, "", ""
	}
	re := regexp.MustCompile(manifestURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 3 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1], s[2]
	}
	return false, "", ""
}
