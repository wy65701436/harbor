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
	"regexp"

	"github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

const (
	toggleReadOnlyURLPattern = `/internal/togglereadonly`
	loginURLPattern          = `/login`
)

//ReadonlyFilter filters the POST/PUT request and returns 503.
func ReadonlyFilter(ctx *context.Context) {
	filter(ctx.Request, ctx.ResponseWriter)
}

func filter(req *http.Request, resp http.ResponseWriter) {
	if !config.IsReadOnly() {
		return
	}
	log.Info("the url is:", req.URL)
	if req.Method == http.MethodPut {
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if req.Method == http.MethodPost {
		if !matchToggleReadonly(req) || !matchLogin(req, resp) {
			resp.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

func matchToggleReadonly(req *http.Request) bool {
	re := regexp.MustCompile(toggleReadOnlyURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 1 {
		return true
	}
	return false
}

func matchLogin(req *http.Request, resp http.ResponseWriter) bool {
	re := regexp.MustCompile(loginURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) != 1 {
		return false
	}
	sc, err := GetSecurityContext(req)
	if err != nil {
		log.Errorf("failed to get security context: %v", err)
		resp.WriteHeader(http.StatusServiceUnavailable)
	}
	if !sc.IsSysAdmin() {
		return false
	}
	return true
}
