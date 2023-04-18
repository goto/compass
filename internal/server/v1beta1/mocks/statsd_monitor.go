// Code generated by mockery v2.25.1. DO NOT EDIT.

package mocks

import (
	statsd "github.com/goto/compass/pkg/statsd"
	mock "github.com/stretchr/testify/mock"
)

// StatsDClient is an autogenerated mock type for the StatsDClient type
type StatsDClient struct {
	mock.Mock
}

type StatsDClient_Expecter struct {
	mock *mock.Mock
}

func (_m *StatsDClient) EXPECT() *StatsDClient_Expecter {
	return &StatsDClient_Expecter{mock: &_m.Mock}
}

// Incr provides a mock function with given fields: name
func (_m *StatsDClient) Incr(name string) *statsd.Metric {
	ret := _m.Called(name)

	var r0 *statsd.Metric
	if rf, ok := ret.Get(0).(func(string) *statsd.Metric); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statsd.Metric)
		}
	}

	return r0
}

// StatsDClient_Incr_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Incr'
type StatsDClient_Incr_Call struct {
	*mock.Call
}

// Incr is a helper method to define mock.On call
//   - name string
func (_e *StatsDClient_Expecter) Incr(name interface{}) *StatsDClient_Incr_Call {
	return &StatsDClient_Incr_Call{Call: _e.mock.On("Incr", name)}
}

func (_c *StatsDClient_Incr_Call) Run(run func(name string)) *StatsDClient_Incr_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *StatsDClient_Incr_Call) Return(_a0 *statsd.Metric) *StatsDClient_Incr_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StatsDClient_Incr_Call) RunAndReturn(run func(string) *statsd.Metric) *StatsDClient_Incr_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewStatsDClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewStatsDClient creates a new instance of StatsDClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStatsDClient(t mockConstructorTestingTNewStatsDClient) *StatsDClient {
	mock := &StatsDClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
