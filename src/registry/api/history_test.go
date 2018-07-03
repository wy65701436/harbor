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
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDumpGCResult(t *testing.T) {
	startTime := time.Now()
	gcr := GCResult{false, "test", startTime, startTime}
	gcr.DumpGCResult()
	res, _ := ioutil.ReadFile("/etc/regsitry/gcresults")
	assert.Equal(t, "{\"status\":true,\"msg\":\"test\",\"starttime\":\"2018-07-03T09:26:55.284866114Z\",\"endtime\":\"2018-07-03T09:26:55.37799371Z\"}", string(res))
}

func TestGetGCResult(t *testing.T) {
	startTime := time.Now()
	gcr := GCResult{false, "test", startTime, startTime}
	gcr.DumpGCResult()
	res, _ := ioutil.ReadFile("/etc/regsitry/gcresults")
	assert.Equal(t, "{\"status\":true,\"msg\":\"test\",\"starttime\":\"2018-07-03T09:26:55.284866114Z\",\"endtime\":\"2018-07-03T09:26:55.37799371Z\"}", string(res))
}
