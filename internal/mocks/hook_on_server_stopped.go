// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	akira "github.com/gsalomao/akira"
	mock "github.com/stretchr/testify/mock"
)

// MockHookOnServerStopped is an autogenerated mock type for the HookOnServerStopped type
type MockHookOnServerStopped struct {
	mock.Mock
}

type MockHookOnServerStopped_Expecter struct {
	mock *mock.Mock
}

func (_m *MockHookOnServerStopped) EXPECT() *MockHookOnServerStopped_Expecter {
	return &MockHookOnServerStopped_Expecter{mock: &_m.Mock}
}

// Name provides a mock function with given fields:
func (_m *MockHookOnServerStopped) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockHookOnServerStopped_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockHookOnServerStopped_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockHookOnServerStopped_Expecter) Name() *MockHookOnServerStopped_Name_Call {
	return &MockHookOnServerStopped_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockHookOnServerStopped_Name_Call) Run(run func()) *MockHookOnServerStopped_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockHookOnServerStopped_Name_Call) Return(_a0 string) *MockHookOnServerStopped_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockHookOnServerStopped_Name_Call) RunAndReturn(run func() string) *MockHookOnServerStopped_Name_Call {
	_c.Call.Return(run)
	return _c
}

// OnServerStopped provides a mock function with given fields: s
func (_m *MockHookOnServerStopped) OnServerStopped(s *akira.Server) {
	_m.Called(s)
}

// MockHookOnServerStopped_OnServerStopped_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnServerStopped'
type MockHookOnServerStopped_OnServerStopped_Call struct {
	*mock.Call
}

// OnServerStopped is a helper method to define mock.On call
//   - s *akira.Server
func (_e *MockHookOnServerStopped_Expecter) OnServerStopped(s interface{}) *MockHookOnServerStopped_OnServerStopped_Call {
	return &MockHookOnServerStopped_OnServerStopped_Call{Call: _e.mock.On("OnServerStopped", s)}
}

func (_c *MockHookOnServerStopped_OnServerStopped_Call) Run(run func(s *akira.Server)) *MockHookOnServerStopped_OnServerStopped_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*akira.Server))
	})
	return _c
}

func (_c *MockHookOnServerStopped_OnServerStopped_Call) Return() *MockHookOnServerStopped_OnServerStopped_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockHookOnServerStopped_OnServerStopped_Call) RunAndReturn(run func(*akira.Server)) *MockHookOnServerStopped_OnServerStopped_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockHookOnServerStopped creates a new instance of MockHookOnServerStopped. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockHookOnServerStopped(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHookOnServerStopped {
	mock := &MockHookOnServerStopped{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
