// Code generated by mockery v2.1.0. DO NOT EDIT.

package policy

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	git "gopkg.in/src-d/go-git.v4"
)

// MockGitService is an autogenerated mock type for the GitService type
type MockGitService struct {
	mock.Mock
}

// Clone provides a mock function with given fields: _a0, _a1
func (_m *MockGitService) Clone(_a0 context.Context, _a1 string) (*git.Repository, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *git.Repository
	if rf, ok := ret.Get(0).(func(context.Context, string) *git.Repository); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*git.Repository)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Commit provides a mock function with given fields: ctx, rootPath, changesPath, msg
func (_m *MockGitService) Commit(ctx context.Context, rootPath string, changesPath string, msg string) error {
	ret := _m.Called(ctx, rootPath, changesPath, msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, rootPath, changesPath, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MasterPath provides a mock function with given fields:
func (_m *MockGitService) MasterPath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
