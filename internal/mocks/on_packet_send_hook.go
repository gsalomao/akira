// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	akira "github.com/gsalomao/akira"
	mock "github.com/stretchr/testify/mock"
)

// MockOnPacketSendHook is an autogenerated mock type for the OnPacketSendHook type
type MockOnPacketSendHook struct {
	mock.Mock
}

type MockOnPacketSendHook_Expecter struct {
	mock *mock.Mock
}

func (_m *MockOnPacketSendHook) EXPECT() *MockOnPacketSendHook_Expecter {
	return &MockOnPacketSendHook_Expecter{mock: &_m.Mock}
}

// Name provides a mock function with given fields:
func (_m *MockOnPacketSendHook) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockOnPacketSendHook_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockOnPacketSendHook_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockOnPacketSendHook_Expecter) Name() *MockOnPacketSendHook_Name_Call {
	return &MockOnPacketSendHook_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockOnPacketSendHook_Name_Call) Run(run func()) *MockOnPacketSendHook_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockOnPacketSendHook_Name_Call) Return(_a0 string) *MockOnPacketSendHook_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockOnPacketSendHook_Name_Call) RunAndReturn(run func() string) *MockOnPacketSendHook_Name_Call {
	_c.Call.Return(run)
	return _c
}

// OnPacketSend provides a mock function with given fields: c, p
func (_m *MockOnPacketSendHook) OnPacketSend(c *akira.Client, p akira.Packet) error {
	ret := _m.Called(c, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(*akira.Client, akira.Packet) error); ok {
		r0 = rf(c, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockOnPacketSendHook_OnPacketSend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnPacketSend'
type MockOnPacketSendHook_OnPacketSend_Call struct {
	*mock.Call
}

// OnPacketSend is a helper method to define mock.On call
//   - c *akira.Client
//   - p akira.Packet
func (_e *MockOnPacketSendHook_Expecter) OnPacketSend(c interface{}, p interface{}) *MockOnPacketSendHook_OnPacketSend_Call {
	return &MockOnPacketSendHook_OnPacketSend_Call{Call: _e.mock.On("OnPacketSend", c, p)}
}

func (_c *MockOnPacketSendHook_OnPacketSend_Call) Run(run func(c *akira.Client, p akira.Packet)) *MockOnPacketSendHook_OnPacketSend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*akira.Client), args[1].(akira.Packet))
	})
	return _c
}

func (_c *MockOnPacketSendHook_OnPacketSend_Call) Return(_a0 error) *MockOnPacketSendHook_OnPacketSend_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockOnPacketSendHook_OnPacketSend_Call) RunAndReturn(run func(*akira.Client, akira.Packet) error) *MockOnPacketSendHook_OnPacketSend_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockOnPacketSendHook creates a new instance of MockOnPacketSendHook. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockOnPacketSendHook(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOnPacketSendHook {
	mock := &MockOnPacketSendHook{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}