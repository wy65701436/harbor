// Code generated by mockery v2.12.3. DO NOT EDIT.

package report

import (
	context "context"

	q "github.com/goharbor/harbor/src/lib/q"
	mock "github.com/stretchr/testify/mock"

	scan "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, r
func (_m *Manager) Create(ctx context.Context, r *scan.Report) (string, error) {
	ret := _m.Called(ctx, r)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *scan.Report) string); ok {
		r0 = rf(ctx, r)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *scan.Report) error); ok {
		r1 = rf(ctx, r)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, uuid
func (_m *Manager) Delete(ctx context.Context, uuid string) error {
	ret := _m.Called(ctx, uuid)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, uuid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByDigests provides a mock function with given fields: ctx, digests
func (_m *Manager) DeleteByDigests(ctx context.Context, digests ...string) error {
	_va := make([]interface{}, len(digests))
	for _i := range digests {
		_va[_i] = digests[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...string) error); ok {
		r0 = rf(ctx, digests...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBy provides a mock function with given fields: ctx, digest, registrationUUID, mimeTypes
func (_m *Manager) GetBy(ctx context.Context, digest string, registrationUUID string, mimeTypes []string) ([]*scan.Report, error) {
	ret := _m.Called(ctx, digest, registrationUUID, mimeTypes)

	var r0 []*scan.Report
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) []*scan.Report); ok {
		r0 = rf(ctx, digest, registrationUUID, mimeTypes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*scan.Report)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []string) error); ok {
		r1 = rf(ctx, digest, registrationUUID, mimeTypes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, query
func (_m *Manager) List(ctx context.Context, query *q.Query) ([]*scan.Report, error) {
	ret := _m.Called(ctx, query)

	var r0 []*scan.Report
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) []*scan.Report); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*scan.Report)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *q.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateReportData provides a mock function with given fields: ctx, uuid, _a2
func (_m *Manager) UpdateReportData(ctx context.Context, uuid string, _a2 string) error {
	ret := _m.Called(ctx, uuid, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, uuid, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewManagerT interface {
	mock.TestingT
	Cleanup(func())
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewManager(t NewManagerT) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
