// Code generated by mockery v2.1.0. DO NOT EDIT.

package label

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/goharbor/harbor/src/pkg/label/model"

	q "github.com/goharbor/harbor/src/lib/q"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// AddTo provides a mock function with given fields: ctx, labelID, artifactID
func (_m *Manager) AddTo(ctx context.Context, labelID int64, artifactID int64) error {
	ret := _m.Called(ctx, labelID, artifactID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(ctx, labelID, artifactID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Count provides a mock function with given fields: ctx, query
func (_m *Manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	ret := _m.Called(ctx, query)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) int64); ok {
		r0 = rf(ctx, query)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *q.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: ctx, _a1
func (_m *Manager) Create(ctx context.Context, _a1 *model.Label) (int64, error) {
	ret := _m.Called(ctx, _a1)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, *model.Label) int64); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.Label) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *Manager) Delete(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *Manager) Get(ctx context.Context, id int64) (*model.Label, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Label
	if rf, ok := ret.Get(0).(func(context.Context, int64) *model.Label); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, query
func (_m *Manager) List(ctx context.Context, query *q.Query) ([]*model.Label, error) {
	ret := _m.Called(ctx, query)

	var r0 []*model.Label
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) []*model.Label); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Label)
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

// ListByArtifact provides a mock function with given fields: ctx, artifactID
func (_m *Manager) ListByArtifact(ctx context.Context, artifactID int64) ([]*model.Label, error) {
	ret := _m.Called(ctx, artifactID)

	var r0 []*model.Label
	if rf, ok := ret.Get(0).(func(context.Context, int64) []*model.Label); ok {
		r0 = rf(ctx, artifactID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, artifactID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveAllFrom provides a mock function with given fields: ctx, artifactID
func (_m *Manager) RemoveAllFrom(ctx context.Context, artifactID int64) error {
	ret := _m.Called(ctx, artifactID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, artifactID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveFrom provides a mock function with given fields: ctx, labelID, artifactID
func (_m *Manager) RemoveFrom(ctx context.Context, labelID int64, artifactID int64) error {
	ret := _m.Called(ctx, labelID, artifactID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(ctx, labelID, artifactID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveFromAllArtifacts provides a mock function with given fields: ctx, labelID
func (_m *Manager) RemoveFromAllArtifacts(ctx context.Context, labelID int64) error {
	ret := _m.Called(ctx, labelID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, labelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *Manager) Update(ctx context.Context, _a1 *model.Label) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Label) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
