package middleware

import (
	"context"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/tag"
	"sync"
)

type contextKey string

const (
	// manifestInfoKey the context key for manifest info
	manifestInfoKey = contextKey("ManifestInfo")
	// ScannerPullCtxKey the context key for robot account to bypass the pull policy check.
	ScannerPullCtxKey = contextKey("ScannerPullCheck")
)

// ManifestInfo ...
type ManifestInfo struct {
	ProjectID  int64
	Repository string
	Tag        string
	Digest     string

	manifestExist     bool
	manifestExistErr  error
	manifestExistOnce sync.Once
}

func (info *ManifestInfo) ManifestExists(ctx context.Context) (bool, error) {
	info.manifestExistOnce.Do(func() {
		total, repos, err := repository.Mgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"Name": info.Repository,
			},
		})
		if err != nil {
			info.manifestExistErr = err
			return
		}
		if total == 0 {
			return
		}

		total, tags, err := tag.Mgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"Name":         info.Tag,
				"RepositoryID": repos[0].RepositoryID,
			},
		})
		if err != nil {
			info.manifestExistErr = err
			return
		}
		if total == 0 {
			return
		}

		info.manifestExist = total > 0
		info.manifestExistErr = err
	})

	return info.manifestExist, info.manifestExistErr
}

// NewManifestInfoContext returns context with manifest info
func NewManifestInfoContext(ctx context.Context, info *ManifestInfo) context.Context {
	return context.WithValue(ctx, manifestInfoKey, info)
}

// ManifestInfoFromContext returns manifest info from context
func ManifestInfoFromContext(ctx context.Context) (*ManifestInfo, bool) {
	info, ok := ctx.Value(manifestInfoKey).(*ManifestInfo)
	return info, ok
}

// NewScannerPullContext returns context with policy check info
func NewScannerPullContext(ctx context.Context, scannerPull bool) context.Context {
	return context.WithValue(ctx, ScannerPullCtxKey, scannerPull)
}
