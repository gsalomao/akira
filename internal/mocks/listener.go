// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	akira "github.com/gsalomao/akira"
	mock "github.com/stretchr/testify/mock"
)

// MockListener is an autogenerated mock type for the Listener type
type MockListener struct {
	mock.Mock
}

type MockListener_Expecter struct {
	mock *mock.Mock
}

func (_m *MockListener) EXPECT() *MockListener_Expecter {
	return &MockListener_Expecter{mock: &_m.Mock}
}

// Listen provides a mock function with given fields: _a0
func (_m *MockListener) Listen(_a0 akira.OnConnectionFunc) (<-chan bool, error) {
	ret := _m.Called(_a0)

	var r0 <-chan bool
	var r1 error
	if rf, ok := ret.Get(0).(func(akira.OnConnectionFunc) (<-chan bool, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(akira.OnConnectionFunc) <-chan bool); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan bool)
		}
	}

	if rf, ok := ret.Get(1).(func(akira.OnConnectionFunc) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockListener_Listen_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Listen'
type MockListener_Listen_Call struct {
	*mock.Call
}

// Listen is a helper method to define mock.On call
//   - _a0 akira.OnConnectionFunc
func (_e *MockListener_Expecter) Listen(_a0 interface{}) *MockListener_Listen_Call {
	return &MockListener_Listen_Call{Call: _e.mock.On("Listen", _a0)}
}

func (_c *MockListener_Listen_Call) Run(run func(_a0 akira.OnConnectionFunc)) *MockListener_Listen_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(akira.OnConnectionFunc))
	})
	return _c
}

func (_c *MockListener_Listen_Call) Return(_a0 <-chan bool, _a1 error) *MockListener_Listen_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockListener_Listen_Call) RunAndReturn(run func(akira.OnConnectionFunc) (<-chan bool, error)) *MockListener_Listen_Call {
	_c.Call.Return(run)
	return _c
}

// Listening provides a mock function with given fields:
func (_m *MockListener) Listening() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockListener_Listening_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Listening'
type MockListener_Listening_Call struct {
	*mock.Call
}

// Listening is a helper method to define mock.On call
func (_e *MockListener_Expecter) Listening() *MockListener_Listening_Call {
	return &MockListener_Listening_Call{Call: _e.mock.On("Listening")}
}

func (_c *MockListener_Listening_Call) Run(run func()) *MockListener_Listening_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockListener_Listening_Call) Return(_a0 bool) *MockListener_Listening_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockListener_Listening_Call) RunAndReturn(run func() bool) *MockListener_Listening_Call {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with given fields:
func (_m *MockListener) Stop() {
	_m.Called()
}

// MockListener_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type MockListener_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *MockListener_Expecter) Stop() *MockListener_Stop_Call {
	return &MockListener_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *MockListener_Stop_Call) Run(run func()) *MockListener_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockListener_Stop_Call) Return() *MockListener_Stop_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockListener_Stop_Call) RunAndReturn(run func()) *MockListener_Stop_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockListener creates a new instance of MockListener. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockListener(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockListener {
	mock := &MockListener{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
