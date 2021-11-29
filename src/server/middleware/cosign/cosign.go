package cosign

import (
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	"io/ioutil"
	"net/http"
	"regexp"
)

var (
	cosignDigest = "digest"
	// repositorySubexp is the name for sub regex that maps to repository name in the url
	repositorySubexp     = "repository"
	cosignRe             = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/manifests/sha256-(?P<%s>%s).sig$`, repositorySubexp, reference.NameRegexp.String(), cosignDigest, reference.IdentifierRegexp))
	mediaTypeCosignLayer = "application/vnd.dev.cosign.simplesigning.v1+json"
)

// CosignSignatureMiddleware middleware to record the linkeage of artifact and its accessory
/* PUT /v2/library/hello-world/manifests/sha256-1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792.sig
{
	"schemaVersion":2,
	"config":{
		"mediaType":"application/vnd.oci.image.config.v1+json",
		"size":233,
		"digest":"sha256:d4e6059ece7bea95266fd7766353130d4bf3dc21048b8a9783c98b8412618c38"
	},
	"layers":[
		{
			"mediaType":"application/vnd.dev.cosign.simplesigning.v1+json",
			"size":250,
			"digest":"sha256:91a821a0e2412f1b99b07bfe176451bcc343568b761388718abbf38076048564",
			"annotations":{
				"dev.cosignproject.cosign/signature":"MEUCIQD/imXjZJlcV82eXu9y9FJGgbDwVPw7AaGFzqva8G+CgwIgYc4CRvEjwoAwkzGoX+aZxQWCASpv5G+EAWDKOJRLbTQ="
			}
		}
	]
}
*/
func CosignSignatureMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusAccepted {
			return nil
		}

		ctx := r.Context()
		logger := log.G(ctx).WithFields(log.Fields{"middleware": "cosign"})
		logger.Info("0000000000000000")

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}
		if info.Tag == "" {
			return nil
		}

		// Needs tag to match the cosign tag pattern.
		_, subjectArtDigest, ok := matchCosignSignaturePattern(r.URL.Path)
		if !ok {
			logger.Info("11111111111")
			return nil
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		manifest, desc, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return err
		}

		var hasSignature bool
		for _, descriptor := range manifest.References() {
			if descriptor.MediaType == mediaTypeCosignLayer {
				hasSignature = true
				break
			}
		}

		if hasSignature {
			subjectArt, err := artifact.Ctl.GetByReference(ctx, info.Repository, fmt.Sprintf("sha256:%s", subjectArtDigest), nil)
			if err != nil {
				return err
			}
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, desc.Digest.String(), nil)
			if err != nil {
				return err
			}

			_, err = accessory.Mgr.Create(ctx, &cosign.Signature{base.Default{
				Opts: &model.Options{
					ArtifactID:        art.ID,
					SubjectArtifactID: subjectArt.ID,
					Size:              desc.Size,
					Digest:            desc.Digest.String(),
				},
			}})
			if err != nil {
				return err
			}
		} else {
			logger.Info("22222222")
		}

		return nil
	})
}

// matchCosignSignaturePattern checks whether the provided path matches the blob upload URL pattern,
// if does, returns the repository as well
func matchCosignSignaturePattern(path string) (repository, digest string, match bool) {
	strs := cosignRe.FindStringSubmatch(path)
	if len(strs) < 3 {
		return "", "", false
	}
	return strs[1], strs[2], true
}
