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

package dump

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"net/http"
	"net/http/httputil"
)

type dumpHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &dumpHandler{
		next: next,
	}
}

// ServeHTTP ...
func (rh dumpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Info(" @@...... in the dump handler.......")
	log.Info(httputil.DumpRequest(req, true))
	rh.next.ServeHTTP(rw, req)
}
