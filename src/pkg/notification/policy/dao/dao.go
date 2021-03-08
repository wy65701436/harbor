package dao

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// DAO defines the interface to access the robot data model
type DAO interface {
	// Create ...
	Create(ctx context.Context, n *models.NotificationPolicy) (int64, error)

	// Update ...
	Update(ctx context.Context, n *models.NotificationPolicy) error

	// Get ...
	Get(ctx context.Context, id int64) (*models.NotificationPolicy, error)

	// Count returns the total count of robots according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// List ...
	List(ctx context.Context, query *q.Query) ([]*models.NotificationPolicy, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error
}

// New creates a default implementation for Dao
func New() DAO {
	return &dao{}
}

type dao struct{}

// Get ...
func (d *dao) Get(ctx context.Context, id int64) (*models.NotificationPolicy, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	j := &models.NotificationPolicy{
		ID: id,
	}
	if err := ormer.Read(j); err != nil {
		if e := orm.AsNotFoundError(err, "notificationPolicy %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return j, nil
}

// Create ...
func (d *dao) Create(ctx context.Context, policy *models.NotificationPolicy) (int64, error) {
	if policy == nil {
		return 0, errors.New("nil policy")
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(policy)
	if err != nil {
		if e := orm.AsConflictError(err, "notification policy named %s already exists", policy.Name); e != nil {
			err = e
			return id, err
		}
		err = fmt.Errorf("failed to create the notification policy: %v", err)
		return id, err
	}
	return id, err
}

// Update ...
func (d *dao) Update(ctx context.Context, policy *models.NotificationPolicy) error {
	if policy == nil {
		return errors.New("nil policy")
	}

	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(policy)
	if n == 0 {
		if e := orm.AsConflictError(err, "notification policy named %s already exists", policy.Name); e != nil {
			err = e
		}
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

// Count ...
func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	query.Sorting = ""
	query.PageNumber = 0
	query.PageSize = 0

	qs, err := orm.QuerySetter(ctx, &models.NotificationPolicy{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// List ...
func (d *dao) List(ctx context.Context, query *q.Query) ([]*models.NotificationPolicy, error) {
	policies := []*models.NotificationPolicy{}

	qs, err := orm.QuerySetter(ctx, &models.NotificationPolicy{}, query)
	if err != nil {
		return nil, err
	}
	if query.Sorting != "" {
		qs = qs.OrderBy(query.Sorting)
	} else {
		qs = qs.OrderBy("-CreationTime")
	}
	if _, err = qs.All(&policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// Delete delete notification policy by id
func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&models.NotificationPolicy{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("notificationPolicy %d not found", id)
	}
	return nil
}
