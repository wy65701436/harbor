package immutable

import (
	"errors"
	"fmt"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	middlerware_err "github.com/goharbor/harbor/src/core/middlewares/util/error"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/tag"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// ManifestPush ...
func ManifestPush() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if err := handlePush(req); err != nil {
				var e *middlerware_err.ErrImmutable
				if errors.As(err, &e) {
					pkgE := internal_errors.New(e).WithCode(internal_errors.PreconditionCode)
					msg := internal_errors.NewErrs(pkgE).Error()
					http.Error(rw, msg, http.StatusPreconditionFailed)
					return
				}
				pkgE := internal_errors.New(fmt.Errorf("error occurred when to handle request in immutable handler: %v", err)).WithCode(internal_errors.GeneralCode)
				msg := internal_errors.NewErrs(pkgE).Error()
				log.Info("==================")
				log.Info(msg)
				http.Error(rw, msg, http.StatusInternalServerError)
			}
			next.ServeHTTP(rw, req)
		})
	}
}

// handlePush ...
func handlePush(req *http.Request) error {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return errors.New("cannot get the manifest information from request context")
	}

	_, repoName := common_util.ParseRepository(mf.Repository)
	var matched bool
	matched, err := rule.NewRuleMatcher(mf.ProjectID).Match(art.Candidate{
		Repository:  repoName,
		Tag:         mf.Tag,
		NamespaceID: mf.ProjectID,
	})
	if err != nil {
		return err
	}
	if !matched {
		return nil
	}

	// match repository ...
	total, repos, err := repository.Mgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"Name": repoName,
		},
	})
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	// match artifacts ...
	total, afs, err := artifact.Mgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID":    mf.ProjectID,
			"RepositoryID": repos[0].RepositoryID,
		},
	})
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	// match tags ...
	total, _, err = tag.Mgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"ArtifactID":   afs[0].ID,
			"RepositoryID": repos[0].RepositoryID,
		},
	})
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	return NewErrImmutable(repoName, mf.Tag)
}
