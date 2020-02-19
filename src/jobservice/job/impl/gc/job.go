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
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"net/http"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/registryctl/client"
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
	artMgr            artifact.Manager
	artrashMgr        artifactrash.Manager
	repoMgr           repository.Manager
	registryCtlClient client.Client
	logger            logger.Interface
	cfgMgr            *config.CfgManager
	CoreURL           string
	redisURL          string
	deleteUntagged    bool
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
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
func (gc *GarbageCollector) Run(ctx job.Context, params job.Parameters) error {
	if err := gc.init(ctx, params); err != nil {
		return err
	}
	readOnlyCur, err := gc.getReadOnly()
	if err != nil {
		return err
	}
	if readOnlyCur != true {
		if err := gc.setReadOnly(true); err != nil {
			return err
		}
		// defer add the delete all for trash
		defer gc.setReadOnly(readOnlyCur)
	}
	if err := gc.registryCtlClient.Health(); err != nil {
		gc.logger.Errorf("failed to start gc as registry controller is unreachable: %v", err)
		return err
	}
	gc.logger.Infof("start to run gc in job.")
	if err := gc.deleteCandidates(ctx); err != nil {
		gc.logger.Errorf("failed to delete GC candidates in gc job, with error: %v", err)
	}
	gcr, err := gc.registryCtlClient.StartGC()
	if err != nil {
		gc.logger.Errorf("failed to get gc result: %v", err)
		return err
	}
	if err := gc.cleanCache(); err != nil {
		return err
	}
	gc.logger.Infof("GC results: status: %t, message: %s, start: %s, end: %s.", gcr.Status, gcr.Msg, gcr.StartTime, gcr.EndTime)
	gc.logger.Infof("success to run gc in job.")
	return nil
}

func (gc *GarbageCollector) init(ctx job.Context, params job.Parameters) error {
	registryctl.Init()
	gc.registryCtlClient = registryctl.RegistryCtlClient
	gc.logger = ctx.GetLogger()
	gc.artMgr = artifact.NewManager()
	gc.artrashMgr = artifactrash.NewManager()
	gc.repoMgr = repository.New()

	errTpl := "failed to get required property: %s"
	if v, ok := ctx.Get(common.CoreURL); ok && len(v.(string)) > 0 {
		gc.CoreURL = v.(string)
	} else {
		return fmt.Errorf(errTpl, common.CoreURL)
	}
	secret := os.Getenv("JOBSERVICE_SECRET")
	configURL := gc.CoreURL + common.CoreConfigPath
	gc.cfgMgr = config.NewRESTCfgManager(configURL, secret)
	gc.redisURL = params["redis_url_reg"].(string)

	// default is to delete the untagged artifact
	if params["delete_untagged"] == "" {
		gc.deleteUntagged = true
	} else {
		gc.deleteUntagged = params["delete_untagged"].(bool)
	}
	return nil
}

func (gc *GarbageCollector) getReadOnly() (bool, error) {
	if err := gc.cfgMgr.Load(); err != nil {
		return false, err
	}
	return gc.cfgMgr.Get(common.ReadOnly).GetBool(), nil
}

func (gc *GarbageCollector) setReadOnly(switcher bool) error {
	cfg := map[string]interface{}{
		common.ReadOnly: switcher,
	}
	gc.cfgMgr.UpdateConfig(cfg)
	return gc.cfgMgr.Save()
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

func delKeys(con redis.Conn, pattern string) error {
	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(con.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return fmt.Errorf("unexpected type for Int, got type %T", err)
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return fmt.Errorf("converts an array command reply to a []string %v", err)
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		_, err := con.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("failed to clean registry cache %v", err)
		}
	}
	return nil
}

type candidate struct {
	repoName string
	digest   string
}

func (gc *GarbageCollector) deleteCandidates(ctx job.Context) error {
	required, err := gc.artrashMgr.Filter(ctx.SystemContext())
	if err != nil {
		return nil
	}

	_, untagged, err := gc.artMgr.List(ctx.SystemContext(), &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": 1,
			"Tags":         "nil",
		},
	})

	for _, art := range untagged {
		repo, err := gc.repoMgr.Get(ctx.SystemContext(), art.RepositoryID)
		if err != nil {
			return err
		}
		if err := gc.deleteManifest(repo.Name, art.Digest); err != nil {
			gc.logger.Errorf("failded to delete mainfest, %s:%s", repo.Name, art.Digest)
		}
	}

	for _, art := range required {
		if err := gc.deleteManifest(art.RepositoryName, art.Digest); err != nil {
			gc.logger.Errorf("failded to delete mainfest, %s:%s", art.RepositoryName, art.Digest)
		}
	}

	return nil
}

func (gc *GarbageCollector) deleteManifest(repository string, digest string) error {
	repoClient, err := gc.newRepositoryClient(repository)
	if err != nil {
		return err
	}
	if err := repoClient.DeleteManifest(digest); err != nil {
		return err
	}
	return nil
}

func (gc *GarbageCollector) newRepositoryClient(repository string) (*registry.Repository, error) {
	uam := &auth.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	authorizer := auth.DefaultBasicAuthorizer()
	transport := registry.NewTransport(http.DefaultTransport, authorizer, uam)
	client := &http.Client{
		Transport: transport,
	}
	gc.cfgMgr.Get(common.RegistryURL)
	endpoint := gc.cfgMgr.Get(common.RegistryURL)
	return registry.NewRepository(repository, endpoint.GetString(), client)
}
