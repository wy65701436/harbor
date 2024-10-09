// Code generated by mockery v2.46.2. DO NOT EDIT.

package quota

import (
	context "context"

	models "github.com/goharbor/harbor/src/pkg/quota/models"
	mock "github.com/stretchr/testify/mock"

	q "github.com/goharbor/harbor/src/lib/q"

	types "github.com/goharbor/harbor/src/pkg/quota/types"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Count provides a mock function with given fields: ctx, query
func (_m *Manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for Count")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) (int64, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) int64); ok {
		r0 = rf(ctx, query)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *q.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: ctx, reference, referenceID, hardLimits, usages
func (_m *Manager) Create(ctx context.Context, reference string, referenceID string, hardLimits types.ResourceList, usages ...types.ResourceList) (int64, error) {
	_va := make([]interface{}, len(usages))
	for _i := range usages {
		_va[_i] = usages[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, reference, referenceID, hardLimits)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, types.ResourceList, ...types.ResourceList) (int64, error)); ok {
		return rf(ctx, reference, referenceID, hardLimits, usages...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, types.ResourceList, ...types.ResourceList) int64); ok {
		r0 = rf(ctx, reference, referenceID, hardLimits, usages...)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, types.ResourceList, ...types.ResourceList) error); ok {
		r1 = rf(ctx, reference, referenceID, hardLimits, usages...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *Manager) Delete(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *Manager) Get(ctx context.Context, id int64) (*models.Quota, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.Quota
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*models.Quota, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *models.Quota); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Quota)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByRef provides a mock function with given fields: ctx, reference, referenceID
func (_m *Manager) GetByRef(ctx context.Context, reference string, referenceID string) (*models.Quota, error) {
	ret := _m.Called(ctx, reference, referenceID)

	if len(ret) == 0 {
		panic("no return value specified for GetByRef")
	}

	var r0 *models.Quota
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*models.Quota, error)); ok {
		return rf(ctx, reference, referenceID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *models.Quota); ok {
		r0 = rf(ctx, reference, referenceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Quota)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, reference, referenceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, query
func (_m *Manager) List(ctx context.Context, query *q.Query) ([]*models.Quota, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*models.Quota
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) ([]*models.Quota, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) []*models.Quota); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Quota)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *q.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *Manager) Update(ctx context.Context, _a1 *models.Quota) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Quota) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
