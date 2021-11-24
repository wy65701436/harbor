package cosign

import (
	"fmt"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io/ioutil"
	"net/http"
	"regexp"
)

var (
	cosignReg = regexp.MustCompile(fmt.Sprintf(`sha256-%s.sig$`, reference.IdentifierRegexp))
)

// CosignSignatureMiddleware middleware to record the linkeage of artifact and its accessory
func CosignSignatureMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		// Only record when patch blob upload success
		if statusCode != http.StatusAccepted {
			return nil
		}

		ctx := r.Context()
		logger := log.G(ctx).WithFields(log.Fields{"middleware": "cosign"})

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}

		// Needs tag to match the cosign tag pattern.
		if info.Tag == "" {
			return nil
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		manifest, descriptor, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return err
		}

		for _, descriptor := range manifest.References() {
			digest := descriptor.Digest.String()
			switch descriptor.MediaType {
			// skip foreign layer
			case schema2.MediaTypeForeignLayer:
				continue
			// manifest or index
			case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList,
				v1.MediaTypeImageManifest, schema2.MediaTypeManifest,
				schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest:
				if err = c.Copy(srcRepo, digest, dstRepo, digest, false); err != nil {
					return err
				}
			// common layer
			default:
				exist, err := c.BlobExist(dstRepo, digest)
				if err != nil {
					return err
				}
				// the layer already exist, skip
				if exist {
					continue
				}
				// when the copy happens inside the same registry, use mount
				if err = c.MountBlob(srcRepo, digest, dstRepo); err != nil {
					return err
				}
			}
		}

		mediaType, payload, err := manifest.Payload()
		if err != nil {
			return err
		}

		return nil
	})
}
