package notification

import (
	"context"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notification/hook"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	jobMgr "github.com/goharbor/harbor/src/pkg/notification/job/manager"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notification/policy/manager"
	n_event "github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

var (
	// PolicyMgr is a global notification policy manager
	PolicyMgr policy.Manager

	// JobMgr is a notification job controller
	JobMgr job.Manager

	// HookManager is a hook manager
	HookManager hook.Manager

	// SupportedEventTypes is a map to store supported event type, eg. pushImage, pullImage etc
	SupportedEventTypes map[string]struct{}

	// SupportedNotifyTypes is a map to store notification type, eg. HTTP, Email etc
	SupportedNotifyTypes map[string]struct{}
)

// Init ...
func Init() {
	// init notification policy manager
	PolicyMgr = manager.NewDefaultManger()
	// init hook manager
	HookManager = hook.NewHookManager()
	// init notification job manager
	JobMgr = jobMgr.NewDefaultManager()

	SupportedEventTypes = make(map[string]struct{})
	SupportedNotifyTypes = make(map[string]struct{})

	initSupportedEventType(
		event.TopicPushArtifact, event.TopicPullArtifact, event.TopicDeleteArtifact,
		event.TopicUploadChart, event.TopicDownloadChart, event.TopicDeleteChart,
	)

	initSupportedNotifyType(model.NotifyTypeHTTP, model.NotifyTypeSlack)

	log.Info("notification initialization completed")
}

func initSupportedEventType(eventTypes ...string) {
	for _, eventType := range eventTypes {
		SupportedEventTypes[eventType] = struct{}{}
	}
}

func initSupportedNotifyType(notifyTypes ...string) {
	for _, notifyType := range notifyTypes {
		SupportedNotifyTypes[notifyType] = struct{}{}
	}
}

type eventKey struct{}

// FromContext returns event from context
func FromContext(ctx context.Context) (n_event.Metadata, error) {
	o, ok := ctx.Value(eventKey{}).(n_event.Metadata)
	if !ok {
		return nil, errors.New("cannot get the EVENT from context")
	}
	return o, nil
}

// NewContext returns new context with event
func NewContext(ctx context.Context, m interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, eventKey{}, m)
}

// AddEvent ....
func AddEvent(ctx context.Context, m n_event.Metadata) error {
	e, ok := ctx.Value(eventKey{}).(*interface{})
	if !ok {
		fmt.Println("1111111")
		return nil
	}
	*e = m
	return nil
}
