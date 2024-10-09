// Code generated by mockery v2.46.2. DO NOT EDIT.

package blob

import (
	context "context"

	models "github.com/goharbor/harbor/src/pkg/blob/models"
	mock "github.com/stretchr/testify/mock"

	q "github.com/goharbor/harbor/src/lib/q"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// AssociateWithArtifact provides a mock function with given fields: ctx, blobDigest, artifactDigest
func (_m *Manager) AssociateWithArtifact(ctx context.Context, blobDigest string, artifactDigest string) (int64, error) {
	ret := _m.Called(ctx, blobDigest, artifactDigest)

	if len(ret) == 0 {
		panic("no return value specified for AssociateWithArtifact")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (int64, error)); ok {
		return rf(ctx, blobDigest, artifactDigest)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) int64); ok {
		r0 = rf(ctx, blobDigest, artifactDigest)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, blobDigest, artifactDigest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssociateWithProject provides a mock function with given fields: ctx, blobID, projectID
func (_m *Manager) AssociateWithProject(ctx context.Context, blobID int64, projectID int64) (int64, error) {
	ret := _m.Called(ctx, blobID, projectID)

	if len(ret) == 0 {
		panic("no return value specified for AssociateWithProject")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) (int64, error)); ok {
		return rf(ctx, blobID, projectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) int64); ok {
		r0 = rf(ctx, blobID, projectID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, int64) error); ok {
		r1 = rf(ctx, blobID, projectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CalculateTotalSize provides a mock function with given fields: ctx, excludeForeignLayer
func (_m *Manager) CalculateTotalSize(ctx context.Context, excludeForeignLayer bool) (int64, error) {
	ret := _m.Called(ctx, excludeForeignLayer)

	if len(ret) == 0 {
		panic("no return value specified for CalculateTotalSize")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, bool) (int64, error)); ok {
		return rf(ctx, excludeForeignLayer)
	}
	if rf, ok := ret.Get(0).(func(context.Context, bool) int64); ok {
		r0 = rf(ctx, excludeForeignLayer)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, bool) error); ok {
		r1 = rf(ctx, excludeForeignLayer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CalculateTotalSizeByProject provides a mock function with given fields: ctx, projectID, excludeForeignLayer
func (_m *Manager) CalculateTotalSizeByProject(ctx context.Context, projectID int64, excludeForeignLayer bool) (int64, error) {
	ret := _m.Called(ctx, projectID, excludeForeignLayer)

	if len(ret) == 0 {
		panic("no return value specified for CalculateTotalSizeByProject")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, bool) (int64, error)); ok {
		return rf(ctx, projectID, excludeForeignLayer)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, bool) int64); ok {
		r0 = rf(ctx, projectID, excludeForeignLayer)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, bool) error); ok {
		r1 = rf(ctx, projectID, excludeForeignLayer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CleanupAssociationsForArtifact provides a mock function with given fields: ctx, artifactDigest
func (_m *Manager) CleanupAssociationsForArtifact(ctx context.Context, artifactDigest string) error {
	ret := _m.Called(ctx, artifactDigest)

	if len(ret) == 0 {
		panic("no return value specified for CleanupAssociationsForArtifact")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, artifactDigest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CleanupAssociationsForProject provides a mock function with given fields: ctx, projectID, blobs
func (_m *Manager) CleanupAssociationsForProject(ctx context.Context, projectID int64, blobs []*models.Blob) error {
	ret := _m.Called(ctx, projectID, blobs)

	if len(ret) == 0 {
		panic("no return value specified for CleanupAssociationsForProject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, []*models.Blob) error); ok {
		r0 = rf(ctx, projectID, blobs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: ctx, digest, contentType, size
func (_m *Manager) Create(ctx context.Context, digest string, contentType string, size int64) (int64, error) {
	ret := _m.Called(ctx, digest, contentType, size)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64) (int64, error)); ok {
		return rf(ctx, digest, contentType, size)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64) int64); ok {
		r0 = rf(ctx, digest, contentType, size)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int64) error); ok {
		r1 = rf(ctx, digest, contentType, size)
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

// FindBlobsShouldUnassociatedWithProject provides a mock function with given fields: ctx, projectID, blobs
func (_m *Manager) FindBlobsShouldUnassociatedWithProject(ctx context.Context, projectID int64, blobs []*models.Blob) ([]*models.Blob, error) {
	ret := _m.Called(ctx, projectID, blobs)

	if len(ret) == 0 {
		panic("no return value specified for FindBlobsShouldUnassociatedWithProject")
	}

	var r0 []*models.Blob
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, []*models.Blob) ([]*models.Blob, error)); ok {
		return rf(ctx, projectID, blobs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, []*models.Blob) []*models.Blob); ok {
		r0 = rf(ctx, projectID, blobs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Blob)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, []*models.Blob) error); ok {
		r1 = rf(ctx, projectID, blobs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, digest
func (_m *Manager) Get(ctx context.Context, digest string) (*models.Blob, error) {
	ret := _m.Called(ctx, digest)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.Blob
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*models.Blob, error)); ok {
		return rf(ctx, digest)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.Blob); ok {
		r0 = rf(ctx, digest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Blob)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, digest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByArt provides a mock function with given fields: ctx, digest
func (_m *Manager) GetByArt(ctx context.Context, digest string) ([]*models.Blob, error) {
	ret := _m.Called(ctx, digest)

	if len(ret) == 0 {
		panic("no return value specified for GetByArt")
	}

	var r0 []*models.Blob
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*models.Blob, error)); ok {
		return rf(ctx, digest)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*models.Blob); ok {
		r0 = rf(ctx, digest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Blob)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, digest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, query
func (_m *Manager) List(ctx context.Context, query *q.Query) ([]*models.Blob, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*models.Blob
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) ([]*models.Blob, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) []*models.Blob); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Blob)
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
func (_m *Manager) Update(ctx context.Context, _a1 *models.Blob) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Blob) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateBlobStatus provides a mock function with given fields: ctx, _a1
func (_m *Manager) UpdateBlobStatus(ctx context.Context, _a1 *models.Blob) (int64, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for UpdateBlobStatus")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Blob) (int64, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.Blob) int64); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.Blob) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UselessBlobs provides a mock function with given fields: ctx, timeWindowHours
func (_m *Manager) UselessBlobs(ctx context.Context, timeWindowHours int64) ([]*models.Blob, error) {
	ret := _m.Called(ctx, timeWindowHours)

	if len(ret) == 0 {
		panic("no return value specified for UselessBlobs")
	}

	var r0 []*models.Blob
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]*models.Blob, error)); ok {
		return rf(ctx, timeWindowHours)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []*models.Blob); ok {
		r0 = rf(ctx, timeWindowHours)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Blob)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, timeWindowHours)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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
