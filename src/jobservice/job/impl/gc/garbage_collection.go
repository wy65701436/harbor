// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gc

import (
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
	"os"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/blob"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/registryctl/client"
)

var (
	regCtlInit = registryctl.Init
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	blobPrefix            = "blobs::*"
	repoPrefix            = "repository::*"
	uploadSizePattern     = "upload:*:size"
)

// GarbageCollector is the struct to run registry's garbage collection
type GarbageCollector struct {
	artCtl            artifact.Controller
	artrashMgr        artifactrash.Manager
	blobMgr           blob.Manager
	projectCtl        project.Controller
	registryCtlClient client.Client
	logger            logger.Interface
	redisURL          string
	deleteUntagged    bool
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (gc *GarbageCollector) MaxCurrency() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (gc *GarbageCollector) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (gc *GarbageCollector) Validate(params job.Parameters) error {
	return nil
}

// Run implements the interface in job/Interface
// The workflow of GC is:
// 1, set harbor to readonly
// 2, select the candidate artifacts from Harbor DB.
// 3, call registry API(--delete-untagged=false) to delete manifest bases on the results of #2
// 4, clean keys of redis DB of registry, clean artifact trash and untagged from DB.
// 5, roll back readonly.
// More details:
// 1, why disable delete untagged when to call registry API
//		Generally because that we introduce Harbor tag in v2.0, it's in database but no corresponding data in registry.
//		Also one failure case example:
// 			there are two parts for putting an manifest in Harbor: write database and write storage, but they're not in a transaction,
//			which leads to the data mismatching in parallel pushing images with same tag but different digest. The valid artifact in
//			harbor DB could be a untagged one in the storage. If we enable the delete untagged, the valid data could be removed from the storage.
// 2, what to be cleaned
//		> the deleted artifact, bases on table of artifact_trash and artifact
//		> the untagged artifact(optional), bases on table of artifact.
func (gc *GarbageCollector) Run(ctx job.Context, params job.Parameters) error {
	if err := gc.init(ctx, params); err != nil {
		return err
	}

	gc.logger.Infof("start to run gc in job.")
	removedArtifacts, err := gc.getDeletedArtifact(ctx)
	if err != nil {
		gc.logger.Errorf("failed to delete GC candidates in gc job, with error: %v", err)
	}
	// no need to execute GC as there is no removed artifacts.
	// Do this is to handle if user trigger GC job several times, only one job should do the following logic.
	if len(removedArtifacts) == 0 {
		return nil
	}

	//  get gc candidates
	blobs, err := gc.blobMgr.UselessBlobs(ctx.SystemContext(), 0)
	if err != nil {
		gc.logger.Errorf("failed to get gc candidate: %v", err)
		return err
	}
	setRepositories(removedArtifacts, blobs)

	// mark delete status
	blobCt := 0
	mfCt := 0
	for _, blob := range blobs {
		blob.Status = blob_models.StatusDelete
		gc.logger.Infof("blob eligible for deletion: %s", blob.Digest)
		_, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
		if err != nil {
			gc.logger.Errorf("failed to mark gc candidate: %s", blob.Digest)
			continue
		}
		if blob.IsManifest() {
			mfCt++
		} else {
			blobCt++
		}
	}
	gc.logger.Infof("%d blobs and %d manifests eligible for deletion", blobCt, mfCt)

	for _, blob := range blobs {
		if blob.IsManifest() {
			for _, repo := range blob.Repositories {
				go func() {
					// Harbor cannot know the existing tags in the backend from its database, so let the v2 DELETE manifest to remove all of them.
					if err := v2DeleteManifest(repo, blob.Digest); err != nil {
						gc.logger.Infof("failed to delete manifest, %s:%s with error: %v", repo, blob.Digest, err)
						return
					}
					gc.logger.Infof("delete the manifest with registry v2 API: RepositoryName(%s)-MediaType:(%s)-Digest:(%s)",
						repo, blob.ContentType, blob.Digest)
					// for manifest, it has to delete the revisions folder of each repository
					if err := gc.registryCtlClient.DeleteManifest(repo, blob.Digest); err != nil {
						gc.logger.Infof("failed to remove manifest revision of repository %s, %v")
					}
				}()
			}
		}

		// delete all of blobs, which include config, layer and manifest
		blob.Status = blob_models.StatusDeleting
		_, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
		if err != nil {
			gc.logger.Errorf("failed to mark gc candidate deleting: %s, %s", blob.Digest, blob.Status)
			continue
		}
		if err := gc.registryCtlClient.DeleteBlob(blob.Digest); err != nil {
			blob.Status = blob_models.StatusDeleteFailed
			_, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
			if err != nil {
				gc.logger.Errorf("failed to mark gc candidate deletefailed: %s, %s", blob.Digest, blob.Status)
				continue
			}
		}
		if err := gc.blobMgr.Delete(ctx.SystemContext(), blob.ID); err != nil {
			// tries to make it as StatusDeleteFailed
			blob.Status = blob_models.StatusDeleteFailed
			_, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
			if err != nil {
				gc.logger.Errorf("failed to mark gc candidate deletefailed: %s, %s", blob.Digest, blob.Status)
				continue
			}
		}
	}

	gc.removeUntaggedBlobs(ctx)
	if err := gc.cleanCache(); err != nil {
		return err
	}
	//gc.logger.Infof("GC results: status: %t, message: %s, start: %s, end: %s.", gcr.Status, gcr.Msg, gcr.StartTime, gcr.EndTime)
	gc.logger.Infof("success to run gc in job.")
	return nil
}

// as table blob has no repository name, here needs to use the ArtifactTrash to fill it in.
func setRepositories(ats []model.ArtifactTrash, blobs []*blob_models.Blob) {
	for _, at := range ats {
		for _, blob := range blobs {
			if at.Digest == blob.Digest {
				blob.Repositories = append(blob.Repositories, at.RepositoryName)
			}
		}
	}
}

func (gc *GarbageCollector) init(ctx job.Context, params job.Parameters) error {
	regCtlInit()
	gc.logger = ctx.GetLogger()
	opCmd, flag := ctx.OPCommand()
	if flag && opCmd.IsStop() {
		gc.logger.Info("received the stop signal, quit GC job.")
		return nil
	}
	// UT will use the mock client, ctl and mgr
	if os.Getenv("UTTEST") != "true" {
		gc.registryCtlClient = registryctl.RegistryCtlClient
		gc.artCtl = artifact.Ctl
		gc.artrashMgr = artifactrash.NewManager()
		gc.blobMgr = blob.NewManager()
		gc.projectCtl = project.Ctl
	}
	if err := gc.registryCtlClient.Health(); err != nil {
		gc.logger.Errorf("failed to start gc as registry controller is unreachable: %v", err)
		return err
	}
	gc.redisURL = params["redis_url_reg"].(string)
	// default is to delete the untagged artifact
	gc.deleteUntagged = true
	deleteUntagged, exist := params["delete_untagged"]
	if exist {
		if untagged, ok := deleteUntagged.(bool); ok && !untagged {
			gc.deleteUntagged = untagged
		}
	}
	return nil
}

// cleanCache is to clean the registry cache for GC.
// To do this is because the issue https://github.com/docker/distribution/issues/2094
func (gc *GarbageCollector) cleanCache() error {
	con, err := redis.DialURL(
		gc.redisURL,
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)

	if err != nil {
		gc.logger.Errorf("failed to connect to redis %v", err)
		return err
	}
	defer con.Close()

	// clean all keys in registry redis DB.

	// sample of keys in registry redis:
	// 1) "blobs::sha256:1a6fd470b9ce10849be79e99529a88371dff60c60aab424c077007f6979b4812"
	// 2) "repository::library/hello-world::blobs::sha256:4ab4c602aa5eed5528a6620ff18a1dc4faef0e1ab3a5eddeddb410714478c67f"
	// 3) "upload:fbd2e0a3-262d-40bb-abe4-2f43aa6f9cda:size"
	patterns := []string{blobPrefix, repoPrefix, uploadSizePattern}
	for _, pattern := range patterns {
		if err := delKeys(con, pattern); err != nil {
			gc.logger.Errorf("failed to clean registry cache %v, pattern %s", err, pattern)
			return err
		}
	}

	return nil
}

// getDeletedArtifact deletes the two parts of artifact from harbor DB
// 1, required part, the artifacts were removed from Harbor.
// 2, optional part, the untagged artifacts.
func (gc *GarbageCollector) getDeletedArtifact(ctx job.Context) ([]model.ArtifactTrash, error) {
	if os.Getenv("UTTEST") == "true" {
		gc.logger = ctx.GetLogger()
	}
	// default is not to clean trash
	flushTrash := false
	defer func() {
		if flushTrash {
			gc.logger.Info("flush artifact trash")
			if err := gc.artrashMgr.Flush(ctx.SystemContext()); err != nil {
				gc.logger.Errorf("failed to flush artifact trash: %v", err)
			}
		}
	}()
	var arts []model.ArtifactTrash

	// handle the optional ones, and the artifact controller will move them into trash.
	if gc.deleteUntagged {
		untagged, err := gc.artCtl.List(ctx.SystemContext(), &q.Query{
			Keywords: map[string]interface{}{
				"Tags": "nil",
			},
		}, nil)
		if err != nil {
			return arts, err
		}
		gc.logger.Info("start to delete untagged artifact.")
		for _, art := range untagged {
			if err := gc.artCtl.Delete(ctx.SystemContext(), art.ID); err != nil {
				// the failure ones can be GCed by the next execution
				gc.logger.Errorf("failed to delete untagged:%d artifact in DB, error, %v", art.ID, err)
				continue
			}
			gc.logger.Infof("delete the untagged artifact: ProjectID:(%d)-RepositoryName(%s)-MediaType:(%s)-Digest:(%s)",
				art.ProjectID, art.RepositoryName, art.ManifestMediaType, art.Digest)
		}
		gc.logger.Info("end to delete untagged artifact.")
	}

	// handle the trash
	arts, err := gc.artrashMgr.Filter(ctx.SystemContext())
	if err != nil {
		return arts, err
	}
	gc.logger.Info("required candidate: %+v", arts)
	flushTrash = true
	return arts, nil
	//for _, art := range required {
	//	if err := v2DeleteManifest(art.RepositoryName, art.Digest); err != nil {
	//		return fmt.Errorf("failed to delete manifest, %s:%s with error: %v", art.RepositoryName, art.Digest, err)
	//	}
	//	gc.logger.Infof("delete the manifest with registry v2 API: RepositoryName(%s)-MediaType:(%s)-Digest:(%s)",
	//		art.RepositoryName, art.ManifestMediaType, art.Digest)
	//}
	//gc.logger.Info("end to delete required artifact.")
	//flushTrash = true
	//return nil
}

// clean the untagged blobs in each project, these blobs are not referenced by any manifest and will be cleaned by GC
func (gc *GarbageCollector) removeUntaggedBlobs(ctx job.Context) {
	// get all projects
	projects := func(chunkSize int) <-chan *models.Project {
		ch := make(chan *models.Project, chunkSize)

		go func() {
			defer close(ch)

			params := &models.ProjectQueryParam{
				Pagination: &models.Pagination{Page: 1, Size: int64(chunkSize)},
			}

			for {
				results, err := gc.projectCtl.List(ctx.SystemContext(), params, project.Metadata(false))
				if err != nil {
					gc.logger.Errorf("list projects failed, error: %v", err)
					return
				}

				for _, p := range results {
					ch <- p
				}

				if len(results) < chunkSize {
					break
				}

				params.Pagination.Page++
			}

		}()

		return ch
	}(50)

	for project := range projects {
		all, err := gc.blobMgr.List(ctx.SystemContext(), blob.ListParams{
			ProjectID: project.ProjectID,
		})
		if err != nil {
			gc.logger.Errorf("failed to get blobs of project, %v", err)
			continue
		}
		if err := gc.blobMgr.CleanupAssociationsForProject(ctx.SystemContext(), project.ProjectID, all); err != nil {
			gc.logger.Errorf("failed to clean untagged blobs of project, %v", err)
			continue
		}
	}
}
