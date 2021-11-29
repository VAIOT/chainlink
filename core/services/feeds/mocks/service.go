// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	context "context"

	feeds "github.com/smartcontractkit/chainlink/core/services/feeds"
	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// ApproveJobProposal provides a mock function with given fields: ctx, id
func (_m *Service) ApproveJobProposal(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CancelJobProposal provides a mock function with given fields: ctx, id
func (_m *Service) CancelJobProposal(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *Service) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CountManagers provides a mock function with given fields:
func (_m *Service) CountManagers() (int64, error) {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateJobProposal provides a mock function with given fields: jp
func (_m *Service) CreateJobProposal(jp *feeds.JobProposal) (int64, error) {
	ret := _m.Called(jp)

	var r0 int64
	if rf, ok := ret.Get(0).(func(*feeds.JobProposal) int64); ok {
		r0 = rf(jp)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*feeds.JobProposal) error); ok {
		r1 = rf(jp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetJobProposal provides a mock function with given fields: id
func (_m *Service) GetJobProposal(id int64) (*feeds.JobProposal, error) {
	ret := _m.Called(id)

	var r0 *feeds.JobProposal
	if rf, ok := ret.Get(0).(func(int64) *feeds.JobProposal); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*feeds.JobProposal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetJobProposalsByManagersIDs provides a mock function with given fields: ids
func (_m *Service) GetJobProposalsByManagersIDs(ids []int64) ([]feeds.JobProposal, error) {
	ret := _m.Called(ids)

	var r0 []feeds.JobProposal
	if rf, ok := ret.Get(0).(func([]int64) []feeds.JobProposal); ok {
		r0 = rf(ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]feeds.JobProposal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int64) error); ok {
		r1 = rf(ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetManager provides a mock function with given fields: id
func (_m *Service) GetManager(id int64) (*feeds.FeedsManager, error) {
	ret := _m.Called(id)

	var r0 *feeds.FeedsManager
	if rf, ok := ret.Get(0).(func(int64) *feeds.FeedsManager); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*feeds.FeedsManager)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetManagers provides a mock function with given fields: ids
func (_m *Service) GetManagers(ids []int64) ([]feeds.FeedsManager, error) {
	ret := _m.Called(ids)

	var r0 []feeds.FeedsManager
	if rf, ok := ret.Get(0).(func([]int64) []feeds.FeedsManager); ok {
		r0 = rf(ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]feeds.FeedsManager)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int64) error); ok {
		r1 = rf(ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Healthy provides a mock function with given fields:
func (_m *Service) Healthy() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsJobManaged provides a mock function with given fields: ctx, jobID
func (_m *Service) IsJobManaged(ctx context.Context, jobID int64) (bool, error) {
	ret := _m.Called(ctx, jobID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, int64) bool); ok {
		r0 = rf(ctx, jobID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, jobID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListJobProposals provides a mock function with given fields:
func (_m *Service) ListJobProposals() ([]feeds.JobProposal, error) {
	ret := _m.Called()

	var r0 []feeds.JobProposal
	if rf, ok := ret.Get(0).(func() []feeds.JobProposal); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]feeds.JobProposal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListManagers provides a mock function with given fields:
func (_m *Service) ListManagers() ([]feeds.FeedsManager, error) {
	ret := _m.Called()

	var r0 []feeds.FeedsManager
	if rf, ok := ret.Get(0).(func() []feeds.FeedsManager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]feeds.FeedsManager)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProposeJob provides a mock function with given fields: jp
func (_m *Service) ProposeJob(jp *feeds.JobProposal) (int64, error) {
	ret := _m.Called(jp)

	var r0 int64
	if rf, ok := ret.Get(0).(func(*feeds.JobProposal) int64); ok {
		r0 = rf(jp)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*feeds.JobProposal) error); ok {
		r1 = rf(jp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ready provides a mock function with given fields:
func (_m *Service) Ready() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterManager provides a mock function with given fields: ms
func (_m *Service) RegisterManager(ms *feeds.FeedsManager) (int64, error) {
	ret := _m.Called(ms)

	var r0 int64
	if rf, ok := ret.Get(0).(func(*feeds.FeedsManager) int64); ok {
		r0 = rf(ms)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*feeds.FeedsManager) error); ok {
		r1 = rf(ms)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RejectJobProposal provides a mock function with given fields: ctx, id
func (_m *Service) RejectJobProposal(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields:
func (_m *Service) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SyncNodeInfo provides a mock function with given fields: id
func (_m *Service) SyncNodeInfo(id int64) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unsafe_SetConnectionsManager provides a mock function with given fields: _a0
func (_m *Service) Unsafe_SetConnectionsManager(_a0 feeds.ConnectionsManager) {
	_m.Called(_a0)
}

// UpdateFeedsManager provides a mock function with given fields: ctx, mgr
func (_m *Service) UpdateFeedsManager(ctx context.Context, mgr feeds.FeedsManager) error {
	ret := _m.Called(ctx, mgr)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, feeds.FeedsManager) error); ok {
		r0 = rf(ctx, mgr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateJobProposalSpec provides a mock function with given fields: ctx, id, spec
func (_m *Service) UpdateJobProposalSpec(ctx context.Context, id int64, spec string) error {
	ret := _m.Called(ctx, id, spec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) error); ok {
		r0 = rf(ctx, id, spec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
