package event

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/pkg/errors"
)

const (
	autoTriggeredOperator = "auto"
)

// Event to publish
type Event struct {
	Topic string
	Data  interface{}
}

// TopicEvent - Events that contains topic information
type TopicEvent interface {
	Topic() string
}

// New ...
func New() *Event {
	return &Event{}
}

// WithTopicEvent - builder method
func (e *Event) WithTopicEvent(topicEvent TopicEvent) *Event {
	e.Topic = topicEvent.Topic()
	e.Data = topicEvent
	return e
}

// Metadata is the event raw data to be processed
type Metadata interface {
	Resolve(event *Event) error
}

// ImageDelMetaData defines images deleting related event data
type ImageDelMetaData struct {
	Project  *models.Project
	Tags     []string
	Digests  map[string]string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image deleting metadata into common image event
func (i *ImageDelMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		EventType: notifyModel.EventTypeDeleteImage,
		Project:   i.Project,
		OccurAt:   i.OccurAt,
		Operator:  i.Operator,
		RepoName:  i.RepoName,
	}
	for _, t := range i.Tags {
		res := &model.ImgResource{
			Tag:    t,
			Digest: i.Digests[t],
		}
		data.Resource = append(data.Resource, res)
	}
	evt.Topic = model.DeleteImageTopic
	evt.Data = data
	return nil
}

// ImagePushMetaData defines images pushing related event data
type ImagePushMetaData struct {
	Project  *models.Project
	Tag      string
	Digest   string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image pushing metadata into common image event
func (i *ImagePushMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		EventType: notifyModel.EventTypePushImage,
		Project:   i.Project,
		OccurAt:   i.OccurAt,
		Operator:  i.Operator,
		RepoName:  i.RepoName,
		Resource: []*model.ImgResource{
			{
				Tag:    i.Tag,
				Digest: i.Digest,
			},
		},
	}

	evt.Topic = model.PushImageTopic
	evt.Data = data
	return nil
}

// ImagePullMetaData defines images pulling related event data
type ImagePullMetaData struct {
	Project  *models.Project
	Tag      string
	Digest   string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image pulling metadata into common image event
func (i *ImagePullMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		EventType: notifyModel.EventTypePullImage,
		Project:   i.Project,
		OccurAt:   i.OccurAt,
		Operator:  i.Operator,
		RepoName:  i.RepoName,
		Resource: []*model.ImgResource{
			{
				Tag:    i.Tag,
				Digest: i.Digest,
			},
		},
	}

	evt.Topic = model.PullImageTopic
	evt.Data = data
	return nil
}

// ChartMetaData defines meta data of chart event
type ChartMetaData struct {
	ProjectName string
	ChartName   string
	Versions    []string
	OccurAt     time.Time
	Operator    string
}

func (cmd *ChartMetaData) convert(evt *model.ChartEvent) {
	evt.ProjectName = cmd.ProjectName
	evt.OccurAt = cmd.OccurAt
	evt.Operator = cmd.Operator
	evt.ChartName = cmd.ChartName
	evt.Versions = cmd.Versions
}

// ChartUploadMetaData defines meta data of chart upload event
type ChartUploadMetaData struct {
	ChartMetaData
}

// Resolve chart uploading metadata into common chart event
func (cu *ChartUploadMetaData) Resolve(evt *Event) error {
	data := &model.ChartEvent{
		EventType: notifyModel.EventTypeUploadChart,
	}
	cu.convert(data)

	evt.Topic = model.UploadChartTopic
	evt.Data = data
	return nil
}

// ChartDownloadMetaData defines meta data of chart download event
type ChartDownloadMetaData struct {
	ChartMetaData
}

// Resolve chart download metadata into common chart event
func (cd *ChartDownloadMetaData) Resolve(evt *Event) error {
	data := &model.ChartEvent{
		EventType: notifyModel.EventTypeDownloadChart,
	}
	cd.convert(data)

	evt.Topic = model.DownloadChartTopic
	evt.Data = data
	return nil
}

// ChartDeleteMetaData defines meta data of chart delete event
type ChartDeleteMetaData struct {
	ChartMetaData
}

// Resolve chart delete metadata into common chart event
func (cd *ChartDeleteMetaData) Resolve(evt *Event) error {
	data := &model.ChartEvent{
		EventType: notifyModel.EventTypeDeleteChart,
	}
	cd.convert(data)

	evt.Topic = model.DeleteChartTopic
	evt.Data = data
	return nil
}

// HookMetaData defines hook notification related event data
type HookMetaData struct {
	PolicyID  int64
	EventType string
	Target    *models.EventTarget
	Payload   *model.Payload
}

// Resolve hook metadata into hook event
func (h *HookMetaData) Resolve(evt *Event) error {
	data := &model.HookEvent{
		PolicyID:  h.PolicyID,
		EventType: h.EventType,
		Target:    h.Target,
		Payload:   h.Payload,
	}

	evt.Topic = h.Target.Type
	evt.Data = data
	return nil
}

// Build an event by metadata
func (e *Event) Build(metadata ...Metadata) error {
	for _, md := range metadata {
		if err := md.Resolve(e); err != nil {
			log.Debugf("failed to resolve event metadata: %v", md)
			return errors.Wrap(err, "failed to resolve event metadata")
		}
	}
	return nil
}

// Publish an event
func (e *Event) Publish() error {
	if err := notifier.Publish(e.Topic, e.Data); err != nil {
		log.Debugf("failed to publish topic %s with event: %v", e.Topic, e.Data)
		return errors.Wrap(err, "failed to publish event")
	}
	return nil
}

// BuildAndPublish builds the event according to the metadata and publish the event
// The process is done in a separated goroutine
func BuildAndPublish(metadata ...Metadata) {
	go func() {
		event := &Event{}
		if err := event.Build(metadata...); err != nil {
			log.Errorf("failed to build the event from metadata: %v", err)
			return
		}
		if err := event.Publish(); err != nil {
			log.Errorf("failed to publish the event %s: %v", event.Topic, err)
			return
		}
		log.Debugf("event %s published", event.Topic)
	}()
}
