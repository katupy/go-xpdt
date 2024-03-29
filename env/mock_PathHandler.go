// Code generated by mockery v2.33.1. DO NOT EDIT.

package env

import mock "github.com/stretchr/testify/mock"

// MockPathHandler is an autogenerated mock type for the PathHandler type
type MockPathHandler struct {
	mock.Mock
}

type MockPathHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPathHandler) EXPECT() *MockPathHandler_Expecter {
	return &MockPathHandler_Expecter{mock: &_m.Mock}
}

// Add provides a mock function with given fields: envVar, value, position
func (_m *MockPathHandler) Add(envVar *environVar, value string, position int) error {
	ret := _m.Called(envVar, value, position)

	var r0 error
	if rf, ok := ret.Get(0).(func(*environVar, string, int) error); ok {
		r0 = rf(envVar, value, position)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPathHandler_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type MockPathHandler_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - envVar *environVar
//   - value string
//   - position int
func (_e *MockPathHandler_Expecter) Add(envVar interface{}, value interface{}, position interface{}) *MockPathHandler_Add_Call {
	return &MockPathHandler_Add_Call{Call: _e.mock.On("Add", envVar, value, position)}
}

func (_c *MockPathHandler_Add_Call) Run(run func(envVar *environVar, value string, position int)) *MockPathHandler_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*environVar), args[1].(string), args[2].(int))
	})
	return _c
}

func (_c *MockPathHandler_Add_Call) Return(_a0 error) *MockPathHandler_Add_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPathHandler_Add_Call) RunAndReturn(run func(*environVar, string, int) error) *MockPathHandler_Add_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockPathHandler creates a new instance of MockPathHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPathHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPathHandler {
	mock := &MockPathHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
