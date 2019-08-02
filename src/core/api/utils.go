// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"net/http"
	"sort"
	"strings"

	"github.com/goharbor/harbor/src/common/dao"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/core/service/token"
	coreutils "github.com/goharbor/harbor/src/core/utils"

	"strconv"
	"sync"
)

// SyncRegistry syncs the repositories of registry with database.
func SyncRegistry(pm promgr.ProjectManager) error {

	log.Infof("Start syncing repositories from registry to DB... ")

	reposInRegistry, err := catalog()
	if err != nil {
		log.Error(err)
		return err
	}

	var repoRecordsInDB []*models.RepoRecord
	repoRecordsInDB, err = dao.GetRepositories()
	if err != nil {
		log.Errorf("error occurred while getting all registories. %v", err)
		return err
	}

	var reposInDB []string
	for _, repoRecordInDB := range repoRecordsInDB {
		reposInDB = append(reposInDB, repoRecordInDB.Name)
	}

	var reposToAdd []string
	var reposToDel []string
	reposToAdd, reposToDel, err = diffRepos(reposInRegistry, reposInDB, pm)
	if err != nil {
		return err
	}

	if len(reposToAdd) > 0 {
		log.Debugf("Start adding repositories into DB... ")
		for _, repoToAdd := range reposToAdd {
			project, _ := utils.ParseRepository(repoToAdd)
			pullCount, err := dao.CountPull(repoToAdd)
			if err != nil {
				log.Errorf("Error happens when counting pull count from access log: %v", err)
			}
			pro, err := pm.Get(project)
			if err != nil {
				log.Errorf("failed to get project %s: %v", project, err)
				continue
			}
			repoRecord := models.RepoRecord{
				Name:      repoToAdd,
				ProjectID: pro.ProjectID,
				PullCount: pullCount,
			}

			if err := dao.AddRepository(repoRecord); err != nil {
				log.Errorf("Error happens when adding the missing repository: %v", err)
			} else {
				log.Debugf("Add repository: %s success.", repoToAdd)
			}
		}
	}

	if len(reposToDel) > 0 {
		log.Debugf("Start deleting repositories from DB... ")
		for _, repoToDel := range reposToDel {
			if err := dao.DeleteRepository(repoToDel); err != nil {
				log.Errorf("Error happens when deleting the repository: %v", err)
			} else {
				log.Debugf("Delete repository: %s success.", repoToDel)
			}
		}
	}

	log.Infof("Sync repositories from registry to DB is done.")
	return nil
}

// DumpRegistry dumps the registry data, and compute quota for each project, then to update harbor DB(quota_usage).
func DumpRegistry() error {
	log.Infof("Start fixing project quota... ")

	reposInRegistry, err := catalog()
	if err != nil {
		log.Error(err)
		return err
	}

	// repoMap : map[project_name : []repo list]
	repoMap := make(map[string][]string)
	for _, item := range reposInRegistry {
		projectName := strings.Split(item, "/")[0]
		_, exist := repoMap[projectName]
		if !exist {
			repoMap[projectName] = []string{item}
		} else {
			repos := repoMap[projectName]
			repos = append(repos, item)
			repoMap[projectName] = repos
		}
	}

	log.Info(" ^^^^^^^^^^^^^^^^^^ ")
	log.Info(repoMap)
	log.Info(" ^^^^^^^^^^^^^^^^^^ ")

	for project, repos := range repoMap {
		go func(project string) {
			err := getProjectUsage(project, repos)
			if err != nil {
				log.Warningf("Error happens when to get quota for project: %s, with error: %v", project, err)
			}
		}(project)
	}

	log.Infof("End fixing project quota... ")

	return nil
}

// fixQuotaUsage fixes the quota usage in the data base.
func fixQuotaUsage(project string, usage quota.ResourceList) error {
	infinite := quota.ResourceList{
		quota.ResourceCount:   -1,
		quota.ResourceStorage: -1,
	}
	projectID, err := GetProjectID(project)
	if err != nil {
		log.Warningf("Error occurred when to get project ID %v", err)
		return err
	}
	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(projectID, 10))
	if err != nil {
		log.Errorf("Error occurred when to new quota manager %v", err)
		return err
	}
	if err := quotaMgr.EnsureQuota(infinite, usage); err != nil {
		log.Errorf("cannot get quota for the project resource: %d, err: %v", projectID, err)
		return err
	}
	return nil
}

func getProjectUsage(project string, repoList []string) error {
	var wg sync.WaitGroup

	for _, repo := range repoList {
		wg.Add(1)

		go func() {
			projectQuotaCount := int64(0)
			projectQuotaSize := int64(0)
			blobMap := make(map[string]int64)
			repoClient, err := coreutils.NewRepositoryClientForUI("harbor-core", repo)
			if err != nil {
				log.Errorf("Failed to create repo client.")
			}
			tags, err := repoClient.ListTag()
			if err != nil {
				log.Errorf("Failed to list tags for repo: %s", repo)
			}
			for _, tag := range tags {
				projectQuotaCount++
				_, mediaType, payload, err := repoClient.PullManifest(tag, []string{
					schema1.MediaTypeManifest,
					schema1.MediaTypeSignedManifest,
					schema2.MediaTypeManifest,
				})
				if err != nil {
					log.Error(err)
				}
				manifest, desc, err := registry.UnMarshal(mediaType, payload)
				if err != nil {
					log.Error(err)
				}
				projectQuotaSize = projectQuotaSize + desc.Size
				for _, layer := range manifest.References() {
					_, exist := blobMap[layer.Digest.String()]
					if !exist {
						blobMap[layer.Digest.String()] = layer.Size
						projectQuotaSize = projectQuotaSize + layer.Size
					}
				}
			}

			usage := quota.ResourceList{
				quota.ResourceCount:   projectQuotaCount,
				quota.ResourceStorage: projectQuotaSize,
			}

			log.Info(" ================= ")
			log.Info(project)
			log.Info(projectQuotaCount)
			log.Info(projectQuotaSize)
			log.Info(" ================= ")

			if err := fixQuotaUsage(project, usage); err != nil {
				log.Error(err)
			}

			wg.Done()
		}()

		wg.Wait()
	}

	return nil

}

func catalog() ([]string, error) {
	repositories := []string{}

	rc, err := initRegistryClient()
	if err != nil {
		return repositories, err
	}

	repositories, err = rc.Catalog()
	if err != nil {
		return repositories, err
	}

	return repositories, nil
}

func diffRepos(reposInRegistry []string, reposInDB []string,
	pm promgr.ProjectManager) ([]string, []string, error) {
	var needsAdd []string
	var needsDel []string

	sort.Strings(reposInRegistry)
	sort.Strings(reposInDB)

	i, j := 0, 0
	repoInR, repoInD := "", ""
	for i < len(reposInRegistry) && j < len(reposInDB) {
		repoInR = reposInRegistry[i]
		repoInD = reposInDB[j]
		d := strings.Compare(repoInR, repoInD)
		if d < 0 {
			i++
			exist, err := projectExists(pm, repoInR)
			if err != nil {
				log.Errorf("failed to check the existence of project %s: %v", repoInR, err)
				continue
			}

			if !exist {
				continue
			}

			// TODO remove the workaround when the bug of registry is fixed
			client, err := coreutils.NewRepositoryClientForUI("harbor-core", repoInR)
			if err != nil {
				return needsAdd, needsDel, err
			}

			exist, err = repositoryExist(repoInR, client)
			if err != nil {
				return needsAdd, needsDel, err
			}

			if !exist {
				continue
			}

			needsAdd = append(needsAdd, repoInR)
		} else if d > 0 {
			needsDel = append(needsDel, repoInD)
			j++
		} else {
			// TODO remove the workaround when the bug of registry is fixed
			client, err := coreutils.NewRepositoryClientForUI("harbor-core", repoInR)
			if err != nil {
				return needsAdd, needsDel, err
			}

			exist, err := repositoryExist(repoInR, client)
			if err != nil {
				return needsAdd, needsDel, err
			}

			if !exist {
				needsDel = append(needsDel, repoInD)
			}

			i++
			j++
		}
	}

	for i < len(reposInRegistry) {
		repoInR = reposInRegistry[i]
		i++
		exist, err := projectExists(pm, repoInR)
		if err != nil {
			log.Errorf("failed to check whether project of %s exists: %v", repoInR, err)
			continue
		}

		if !exist {
			continue
		}

		client, err := coreutils.NewRepositoryClientForUI("harbor-core", repoInR)
		if err != nil {
			log.Errorf("failed to create repository client: %v", err)
			continue
		}

		exist, err = repositoryExist(repoInR, client)
		if err != nil {
			log.Errorf("failed to check the existence of repository %s: %v", repoInR, err)
			continue
		}

		if !exist {
			continue
		}

		needsAdd = append(needsAdd, repoInR)
	}

	for j < len(reposInDB) {
		needsDel = append(needsDel, reposInDB[j])
		j++
	}

	return needsAdd, needsDel, nil
}

func projectExists(pm promgr.ProjectManager, repository string) (bool, error) {
	project, _ := utils.ParseRepository(repository)
	return pm.Exists(project)
}

func initRegistryClient() (r *registry.Registry, err error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	addr := endpoint
	if strings.Contains(endpoint, "://") {
		addr = strings.Split(endpoint, "://")[1]
	}

	if err := utils.TestTCPConn(addr, 60, 2); err != nil {
		return nil, err
	}

	authorizer := auth.NewRawTokenAuthorizer("harbor-core", token.Registry)
	return registry.NewRegistry(endpoint, &http.Client{
		Transport: registry.NewTransport(registry.GetHTTPTransport(), authorizer),
	})
}

func buildReplicationURL() string {
	url := config.InternalJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication", url)
}

func buildJobLogURL(jobID string, jobType string) string {
	url := config.InternalJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/%s/%s/log", url, jobType, jobID)
}

func buildReplicationActionURL() string {
	url := config.InternalJobServiceURL()
	return fmt.Sprintf("%s/api/jobs/replication/actions", url)
}

func repositoryExist(name string, client *registry.Repository) (bool, error) {
	tags, err := client.ListTag()
	if err != nil {
		if regErr, ok := err.(*commonhttp.Error); ok && regErr.Code == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return len(tags) != 0, nil
}

// GetProjectID ...
func GetProjectID(name string) (int64, error) {
	project, err := dao.GetProjectByName(name)
	if err != nil {
		return 0, err
	}
	if project != nil {
		return project.ProjectID, nil
	}
	return 0, fmt.Errorf("project %s is not found", name)
}
