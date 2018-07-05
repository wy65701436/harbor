// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package registryctl

import (
	"os"

	"github.com/vmware/harbor/src/common"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/secret"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/registryctl/client"
)

const (
	defaultKeyPath = "/etc/ui/key"
)

var (
	// SecretStore manages secrets
	SecretStore *secret.Store
	// RegistryCtlClient is a client for registry
	RegistryCtlClient client.Client

	keyProvider comcfg.KeyProvider
)

// Init ...
func Init() {
	initKeyProvider()
	initRegistryCtlClient()
}

func initRegistryCtlClient() {
	registryCtlURL := os.Getenv("REGISTRY_CONTROLLER_URL")
	if len(registryCtlURL) == 0 {
		registryCtlURL = common.DefaultRegistryControllerEndpoint
	}

	log.Infof("initializing client for reigstry %s ...", registryCtlURL)
	cfg := &client.Config{
		Secret: os.Getenv("UI_SECRET"),
	}
	RegistryCtlClient = client.NewClient(registryCtlURL, cfg)
}

func initKeyProvider() {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}
	log.Infof("key path: %s", path)

	keyProvider = comcfg.NewFileKeyProvider(path)
}
