package job

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/job/dao"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
)

// Manager manages notification jobs recorded in database
type Manager interface {
	// Create create a notification job
	Create(ctx context.Context, job *model.Job) (int64, error)

	// List list notification jobs
	List(ctx context.Context, query *q.Query) ([]*model.Job, error)

	// Update update notification job
	Update(ctx context.Context, job *model.Job, props ...string) error

	// ListJobsGroupByEventType lists last triggered jobs group by event type
	ListJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error)

	// Count ...
	Count(ctx context.Context, query *q.Query) (total int64, err error)
}

// DefaultManager ..
type DefaultManager struct {
	dao dao.DAO
}

// NewDefaultManager ...
func NewDefaultManager() Manager {
	return &DefaultManager{
		dao: dao.New(),
	}
}

// Create ...
func (d *DefaultManager) Create(ctx context.Context, job *model.Job) (int64, error) {
	return d.dao.Create(ctx, job)
}

// Count ...
func (d *DefaultManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return d.dao.Count(ctx, query)
}

// List ...
func (d *DefaultManager) List(ctx context.Context, query *q.Query) ([]*model.Job, error) {
	return d.dao.List(ctx, query)
}

// Update ...
func (d *DefaultManager) Update(ctx context.Context, job *model.Job, props ...string) error {
	return d.dao.Update(ctx, job, props...)
}

// ListJobsGroupByEventType lists last triggered jobs group by event type
func (d *DefaultManager) ListJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error) {
	return d.dao.GetLastTriggerJobsGroupByEventType(ctx, policyID)
}
