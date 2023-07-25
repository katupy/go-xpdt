// Code generated by mockery v2.32.0. DO NOT EDIT.

package env

import mock "github.com/stretchr/testify/mock"

// MockCommandMethods is an autogenerated mock type for the CommandMethods type
type MockCommandMethods struct {
	mock.Mock
}

type MockCommandMethods_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCommandMethods) EXPECT() *MockCommandMethods_Expecter {
	return &MockCommandMethods_Expecter{mock: &_m.Mock}
}

// Add provides a mock function with given fields: cmd
func (_m *MockCommandMethods) Add(cmd *Command) error {
	ret := _m.Called(cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(*Command) error); ok {
		r0 = rf(cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCommandMethods_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type MockCommandMethods_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - cmd *Command
func (_e *MockCommandMethods_Expecter) Add(cmd interface{}) *MockCommandMethods_Add_Call {
	return &MockCommandMethods_Add_Call{Call: _e.mock.On("Add", cmd)}
}

func (_c *MockCommandMethods_Add_Call) Run(run func(cmd *Command)) *MockCommandMethods_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*Command))
	})
	return _c
}

func (_c *MockCommandMethods_Add_Call) Return(_a0 error) *MockCommandMethods_Add_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCommandMethods_Add_Call) RunAndReturn(run func(*Command) error) *MockCommandMethods_Add_Call {
	_c.Call.Return(run)
	return _c
}

// Del provides a mock function with given fields: cmd
func (_m *MockCommandMethods) Del(cmd *Command) error {
	ret := _m.Called(cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(*Command) error); ok {
		r0 = rf(cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCommandMethods_Del_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Del'
type MockCommandMethods_Del_Call struct {
	*mock.Call
}

// Del is a helper method to define mock.On call
//   - cmd *Command
func (_e *MockCommandMethods_Expecter) Del(cmd interface{}) *MockCommandMethods_Del_Call {
	return &MockCommandMethods_Del_Call{Call: _e.mock.On("Del", cmd)}
}

func (_c *MockCommandMethods_Del_Call) Run(run func(cmd *Command)) *MockCommandMethods_Del_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*Command))
	})
	return _c
}

func (_c *MockCommandMethods_Del_Call) Return(_a0 error) *MockCommandMethods_Del_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCommandMethods_Del_Call) RunAndReturn(run func(*Command) error) *MockCommandMethods_Del_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: cmd
func (_m *MockCommandMethods) Set(cmd *Command) error {
	ret := _m.Called(cmd)

	var r0 error
	if rf, ok := ret.Get(0).(func(*Command) error); ok {
		r0 = rf(cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCommandMethods_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type MockCommandMethods_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - cmd *Command
func (_e *MockCommandMethods_Expecter) Set(cmd interface{}) *MockCommandMethods_Set_Call {
	return &MockCommandMethods_Set_Call{Call: _e.mock.On("Set", cmd)}
}

func (_c *MockCommandMethods_Set_Call) Run(run func(cmd *Command)) *MockCommandMethods_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*Command))
	})
	return _c
}

func (_c *MockCommandMethods_Set_Call) Return(_a0 error) *MockCommandMethods_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCommandMethods_Set_Call) RunAndReturn(run func(*Command) error) *MockCommandMethods_Set_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCommandMethods creates a new instance of MockCommandMethods. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCommandMethods(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCommandMethods {
	mock := &MockCommandMethods{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
