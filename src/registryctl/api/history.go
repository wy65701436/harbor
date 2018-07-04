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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/vmware/harbor/src/common/utils/log"
)

// GCResults is the file to storage gc result.
var GCResults = "/etc/registry/gcresults"

// GetGCResult ...
func GetGCResult(w http.ResponseWriter, r *http.Request) {
	res, err := ioutil.ReadFile(GCResults)
	if err != nil {
		log.Errorf("Error occured getting gc result: %v", err)
		handleInternalServerError(w)
	}

	if _, err = w.Write(res); err != nil {
		log.Errorf("Failed to write response: %v", err)
		return
	}
}

// GCResult ...
type GCResult struct {
	Status    bool      `json:"status"`
	Msg       string    `json:"msg"`
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
}

// DumpGCResult ...
func (gch *GCResult) DumpGCResult() error {
	if _, err := os.Stat(GCResults); os.IsNotExist(err) {
		_, err := os.Create(GCResults)
		if err != nil {
			return err
		}
	}

	gchJSON, err := json.Marshal(gch)
	if err != nil {
		log.Errorf("Error occured getting gc result: %v", err)
		return err
	}
	if err = ioutil.WriteFile(GCResults, gchJSON, os.FileMode(0644)); err != nil {
		log.Errorf("Error occured writting gc result: %v", err)
		return err
	}
	return nil
}
