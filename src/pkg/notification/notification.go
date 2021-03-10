package notification

import (
	"container/list"
	"context"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification/hook"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	n_event "github.com/goharbor/harbor/src/pkg/notifier/event"
)

var (
	// PolicyMgr is a global notification policy manager
	PolicyMgr policy.Manager

	// JobMgr is a notification job controller
	JobMgr job.Manager

	// HookManager is a hook manager
	HookManager hook.Manager
)

// Init ...
func Init() {
	// init notification policy manager
	PolicyMgr = policy.Mgr
	// init hook manager
	HookManager = hook.NewHookManager()
	// init notification job manager
	JobMgr = job.Mgr

	log.Info("notification initialization completed")
}

type eventKey struct{}

// EventCtx ...
type EventCtx struct {
	Events     *list.List
	MustNotify bool
}

// NewEventCtx returns instance of EventCtx
func NewEventCtx() *EventCtx {
	return &EventCtx{
		Events:     list.New(),
		MustNotify: false,
	}
}

// NewContext returns new context with event
func NewContext(ctx context.Context, ec *EventCtx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, eventKey{}, ec)
}

// AddEvent add events into request context, the event will be sent by the notification middleware eventually.
func AddEvent(ctx context.Context, m n_event.Metadata, notify ...bool) {
	if m == nil {
		return
	}

	e, ok := ctx.Value(eventKey{}).(*EventCtx)
	if !ok {
		log.Debug("request has not event list, cannot add event into context")
		return
	}
	if len(notify) != 0 {
		e.MustNotify = notify[0]
	}
	e.Events.PushBack(m)
	return
}
