package gc

import (
	"time"
)

type history struct {
	ID           int64     `json:"id"`
	Trigger      string    `json:"trigger"`
	DryRun       string    `json:"dry_run"`
	Status       string    `json:"status"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}
