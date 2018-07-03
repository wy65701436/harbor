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

package api

import (
	"errors"
	"net/http"

	"github.com/vmware/harbor/src/common/utils/log"
)

// Ping monitor the docker registry status
func Ping(w http.ResponseWriter, r *http.Request) {
	isRuning, err := IsRegRunning()
	if !isRuning {
		log.Errorf("Registry is not running : %v", err)
		handleInternalServerError(w)
		return
	}

	if err := writeJSON(w, "Pong"); err != nil {
		log.Errorf("Failed to write response: %v", err)
		return
	}
}

// IsRegRunning check the health status
func IsRegRunning() (bool, error) {
	addr := "http://127.0.0.1:5001/debug/health"
	resp, err := http.Get(addr)

	if err != nil {
		log.Errorf("Failed to validate registry status:%v : %v", addr, err)
		return false, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 401 {
		return true, nil
	}

	log.Errorf("Failed to validate registry status:%v : %v", addr, resp.StatusCode)
	return false, errors.New("registry is running abnormal")
}
