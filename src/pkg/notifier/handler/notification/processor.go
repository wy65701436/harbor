package notification

import (
	"errors"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
)

func resolveTagEventToImageEvent(value interface{}) (*notifyModel.ImageEvent, error) {
	tagEvent, ok := value.(*notifyModel.TagEvent)
	if !ok || tagEvent == nil {
		return nil, errors.New("invalid image event")
	}
	imageEvent := notifyModel.ImageEvent{
		EventType: notifyModel.PushImageTopic,
		Project:   tagEvent.Project,
		RepoName:  tagEvent.RepoName,
		Resource: []*notifyModel.ImgResource{
			{Tag: tagEvent.TagName},
		},
		OccurAt:  tagEvent.OccurAt,
		Operator: tagEvent.Operator,
	}
	return &imageEvent, nil
}
