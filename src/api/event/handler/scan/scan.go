package scan

import (
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/api/event/handler"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"time"

	"github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/project"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
)

func init() {
	handler := &ScanHandler{}
	notifier.Subscribe(event.TopicScanningFailed, handler)
	notifier.Subscribe(event.TopicScanningCompleted, handler)
}

// ScanHandler preprocess scan artifact event
type ScanHandler struct {
}

// Handle preprocess chart event data and then publish hook event
func (si *ScanHandler) Handle(value interface{}) error {
	// if global notification configured disabled, return directly
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	if value == nil {
		return errors.New("empty scan artifact event")
	}

	e, ok := value.(*event.ScanImageEvent)
	if !ok {
		return errors.New("invalid scan artifact event type")
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(e.Artifact.NamespaceID, e.EventType)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	// If we cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("Cannot find policy for %s event: %v", e.EventType, e)
		return nil
	}

	// Get project
	project, err := project.Mgr.Get(e.Artifact.NamespaceID)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	payload, err := constructScanImagePayload(e, project)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	err = handler.SendHookWithPolicies(policies, payload, e.EventType)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	return nil
}

// IsStateful ...
func (si *ScanHandler) IsStateful() bool {
	return false
}

func constructScanImagePayload(event *event.ScanImageEvent, project *models.Project) (*model.Payload, error) {
	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	repoName := handler.GetNameFromImgRepoFullName(event.Artifact.Repository)

	payload := &model.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         repoName,
				Namespace:    project.Name,
				RepoFullName: event.Artifact.Repository,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	resURL, err := handler.BuildImageResourceURL(event.Artifact.Repository, event.Artifact.Tag)
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	// Wait for reasonable time to make sure the report is ready
	// Interval=500ms and total time = 5s
	// If the report is still not ready in the total time, then failed at then
	for i := 0; i < 10; i++ {
		// First check in case it is ready
		if re, err := scan.DefaultController.GetReport(event.Artifact, []string{v1.MimeTypeNativeReport}); err == nil {
			if len(re) > 0 && len(re[0].Report) > 0 {
				break
			}
		} else {
			log.Error(errors.Wrap(err, "construct scan payload: wait report ready loop"))
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Add scan overview
	summaries, err := scan.DefaultController.GetSummary(event.Artifact, []string{v1.MimeTypeNativeReport})
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	resource := &model.Resource{
		Tag:          event.Artifact.Tag,
		Digest:       event.Artifact.Digest,
		ResourceURL:  resURL,
		ScanOverview: summaries,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
