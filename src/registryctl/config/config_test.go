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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDoesNotExists(t *testing.T) {
	cfg := &Configuration{}
	if err := cfg.Load("./config.not-existing.yaml", false); err == nil {
		t.Fatalf("Load config from none-existing document, expect none nil error but got '%s'\n", err)
	}
}

func TestConfigLoadingWithEnv(t *testing.T) {
	os.Setenv("REGISTRYCTL_PROTOCOL", "https")
	os.Setenv("PORT", "1000")
	os.Setenv("LOG_LEVEL", "DEBUG")

	cfg := &Configuration{}
	assert.Equal(t, "https", cfg.Protocol)
	assert.Equal(t, "1000", cfg.Port)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
}
