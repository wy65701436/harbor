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

package huggingface

import (
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHuggingFace, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeHuggingFace, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeHuggingFace)
}

func newAdapter(registry *model.Registry) *adapter {
	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapter(registry),
	}
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r), nil
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterInfo()
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

type adapter struct {
	*native.Adapter
	registry *model.Registry
}

var _ adp.Adapter = adapter{}

func (adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeHuggingFace,
		SupportedResourceTypes: []string{
			model.ResourceTypeImage,
		},
		SupportedResourceFilters: []*model.FilterStyle{
			{
				Type:  model.FilterTypeName,
				Style: model.FilterStyleTypeText,
			},
			{
				Type:  model.FilterTypeTag,
				Style: model.FilterStyleTypeText,
			},
		},
		SupportedTriggers: []string{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}, nil
}

func getAdapterInfo() *model.AdapterPattern {
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "huggingface.co",
					Value: "https://huggingface.co/",
				},
			},
		},
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeFix,
			AccessKeyData:    "_json_key",
			AccessSecretType: model.AccessSecretTypeFile,
			AccessSecretData: "No Change",
		},
	}
	return info
}

// HealthCheck checks health status of a registry
func (a adapter) HealthCheck() (string, error) {
	return model.Healthy, nil
}

func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	var resources []*model.Resource
	return resources, nil
}
