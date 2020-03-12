package artifact

import (
	"context"
	"fmt"
	beegorm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/api/event/handler"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
)

func init() {
	handler := &ArtifactHandler{}
	notifier.Subscribe(event.TopicPushArtifact, handler)
	notifier.Subscribe(event.TopicDeleteArtifact, handler)
}

// ArtifactHandler preprocess artifact event data
type ArtifactHandler struct {
	project *models.Project
}

// Handle preprocess artifact event data and then publish hook event
func (a *ArtifactHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}
	pushArtEvent, ok := value.(*event.PushArtifactEvent)
	if ok {
		return a.handle(pushArtEvent.ArtifactEvent)
	}
	pullArtEvent, ok := value.(*event.PullArtifactEvent)
	if ok {
		return a.handle(pullArtEvent.ArtifactEvent)
	}
	return nil
}

// IsStateful ...
func (a *ArtifactHandler) IsStateful() bool {
	return false
}

func (a *ArtifactHandler) handle(event *event.ArtifactEvent) error {
	var err error
	a.project, err = project.Mgr.Get(event.Artifact.ProjectID)
	if err != nil {
		log.Errorf("failed to get project:%d, error: %v", event.Artifact.ProjectID, err)
		return err
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(a.project.ProjectID, event.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", event.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", event.EventType, event)
		return nil
	}

	payload, err := a.constructArtifactPayload(event)
	if err != nil {
		return err
	}

	err = handler.SendHookWithPolicies(policies, payload, event.EventType)
	if err != nil {
		return err
	}
	return nil
}

func (a *ArtifactHandler) constructArtifactPayload(event *event.ArtifactEvent) (*model.Payload, error) {
	repoName := event.Repository
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := models.ProjectPrivate
	if a.project.IsPublic() {
		repoType = models.ProjectPublic
	}

	imageName := handler.GetNameFromImgRepoFullName(repoName)

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
	payload.EventData.Repository.DateCreated = repoRecord.CreationTime.Unix()

	resURL, err := handler.BuildImageResourceURL(repoName, event.Tag)
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
