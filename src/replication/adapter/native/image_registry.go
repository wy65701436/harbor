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

package native

import (
	"errors"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

var _ adp.ImageRegistry = native{}

func (n native) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	if len(namespaces) > 0 {
		return nil, errors.New("native registry adapter not support namespace")
	}

	nameFilterPattern := ""
	tagFilterPattern := ""
	for _, filter := range filters {
		switch filter.Type {
		case model.FilterTypeName:
			nameFilterPattern = filter.Value.(string)
		case model.FilterTypeTag:
			tagFilterPattern = filter.Value.(string)
		}
	}
	repositories, err := n.filterRepositories(nameFilterPattern)
	if err != nil {
		return nil, err
	}

	resources := []*model.Resource{}
	for _, repository := range repositories {
		tags, err := n.filterTags(repository, tagFilterPattern)
		if err != nil {
			return nil, err
		}
		if len(tags) == 0 {
			continue
		}
		resources = append(resources, &model.Resource{
			Type:     model.ResourceTypeRepository,
			Registry: n.registry,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: repository,
				},
				Vtags: tags,
			},
		})
	}

	return resources, nil
}

func (n native) filterRepositories(pattern string) ([]string, error) {
	// if the pattern contains no "*" and "?", it is a specific repository name
	// just to make sure the repository exists
	if len(pattern) > 0 && !strings.ContainsAny(pattern, "*?") {
		_, err := n.ListTag(pattern)
		// the repository exists
		if err == nil {
			return []string{pattern}, nil
		}
		// the repository doesn't exist
		if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusNotFound {
			return nil, nil
		}
		// other error
		return nil, err
	}
	// search repositories from catalog api
	repositories, err := n.Catalog()
	if err != nil {
		return nil, err
	}
	// if the pattern is null, just return the result of catalog API
	if len(pattern) == 0 {
		return repositories, nil
	}
	result := []string{}
	for _, repository := range repositories {
		match, err := util.Match(pattern, repository)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, repository)
		}
	}
	return result, nil
}

func (n native) filterTags(repository, pattern string) ([]string, error) {
	tags, err := n.ListTag(repository)
	if err != nil {
		return nil, err
	}
	if len(pattern) == 0 {
		return tags, nil
	}

	result := []string{}
	for _, tag := range tags {
		match, err := util.Match(pattern, tag)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, tag)
		}
	}
	return result, nil
}
