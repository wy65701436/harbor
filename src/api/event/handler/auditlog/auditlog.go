package auditlog

import (
	"context"
	beegorm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/audit"
	am "github.com/goharbor/harbor/src/pkg/audit/model"
	"github.com/goharbor/harbor/src/pkg/notifier"
)

func init() {
	handler := &Handler{}
	notifier.Subscribe(event.TopicPushArtifact, handler)
	notifier.Subscribe(event.TopicPullArtifact, handler)
	notifier.Subscribe(event.TopicDeleteArtifact, handler)
}

// Handler - audit log handler
type Handler struct {
}

// AuditResolver - interface to resolve to AuditLog
type AuditResolver interface {
	ResolveToAuditLog() (*am.AuditLog, error)
}

// Handle ...
func (h *Handler) Handle(value interface{}) error {
	ctx := orm.NewContext(context.Background(), beegorm.NewOrm())
	var auditLog *am.AuditLog
	switch v := value.(type) {
	case *event.PushArtifactEvent:
		resolver := value.(AuditResolver)
		al, err := resolver.ResolveToAuditLog()
		if err != nil {
			log.Errorf("failed to handler event %v", err)
			return err
		}
		auditLog = al
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	if auditLog != nil {
		_, err := audit.Mgr.Create(ctx, auditLog)
		if err != nil {
			log.Debugf("add audit log err: %v", err)
		}
	}
	return nil
}

// IsStateful ...
func (h *Handler) IsStateful() bool {
	return false
}
