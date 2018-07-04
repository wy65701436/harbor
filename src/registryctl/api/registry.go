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
	"bytes"
	"net/http"
	"sync"
	"time"

	"os/exec"

	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	regConf = "/etc/registry/config.yml"
)

var lock *sync.RWMutex = new(sync.RWMutex)

// StartGC ...
func StartGC(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	cmd := exec.Command("/bin/bash", "-c", "registry garbage-collect "+regConf)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	if err := cmd.Run(); err != nil {
		log.Errorf("Fail to execute GC: %s", errBuf.String())
		gcr := GCResult{false, errBuf.String(), start, time.Now()}
		gcr.DumpGCResult()
		handleInternalServerError(w)
	}

	gcr := GCResult{true, outBuf.String(), start, time.Now()}
	gcr.DumpGCResult()
	if err := writeJSON(w, gcr); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
}

func GetGCStatus(w http.ResponseWriter, r *http.Request) {
	return
}
