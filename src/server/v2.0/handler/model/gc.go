package model

import (
	"encoding/json"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

const (
	// ScheduleHourly : 'Hourly'
	ScheduleHourly = "Hourly"
	// ScheduleDaily : 'Daily'
	ScheduleDaily = "Daily"
	// ScheduleWeekly : 'Weekly'
	ScheduleWeekly = "Weekly"
	// ScheduleCustom : 'Custom'
	ScheduleCustom = "Custom"
	// ScheduleManual : 'Manual'
	ScheduleManual = "Manual"
	// ScheduleNone : 'None'
	ScheduleNone = "None"
)

type GCHistory struct {
	*gc.History
}

// ToSwagger converts the history to the swagger model
func (h *GCHistory) ToSwagger() *models.GCHistory {
	return &models.GCHistory{
		ID:            h.ID,
		JobName:       h.Name,
		JobKind:       h.Kind,
		JobParameters: h.Parameters,
		Deleted:       h.Deleted,
		JobStatus:     h.Status,
		CreationTime:  strfmt.DateTime(h.CreationTime),
		UpdateTime:    strfmt.DateTime(h.UpdateTime),
	}
}

type Schedule struct {
	*scheduler.Schedule
}

// A GC schedule
//{
//    "schedule": {
//        "type": "Daily",
//        "cron": "0 0 0 * * *"
//    },
//    "job_parameters": "{\"delete_untagged\":false,\"time_window\":0}",
//}
// TODO remove the hard code when after issue https://github.com/goharbor/harbor/issues/13047 is resolved.
// ToSwagger converts the schedule to the swagger model
func (s *Schedule) ToSwagger() *models.GCHistory {
	sche := &models.ScheduleObj{
		Cron: s.CRON,
		Type: "Custom",
	}
	para := make(map[string]interface{})
	para["delete_untagged"] = true
	paraStr, err := json.Marshal(para)
	if err != nil {
		return nil
	}

	return &models.GCHistory{
		ID:            0,
		JobName:       "",
		JobKind:       s.CRON,
		JobParameters: string(paraStr),
		Deleted:       false,
		JobStatus:     "",
		Schedule:      sche,
		CreationTime:  strfmt.DateTime(s.CreationTime),
		UpdateTime:    strfmt.DateTime(s.UpdateTime),
	}
}

// NewSchedule ...
func NewSchedule(s *scheduler.Schedule) *Schedule {
	return &Schedule{Schedule: s}
}
