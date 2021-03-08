package job

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/job/dao"
)

// Manager manages notification jobs recorded in database
type Manager interface {
	// Create create a notification job
	Create(ctx context.Context, job *models.NotificationJob) (int64, error)

	// List list notification jobs
	List(ctx context.Context, query *q.Query) ([]*models.NotificationJob, error)

	// Update update notification job
	Update(ctx context.Context, job *models.NotificationJob, props ...string) error

	// ListJobsGroupByEventType lists last triggered jobs group by event type
	ListJobsGroupByEventType(ctx context.Context, policyID int64) ([]*models.NotificationJob, error)

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
func (d *DefaultManager) Create(ctx context.Context, job *models.NotificationJob) (int64, error) {
	return d.dao.Create(ctx, job)
}

// Count ...
func (d *DefaultManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return d.dao.Count(ctx, query)
}

// List ...
func (d *DefaultManager) List(ctx context.Context, query *q.Query) ([]*models.NotificationJob, error) {
	return d.dao.List(ctx, query)
}

// Update ...
func (d *DefaultManager) Update(ctx context.Context, job *models.NotificationJob, props ...string) error {
	return d.dao.Update(ctx, job, props...)
}

// ListJobsGroupByEventType lists last triggered jobs group by event type
func (d *DefaultManager) ListJobsGroupByEventType(ctx context.Context, policyID int64) ([]*models.NotificationJob, error) {
	return d.dao.GetLastTriggerJobsGroupByEventType(ctx, policyID)
}
