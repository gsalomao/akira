// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	akira "github.com/gsalomao/akira"
	mock "github.com/stretchr/testify/mock"
)

// MockHookOnPacketReceiveError is an autogenerated mock type for the HookOnPacketReceiveError type
type MockHookOnPacketReceiveError struct {
	mock.Mock
}

type MockHookOnPacketReceiveError_Expecter struct {
	mock *mock.Mock
}

func (_m *MockHookOnPacketReceiveError) EXPECT() *MockHookOnPacketReceiveError_Expecter {
	return &MockHookOnPacketReceiveError_Expecter{mock: &_m.Mock}
}

// Name provides a mock function with given fields:
func (_m *MockHookOnPacketReceiveError) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockHookOnPacketReceiveError_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockHookOnPacketReceiveError_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockHookOnPacketReceiveError_Expecter) Name() *MockHookOnPacketReceiveError_Name_Call {
	return &MockHookOnPacketReceiveError_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockHookOnPacketReceiveError_Name_Call) Run(run func()) *MockHookOnPacketReceiveError_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockHookOnPacketReceiveError_Name_Call) Return(_a0 string) *MockHookOnPacketReceiveError_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockHookOnPacketReceiveError_Name_Call) RunAndReturn(run func() string) *MockHookOnPacketReceiveError_Name_Call {
	_c.Call.Return(run)
	return _c
}

// OnPacketReceiveError provides a mock function with given fields: c, err
func (_m *MockHookOnPacketReceiveError) OnPacketReceiveError(c *akira.Client, err error) error {
	ret := _m.Called(c, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(*akira.Client, error) error); ok {
		r0 = rf(c, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockHookOnPacketReceiveError_OnPacketReceiveError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnPacketReceiveError'
type MockHookOnPacketReceiveError_OnPacketReceiveError_Call struct {
	*mock.Call
}

// OnPacketReceiveError is a helper method to define mock.On call
//   - c *akira.Client
//   - err error
func (_e *MockHookOnPacketReceiveError_Expecter) OnPacketReceiveError(c interface{}, err interface{}) *MockHookOnPacketReceiveError_OnPacketReceiveError_Call {
	return &MockHookOnPacketReceiveError_OnPacketReceiveError_Call{Call: _e.mock.On("OnPacketReceiveError", c, err)}
}

func (_c *MockHookOnPacketReceiveError_OnPacketReceiveError_Call) Run(run func(c *akira.Client, err error)) *MockHookOnPacketReceiveError_OnPacketReceiveError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*akira.Client), args[1].(error))
	})
	return _c
}

func (_c *MockHookOnPacketReceiveError_OnPacketReceiveError_Call) Return(_a0 error) *MockHookOnPacketReceiveError_OnPacketReceiveError_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockHookOnPacketReceiveError_OnPacketReceiveError_Call) RunAndReturn(run func(*akira.Client, error) error) *MockHookOnPacketReceiveError_OnPacketReceiveError_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockHookOnPacketReceiveError creates a new instance of MockHookOnPacketReceiveError. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockHookOnPacketReceiveError(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHookOnPacketReceiveError {
	mock := &MockHookOnPacketReceiveError{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
