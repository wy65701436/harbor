package notification

import (
	"context"
	"errors"
	"fmt"
	beegorm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"time"
)

// ArtifactPreprocessHandler preprocess artifact event data
type ArtifactPreprocessHandler struct {
	project *models.Project
}

// Handle preprocess artifact event data and then publish hook event
func (a *ArtifactPreprocessHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}

	time.Sleep(500 * time.Millisecond)

	pushArtEvent, ok := value.(*event.PushArtifactEvent)
	if !ok {
		return errors.New("invalid push artifact event type")
	}
	if pushArtEvent == nil {
		return fmt.Errorf("nil push artifact event")
	}

	var err error
	a.project, err = project.Mgr.Get(pushArtEvent.Artifact.ProjectID)
	if err != nil {
		log.Errorf("failed to get project:%d, error: %v", pushArtEvent.Artifact.ProjectID, err)
		return err
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(a.project.ProjectID, pushArtEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", pushArtEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", pushArtEvent.EventType, pushArtEvent)
		return nil
	}

	payload, err := a.constructArtifactPayload(pushArtEvent)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, pushArtEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

// IsStateful ...
func (a *ArtifactPreprocessHandler) IsStateful() bool {
	return false
}

func (a *ArtifactPreprocessHandler) constructArtifactPayload(event *event.PushArtifactEvent) (*model.Payload, error) {
	repoName := event.Repository
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := models.ProjectPrivate
	if a.project.IsPublic() {
		repoType = models.ProjectPublic
	}

	imageName := getNameFromImgRepoFullName(repoName)

	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         imageName,
				Namespace:    a.project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	ctx := orm.NewContext(context.Background(), beegorm.NewOrm())
	repoRecord, err := repository.Mgr.GetByName(ctx, repoName)
	if err != nil {
		log.Errorf("failed to get repository with name %s: %v", repoName, err)
		return nil, err
	}
	// once repo has been delete, cannot ensure to get repo record
	if repoRecord == nil {
		log.Debugf("cannot find repository info with repo %s", repoName)
	} else {
		payload.EventData.Repository.DateCreated = repoRecord.CreationTime.Unix()
	}

	extURL, err := config.ExtURL()
	if err != nil {
		return nil, fmt.Errorf("get external endpoint failed: %v", err)
	}

	resURL, err := buildImageResourceURL(extURL, repoName, event.Tag)
	if err != nil {
		log.Errorf("get resource URL failed: %v", err)
		return nil, err
	}

	resource := &notifyModel.Resource{
		Tag:         event.Tag,
		Digest:      event.Artifact.Digest,
		ResourceURL: resURL,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
