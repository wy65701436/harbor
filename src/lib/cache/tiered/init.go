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

package tiered

import (
	"os"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/log"
)

// InitializeTieredCache initializes the tiered cache from environment variables
// This should be called during Harbor initialization
func InitializeTieredCache() (cache.Cache, error) {
	config := DefaultConfig()

	// Read configuration from environment variables
	if val := os.Getenv("HARBOR_CACHE_L1_MAX_SIZE_MB"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.L1MaxSize = size * 1024 * 1024
			log.Infof("Tiered cache L1 max size set to %dMB", size)
		}
	}

	if val := os.Getenv("HARBOR_CACHE_L1_MAX_ENTRIES"); val != "" {
		if entries, err := strconv.Atoi(val); err == nil {
			config.L1MaxEntries = entries
			log.Infof("Tiered cache L1 max entries set to %d", entries)
		}
	}

	if val := os.Getenv("HARBOR_CACHE_L1_DEFAULT_TTL_MINUTES"); val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.L1DefaultTTL = time.Duration(minutes) * time.Minute
			log.Infof("Tiered cache L1 default TTL set to %d minutes", minutes)
		}
	}

	if val := os.Getenv("HARBOR_CACHE_L2_DEFAULT_TTL_MINUTES"); val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.L2DefaultTTL = time.Duration(minutes) * time.Minute
			log.Infof("Tiered cache L2 default TTL set to %d minutes", minutes)
		}
	}

	// Redis address from environment
	if val := os.Getenv("_REDIS_URL_CORE"); val != "" {
		config.L2Address = val
	}

	// Enable/disable flags
	if val := os.Getenv("HARBOR_CACHE_L1_ENABLED"); val != "" {
		config.EnableL1 = val == "true"
		log.Infof("Tiered cache L1 enabled: %v", config.EnableL1)
	}

	if val := os.Getenv("HARBOR_CACHE_L2_ENABLED"); val != "" {
		config.EnableL2 = val == "true"
		log.Infof("Tiered cache L2 enabled: %v", config.EnableL2)
	}

	// Content-specific TTLs
	if val := os.Getenv("HARBOR_CACHE_MANIFEST_DIGEST_TTL_HOURS"); val != "" {
		if hours, err := strconv.Atoi(val); err == nil {
			config.ManifestByDigestTTL = time.Duration(hours) * time.Hour
		}
	}

	if val := os.Getenv("HARBOR_CACHE_MANIFEST_TAG_TTL_MINUTES"); val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.ManifestByTagTTL = time.Duration(minutes) * time.Minute
		}
	}

	if val := os.Getenv("HARBOR_CACHE_PROJECT_META_TTL_MINUTES"); val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.ProjectMetaTTL = time.Duration(minutes) * time.Minute
		}
	}

	if val := os.Getenv("HARBOR_CACHE_ARTIFACT_META_TTL_MINUTES"); val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.ArtifactMetaTTL = time.Duration(minutes) * time.Minute
		}
	}

	log.Infof("Initializing tiered cache with config: L1Size=%dMB, L1Entries=%d, L1TTL=%v, L2TTL=%v",
		config.L1MaxSize/(1024*1024), config.L1MaxEntries, config.L1DefaultTTL, config.L2DefaultTTL)

	return NewTieredCache(config)
}
