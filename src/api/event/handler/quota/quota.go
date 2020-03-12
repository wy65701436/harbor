package notification

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/api/event/handler"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/project"
)

func init() {
	handler := &QuotaHandler{}
	notifier.Subscribe(event.TopicQuotaExceed, handler)
	notifier.Subscribe(event.TopicQuotaWarning, handler)
}

// QuotaHandler preprocess image event data
type QuotaHandler struct {
}

// Handle ...
func (qp *QuotaHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	quotaEvent, ok := value.(*event.QuotaEvent)
	if !ok {
		return errors.New("invalid quota event type")
	}
	if quotaEvent == nil {
		return fmt.Errorf("nil quota event")
	}

	project, err := project.Mgr.Get(quotaEvent.Project.Name)
	if err != nil {
		log.Errorf("failed to get project:%s, error: %v", quotaEvent.Project.Name, err)
		return err
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(project.ProjectID, quotaEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", quotaEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", quotaEvent.EventType, quotaEvent)
		return nil
	}

	payload, err := constructQuotaPayload(quotaEvent)
	if err != nil {
		return err
	}

	err = handler.SendHookWithPolicies(policies, payload, quotaEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (qp *QuotaHandler) IsStateful() bool {
	return false
}

func constructQuotaPayload(event *event.QuotaEvent) (*model.Payload, error) {
	repoName := event.RepoName
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := models.ProjectPrivate
	if event.Project.IsPublic() {
		repoType = models.ProjectPublic
	}

	imageName := handler.GetNameFromImgRepoFullName(repoName)
	quotaCustom := make(map[string]string)
	quotaCustom["Details"] = event.Msg

	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         imageName,
				Namespace:    event.Project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
			Custom: quotaCustom,
		},
	}
	resource := &notifyModel.Resource{
		Tag:    event.Resource.Tag,
		Digest: event.Resource.Digest,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
