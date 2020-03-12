package event

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
	"time"
)

const (
	autoTriggeredOperator = "auto"
)

// ScanImageMetaData defines meta data of image scanning event
type ScanImageMetaData struct {
	Artifact *v1.Artifact
	Status   string
}

// Resolve image scanning metadata into common chart event
func (si *ScanImageMetaData) Resolve(evt *event.Event) error {
	var eventType string
	var topic string

	switch si.Status {
	case models.JobFinished:
		eventType = notifyModel.EventTypeScanningCompleted
		topic = TopicScanningCompleted
	case models.JobError, models.JobStopped:
		eventType = notifyModel.EventTypeScanningFailed
		topic = TopicScanningFailed
	default:
		return errors.New("not supported scan hook status")
	}

	data := &model.ScanImageEvent{
		EventType: eventType,
		Artifact:  si.Artifact,
		OccurAt:   time.Now(),
		Operator:  autoTriggeredOperator,
	}

	evt.Topic = topic
	evt.Data = data
	return nil
}
