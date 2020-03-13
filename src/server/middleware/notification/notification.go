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

package notification

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"

	"github.com/goharbor/harbor/src/internal"
	evt "github.com/goharbor/harbor/src/pkg/notifier/event"
)

// publishEvent publishes the events in the context, it ensures publish happens after transaction success.
func publishEvent(r *http.Request) error {
	es, err := notification.FromContext(r.Context())
	if err != nil {
		return nil
	}
	if es == nil {
		fmt.Println("5555555")
		return nil
	}
	fmt.Println(es)
	for _, e := range es {
		evt.BuildAndPublish(*e)
	}
	return nil
}

// Middleware sends the notification after transaction success
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		res, ok := w.(*internal.ResponseBuffer)
		if !ok {
			res = internal.NewResponseBuffer(w)
			defer res.Flush()
		}
		placeholder := make(notification.Events, 0)
		ctx := notification.NewContext(r.Context(), placeholder)
		next.ServeHTTP(res, r.WithContext(ctx))
		if res.Success() {
			if err := publishEvent(r); err != nil {
				log.Errorf("send webhook error, %v", err)
			}
		}
	}, skippers...)
}
