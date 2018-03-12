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

package filter

import (
	beegoctx "github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/common/utils/log"
	"net/http"
	"strconv"
	"os"
)

//ReadonlyFilter filters the POST/PUT request and returns 503.
func ReadonlyFilter() func(*beegoctx.Context) {
	return func(ctx *beegoctx.Context) {
		filter(ctx.Request, ctx.ResponseWriter)
	}
}

func filter(req *http.Request, resp http.ResponseWriter) {
	readOnly := os.Getenv("READ_ONLY")
	isReadOnly, err := strconv.ParseBool(readOnly)
	if err != nil {
		log.Errorf("Failed to parse read only in env, error: %v", err)
	}
	if isReadOnly {
		if req.Method == http.MethodPost || req.Method == http.MethodPut {
			return
		}
		resp.WriteHeader(http.StatusServiceUnavailable)
	}
}
