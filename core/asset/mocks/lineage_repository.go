// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/goto/compass/core/asset"

	mock "github.com/stretchr/testify/mock"
)

// LineageRepository is an autogenerated mock type for the LineageRepository type
type LineageRepository struct {
	mock.Mock
}

type LineageRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *LineageRepository) EXPECT() *LineageRepository_Expecter {
	return &LineageRepository_Expecter{mock: &_m.Mock}
}

// GetGraph provides a mock function with given fields: ctx, urn, query
func (_m *LineageRepository) GetGraph(ctx context.Context, urn string, query asset.LineageQuery) (asset.LineageGraph, error) {
	ret := _m.Called(ctx, urn, query)

	var r0 asset.LineageGraph
	if rf, ok := ret.Get(0).(func(context.Context, string, asset.LineageQuery) asset.LineageGraph); ok {
		r0 = rf(ctx, urn, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(asset.LineageGraph)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, asset.LineageQuery) error); ok {
		r1 = rf(ctx, urn, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LineageRepository_GetGraph_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGraph'
type LineageRepository_GetGraph_Call struct {
	*mock.Call
}

// GetGraph is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
//   - query asset.LineageQuery
func (_e *LineageRepository_Expecter) GetGraph(ctx interface{}, urn interface{}, query interface{}) *LineageRepository_GetGraph_Call {
	return &LineageRepository_GetGraph_Call{Call: _e.mock.On("GetGraph", ctx, urn, query)}
}

func (_c *LineageRepository_GetGraph_Call) Run(run func(ctx context.Context, urn string, query asset.LineageQuery)) *LineageRepository_GetGraph_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(asset.LineageQuery))
	})
	return _c
}

func (_c *LineageRepository_GetGraph_Call) Return(_a0 asset.LineageGraph, _a1 error) *LineageRepository_GetGraph_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Upsert provides a mock function with given fields: ctx, urn, upstreams, downstreams
func (_m *LineageRepository) Upsert(ctx context.Context, urn string, upstreams []string, downstreams []string) error {
	ret := _m.Called(ctx, urn, upstreams, downstreams)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, []string) error); ok {
		r0 = rf(ctx, urn, upstreams, downstreams)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LineageRepository_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type LineageRepository_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
//   - upstreams []string
//   - downstreams []string
func (_e *LineageRepository_Expecter) Upsert(ctx interface{}, urn interface{}, upstreams interface{}, downstreams interface{}) *LineageRepository_Upsert_Call {
	return &LineageRepository_Upsert_Call{Call: _e.mock.On("Upsert", ctx, urn, upstreams, downstreams)}
}

func (_c *LineageRepository_Upsert_Call) Run(run func(ctx context.Context, urn string, upstreams []string, downstreams []string)) *LineageRepository_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]string), args[3].([]string))
	})
	return _c
}

func (_c *LineageRepository_Upsert_Call) Return(_a0 error) *LineageRepository_Upsert_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewLineageRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewLineageRepository creates a new instance of LineageRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLineageRepository(t mockConstructorTestingTNewLineageRepository) *LineageRepository {
	mock := &LineageRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
