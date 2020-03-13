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
	"context"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"net/http"

	"github.com/goharbor/harbor/src/internal"
	evt "github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/server/middleware"
)

type EventContext struct {
	context.Context
	Events []event.Metadata
}

// AddEvent ....
func AddEvent(ctx context.Context, e event.Metadata) error {
	c, ok := ctx.(EventContext)
	if !ok {
		return fmt.Errorf("%s URL %s without event, no event send", r.Method, r.URL.Path)
	}
	c.Events = append(c.Events, e)
	return nil
}

// publishEvent publishes the events in the context, it ensures publish happens after transaction success.
func publishEvent(r *http.Request) error {
	c, ok := r.Context().(EventContext)
	if !ok {
		return fmt.Errorf("%s URL %s without event, no event send", r.Method, r.URL.Path)
	}

	if len(c.Events) != 0 {
		return nil
	}

	for _, e := range c.Events {
		evt.BuildAndPublish(e)
	}

	return nil
}

type eventKey struct{}

// FromContext returns orm from context
func FromContext(ctx context.Context) (orm.Ormer, error) {
	o, ok := ctx.Value(eventKey{}).(orm.Ormer)
	if !ok {
		return nil, errors.New("cannot get the Event from context")
	}
	return o, nil
}

// NewContext returns new context with orm
func NewContext(ctx context.Context, o orm.Ormer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, eventKey{}, o)
}

// Middleware sends the notification after transaction success
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		res, ok := w.(*internal.ResponseBuffer)
		if !ok {
			res = internal.NewResponseBuffer(w)
			defer res.Flush()
		}
		ec := &EventContext{Context: r.Context()}
		next.ServeHTTP(res, r.WithContext(ec))
		if res.Success() {
			if err := publishEvent(r); err != nil {
				log.Errorf("send webhook error, %v", err)
			}
		}
	}, skippers...)
}
