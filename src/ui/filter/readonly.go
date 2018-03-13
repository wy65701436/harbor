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
	"net/http"

	"github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/ui/config"
)

//ReadonlyFilter filters the POST/PUT request and returns 503.
func ReadonlyFilter() func(ctx *context.Context) {
	return func(ctx *context.Context) {
		filter(ctx.Request, ctx.ResponseWriter)
	}
}

func filter(req *http.Request, resp http.ResponseWriter) {
	if !config.IsReadOnly() {
		return
	}
	// Any data updates will be blocked.
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		resp.WriteHeader(http.StatusServiceUnavailable)
	}
}
