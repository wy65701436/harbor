package gc

import (
	"encoding/json"
	"fmt"
	common_models "github.com/goharbor/harbor/src/common/models"
	common_utils "github.com/goharbor/harbor/src/common/utils"
	"strings"
	"time"
)

// Schedule ...
type Schedule struct {
	Schedule *ScheduleParam `json:"schedule"`
}

// ScheduleParam defines the parameter of schedule trigger
type ScheduleParam struct {
	// Daily, Weekly, Custom, Manual, None
	Type string `json:"type"`
	// The cron string of scheduled job
	Cron string `json:"cron"`
}

// History gc execution history
type History struct {
	Schedule
	ID           int64     `json:"id"`
	Name         string    `json:"job_name"`
	Kind         string    `json:"job_kind"`
	Parameters   string    `json:"job_parameters"`
	Status       string    `json:"job_status"`
	UUID         string    `json:"-"`
	Deleted      bool      `json:"deleted"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// ConvertSchedule converts different kinds of cron string into one standard for UI to show.
// in the latest design, it uses {"type":"Daily","cron":"0 0 0 * * *"} as the cron item.
// As for supporting migration from older version, it needs to convert {"parameter":{"daily_time":0},"type":"daily"}
// and {"type":"Daily","weekday":0,"offtime":57600} into one standard.
func ConvertSchedule(cronStr string) (ScheduleParam, error) {
	if cronStr == "" {
		return ScheduleParam{}, nil
	}

	convertedSchedule := ScheduleParam{}
	convertedSchedule.Type = "custom"

	if strings.Contains(cronStr, "parameter") {
		scheduleModel := ScanAllPolicy{}
		if err := json.Unmarshal([]byte(cronStr), &scheduleModel); err != nil {
			return ScheduleParam{}, err
		}
		h, m, s := common_utils.ParseOfftime(int64(scheduleModel.Param["daily_time"].(float64)))
		cron := fmt.Sprintf("%d %d %d * * *", s, m, h)
		convertedSchedule.Cron = cron
		return convertedSchedule, nil
	} else if strings.Contains(cronStr, "offtime") {
		scheduleModel := common_models.ScheduleParam{}
		if err := json.Unmarshal([]byte(cronStr), &scheduleModel); err != nil {
			return ScheduleParam{}, err
		}
		convertedSchedule.Cron = common_models.ParseScheduleParamToCron(&scheduleModel)
		return convertedSchedule, nil
	} else if strings.Contains(cronStr, "cron") {
		scheduleModel := ScheduleParam{}
		if err := json.Unmarshal([]byte(cronStr), &scheduleModel); err != nil {
			return ScheduleParam{}, err
		}
		return scheduleModel, nil
	}

	return ScheduleParam{}, fmt.Errorf("unsupported cron format, %s", cronStr)
}

// ScanAllPolicy is represent the json request and object for scan all policy
// Only for migrating from the legacy schedule.
type ScanAllPolicy struct {
	Type  string                 `json:"type"`
	Param map[string]interface{} `json:"parameter,omitempty"`
}
