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

package registry

import (
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/core/api"
	quota "github.com/goharbor/harbor/src/core/api/quota"
	"github.com/goharbor/harbor/src/core/promgr"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"strings"
	"sync"
	"time"
)

// RegistryMigrator ...
type RegistryMigrator struct {
	pm promgr.ProjectManager
}

// NewRegistryMigrator returns a new RegistryMigrator.
func NewRegistryMigrator(pm promgr.ProjectManager) quota.QuotaMigrator {
	migrator := RegistryMigrator{
		pm: pm,
	}
	return &migrator
}

// Load ...
func (rm *RegistryMigrator) Dump() ([]quota.ProjectInfo, error) {
	reposInRegistry, err := api.Catalog()
	if err != nil {
		return nil, err
	}

	// repoMap : map[project_name : []repo list]
	repoMap := make(map[string][]string)
	for _, item := range reposInRegistry {
		projectName := strings.Split(item, "/")[0]
		pro, err := rm.pm.Get(projectName)
		if err != nil {
			log.Errorf("failed to get project %s: %v", projectName, err)
			continue
		}
		_, exist := repoMap[pro.Name]
		if !exist {
			repoMap[pro.Name] = []string{item}
		} else {
			repos := repoMap[pro.Name]
			repos = append(repos, item)
			repoMap[pro.Name] = repos
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(repoMap))
	infoChan := make(chan interface{})
	for project, repos := range repoMap {
		go func(project string, repos []string) {
			defer wg.Done()
			info := infoOfProject(project, repos)
			infoChan <- info
		}(project, repos)
	}

	go func() {
		wg.Wait()
		close(infoChan)
	}()

	var projects []quota.ProjectInfo
	for item := range infoChan {
		projects = append(projects, item.(quota.ProjectInfo))
	}
	return projects, nil
}

// Usage
// registry needs to merge the shard blobs of different repositories.
func (rm *RegistryMigrator) Usage(projects []quota.ProjectInfo) ([]quota.ProjectUsage, error) {
	var pros []quota.ProjectUsage

	for _, project := range projects {
		var size, count int64
		var blobs = make(map[string]int64)

		// usage count
		for _, repo := range project.Repos {
			count = count + int64(len(repo.Afs))
			// Because that there are some shared blobs between repositories, it needs to remove the duplicate items.
			for _, blob := range repo.Blobs {
				_, exist := blobs[blob.Digest]
				if !exist {
					blobs[blob.Digest] = size
				}
			}
		}
		// size
		for _, item := range blobs {
			size = size + item
		}

		proUsage := quota.ProjectUsage{
			Project: project.Name,
			Used: common_quota.ResourceList{
				common_quota.ResourceCount:   count,
				common_quota.ResourceStorage: size,
			},
		}
		pros = append(pros, proUsage)
	}

	return pros, nil
}

// Persist ...
func (rm *RegistryMigrator) Persist(projects []quota.ProjectInfo) error {
	for _, project := range projects {
		for _, repo := range project.Repos {
			if len(repo.Afs) != 0 {
				var wg sync.WaitGroup
				wg.Add(len(repo.Afs))
				for _, af := range repo.Afs {
					go func(interface{}) {
						_, err := dao.AddArtifact(af)
						if err != nil {
							log.Error(err)
						}
					}(af)
				}
				wg.Wait()
			}
			if len(repo.Afnbs) != 0 {
				var wg sync.WaitGroup
				wg.Add(len(repo.Afs))
				for _, afnb := range repo.Afnbs {
					go func(interface{}) {
						_, err := dao.AddArtifactNBlob(afnb)
						if err != nil {
							log.Error(err)
						}
					}(afnb)
				}
				wg.Wait()
			}
			if len(repo.Blobs) != 0 {
				var wg sync.WaitGroup
				wg.Add(len(repo.Afs))
				for _, blob := range repo.Blobs {
					go func(interface{}) {
						_, err := dao.AddBlob(blob)
						if err != nil {
							log.Error(err)
						}
					}(blob)
				}
				wg.Wait()
			}
		}
	}

	return nil
}

func infoOfProject(project string, repoList []string) quota.ProjectInfo {
	var wg sync.WaitGroup
	wg.Add(len(repoList))

	errChan := make(chan error, 1)
	infoChan := make(chan interface{})

	for _, repo := range repoList {
		go func(repo string) {
			defer func() {
				wg.Done()
			}()
			info, err := infoOfRepo(repo)
			if err != nil {
				errChan <- err
				return
			}
			infoChan <- info
		}(repo)
	}

	go func() {
		wg.Wait()
		close(infoChan)
	}()

	var repos []quota.RepoData
	for item := range infoChan {
		repos = append(repos, item.(quota.RepoData))
	}

	return quota.ProjectInfo{
		Name:  project,
		Repos: repos,
	}
}

func infoOfRepo(repo string) (quota.RepoData, error) {
	repoClient, err := coreutils.NewRepositoryClientForUI("harbor-core", repo)
	if err != nil {
		return quota.RepoData{}, err
	}
	tags, err := repoClient.ListTag()
	if err != nil {
		return quota.RepoData{}, err
	}
	var afnbs []*models.ArtifactAndBlob
	var afs []*models.Artifact
	var blobs []*models.Blob

	for _, tag := range tags {
		_, mediaType, payload, err := repoClient.PullManifest(tag, []string{
			schema1.MediaTypeManifest,
			schema1.MediaTypeSignedManifest,
			schema2.MediaTypeManifest,
		})
		if err != nil {
			log.Error(err)
			continue
		}
		manifest, desc, err := registry.UnMarshal(mediaType, payload)
		if err != nil {
			log.Error(err)
			continue
		}
		// self
		afnb := &models.ArtifactAndBlob{
			DigestAF:   desc.Digest.String(),
			DigestBlob: desc.Digest.String(),
		}
		afnbs = append(afnbs, afnb)
		for _, layer := range manifest.References() {
			afnb := &models.ArtifactAndBlob{
				DigestAF:   desc.Digest.String(),
				DigestBlob: layer.Digest.String(),
			}
			afnbs = append(afnbs, afnb)
			blob := &models.Blob{
				Digest:       layer.Digest.String(),
				ContentType:  layer.MediaType,
				Size:         layer.Size,
				CreationTime: time.Now(),
			}
			blobs = append(blobs, blob)
		}
		af := &models.Artifact{
			Repo:         strings.Split(repo, "/")[1],
			Tag:          tag,
			Digest:       desc.Digest.String(),
			Kind:         "Docker-Image",
			CreationTime: time.Now(),
		}
		afs = append(afs, af)
	}
	return quota.RepoData{
		Name:  repo,
		Afs:   afs,
		Afnbs: afnbs,
		Blobs: blobs,
	}, nil
}

func init() {
	quota.Register("registry", NewRegistryMigrator)
}
