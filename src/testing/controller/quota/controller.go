// Code generated by mockery v2.46.2. DO NOT EDIT.

package quota

import (
	context "context"

	models "github.com/goharbor/harbor/src/pkg/quota/models"
	mock "github.com/stretchr/testify/mock"

	q "github.com/goharbor/harbor/src/lib/q"

	quota "github.com/goharbor/harbor/src/controller/quota"

	types "github.com/goharbor/harbor/src/pkg/quota/types"
)

// Controller is an autogenerated mock type for the Controller type
type Controller struct {
	mock.Mock
}

// Count provides a mock function with given fields: ctx, query
func (_m *Controller) Count(ctx context.Context, query *q.Query) (int64, error) {
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

// Create provides a mock function with given fields: ctx, reference, referenceID, hardLimits, used
func (_m *Controller) Create(ctx context.Context, reference string, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error) {
	_va := make([]interface{}, len(used))
	for _i := range used {
		_va[_i] = used[_i]
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
		return rf(ctx, reference, referenceID, hardLimits, used...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, types.ResourceList, ...types.ResourceList) int64); ok {
		r0 = rf(ctx, reference, referenceID, hardLimits, used...)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, types.ResourceList, ...types.ResourceList) error); ok {
		r1 = rf(ctx, reference, referenceID, hardLimits, used...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *Controller) Delete(ctx context.Context, id int64) error {
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

// Get provides a mock function with given fields: ctx, id, options
func (_m *Controller) Get(ctx context.Context, id int64, options ...quota.Option) (*models.Quota, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.Quota
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...quota.Option) (*models.Quota, error)); ok {
		return rf(ctx, id, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...quota.Option) *models.Quota); ok {
		r0 = rf(ctx, id, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Quota)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, ...quota.Option) error); ok {
		r1 = rf(ctx, id, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByRef provides a mock function with given fields: ctx, reference, referenceID, options
func (_m *Controller) GetByRef(ctx context.Context, reference string, referenceID string, options ...quota.Option) (*models.Quota, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, reference, referenceID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetByRef")
	}

	var r0 *models.Quota
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...quota.Option) (*models.Quota, error)); ok {
		return rf(ctx, reference, referenceID, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...quota.Option) *models.Quota); ok {
		r0 = rf(ctx, reference, referenceID, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Quota)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...quota.Option) error); ok {
		r1 = rf(ctx, reference, referenceID, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsEnabled provides a mock function with given fields: ctx, reference, referenceID
func (_m *Controller) IsEnabled(ctx context.Context, reference string, referenceID string) (bool, error) {
	ret := _m.Called(ctx, reference, referenceID)

	if len(ret) == 0 {
		panic("no return value specified for IsEnabled")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, reference, referenceID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, reference, referenceID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, reference, referenceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, query, options
func (_m *Controller) List(ctx context.Context, query *q.Query, options ...quota.Option) ([]*models.Quota, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*models.Quota
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query, ...quota.Option) ([]*models.Quota, error)); ok {
		return rf(ctx, query, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query, ...quota.Option) []*models.Quota); ok {
		r0 = rf(ctx, query, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Quota)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *q.Query, ...quota.Option) error); ok {
		r1 = rf(ctx, query, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Refresh provides a mock function with given fields: ctx, reference, referenceID, options
func (_m *Controller) Refresh(ctx context.Context, reference string, referenceID string, options ...quota.Option) error {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, reference, referenceID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Refresh")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...quota.Option) error); ok {
		r0 = rf(ctx, reference, referenceID, options...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Request provides a mock function with given fields: ctx, reference, referenceID, resources, f
func (_m *Controller) Request(ctx context.Context, reference string, referenceID string, resources types.ResourceList, f func() error) error {
	ret := _m.Called(ctx, reference, referenceID, resources, f)

	if len(ret) == 0 {
		panic("no return value specified for Request")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, types.ResourceList, func() error) error); ok {
		r0 = rf(ctx, reference, referenceID, resources, f)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *Controller) Update(ctx context.Context, _a1 *models.Quota) error {
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

// NewController creates a new instance of Controller. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewController(t interface {
	mock.TestingT
	Cleanup(func())
}) *Controller {
	mock := &Controller{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
