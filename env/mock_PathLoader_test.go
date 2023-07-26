// Code generated by mockery v2.32.0. DO NOT EDIT.

package env

import mock "github.com/stretchr/testify/mock"

// MockPathLoader is an autogenerated mock type for the PathLoader type
type MockPathLoader struct {
	mock.Mock
}

type MockPathLoader_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPathLoader) EXPECT() *MockPathLoader_Expecter {
	return &MockPathLoader_Expecter{mock: &_m.Mock}
}

// Load provides a mock function with given fields: envVar
func (_m *MockPathLoader) Load(envVar *environVar) error {
	ret := _m.Called(envVar)

	var r0 error
	if rf, ok := ret.Get(0).(func(*environVar) error); ok {
		r0 = rf(envVar)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockPathLoader_Load_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Load'
type MockPathLoader_Load_Call struct {
	*mock.Call
}

// Load is a helper method to define mock.On call
//   - envVar *environVar
func (_e *MockPathLoader_Expecter) Load(envVar interface{}) *MockPathLoader_Load_Call {
	return &MockPathLoader_Load_Call{Call: _e.mock.On("Load", envVar)}
}

func (_c *MockPathLoader_Load_Call) Run(run func(envVar *environVar)) *MockPathLoader_Load_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*environVar))
	})
	return _c
}

func (_c *MockPathLoader_Load_Call) Return(_a0 error) *MockPathLoader_Load_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockPathLoader_Load_Call) RunAndReturn(run func(*environVar) error) *MockPathLoader_Load_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockPathLoader creates a new instance of MockPathLoader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPathLoader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPathLoader {
	mock := &MockPathLoader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
