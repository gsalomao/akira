// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	akira "github.com/gsalomao/akira"
	mock "github.com/stretchr/testify/mock"
)

// MockOnPacketReceivedHook is an autogenerated mock type for the OnPacketReceivedHook type
type MockOnPacketReceivedHook struct {
	mock.Mock
}

type MockOnPacketReceivedHook_Expecter struct {
	mock *mock.Mock
}

func (_m *MockOnPacketReceivedHook) EXPECT() *MockOnPacketReceivedHook_Expecter {
	return &MockOnPacketReceivedHook_Expecter{mock: &_m.Mock}
}

// Name provides a mock function with given fields:
func (_m *MockOnPacketReceivedHook) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockOnPacketReceivedHook_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockOnPacketReceivedHook_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockOnPacketReceivedHook_Expecter) Name() *MockOnPacketReceivedHook_Name_Call {
	return &MockOnPacketReceivedHook_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockOnPacketReceivedHook_Name_Call) Run(run func()) *MockOnPacketReceivedHook_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockOnPacketReceivedHook_Name_Call) Return(_a0 string) *MockOnPacketReceivedHook_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockOnPacketReceivedHook_Name_Call) RunAndReturn(run func() string) *MockOnPacketReceivedHook_Name_Call {
	_c.Call.Return(run)
	return _c
}

// OnPacketReceived provides a mock function with given fields: c, p
func (_m *MockOnPacketReceivedHook) OnPacketReceived(c *akira.Client, p akira.Packet) error {
	ret := _m.Called(c, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(*akira.Client, akira.Packet) error); ok {
		r0 = rf(c, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockOnPacketReceivedHook_OnPacketReceived_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnPacketReceived'
type MockOnPacketReceivedHook_OnPacketReceived_Call struct {
	*mock.Call
}

// OnPacketReceived is a helper method to define mock.On call
//   - c *akira.Client
//   - p akira.Packet
func (_e *MockOnPacketReceivedHook_Expecter) OnPacketReceived(c interface{}, p interface{}) *MockOnPacketReceivedHook_OnPacketReceived_Call {
	return &MockOnPacketReceivedHook_OnPacketReceived_Call{Call: _e.mock.On("OnPacketReceived", c, p)}
}

func (_c *MockOnPacketReceivedHook_OnPacketReceived_Call) Run(run func(c *akira.Client, p akira.Packet)) *MockOnPacketReceivedHook_OnPacketReceived_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*akira.Client), args[1].(akira.Packet))
	})
	return _c
}

func (_c *MockOnPacketReceivedHook_OnPacketReceived_Call) Return(_a0 error) *MockOnPacketReceivedHook_OnPacketReceived_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockOnPacketReceivedHook_OnPacketReceived_Call) RunAndReturn(run func(*akira.Client, akira.Packet) error) *MockOnPacketReceivedHook_OnPacketReceived_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockOnPacketReceivedHook creates a new instance of MockOnPacketReceivedHook. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockOnPacketReceivedHook(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOnPacketReceivedHook {
	mock := &MockOnPacketReceivedHook{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}