// Code generated by mockery v2.25.1. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/goto/compass/core/asset"

	mock "github.com/stretchr/testify/mock"
)

// Worker is an autogenerated mock type for the Worker type
type Worker struct {
	mock.Mock
}

type Worker_Expecter struct {
	mock *mock.Mock
}

func (_m *Worker) EXPECT() *Worker_Expecter {
	return &Worker_Expecter{mock: &_m.Mock}
}

// EnqueueDeleteAssetJob provides a mock function with given fields: ctx, urn
func (_m *Worker) EnqueueDeleteAssetJob(ctx context.Context, urn string) error {
	ret := _m.Called(ctx, urn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, urn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Worker_EnqueueDeleteAssetJob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnqueueDeleteAssetJob'
type Worker_EnqueueDeleteAssetJob_Call struct {
	*mock.Call
}

// EnqueueDeleteAssetJob is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
func (_e *Worker_Expecter) EnqueueDeleteAssetJob(ctx interface{}, urn interface{}) *Worker_EnqueueDeleteAssetJob_Call {
	return &Worker_EnqueueDeleteAssetJob_Call{Call: _e.mock.On("EnqueueDeleteAssetJob", ctx, urn)}
}

func (_c *Worker_EnqueueDeleteAssetJob_Call) Run(run func(ctx context.Context, urn string)) *Worker_EnqueueDeleteAssetJob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Worker_EnqueueDeleteAssetJob_Call) Return(_a0 error) *Worker_EnqueueDeleteAssetJob_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Worker_EnqueueDeleteAssetJob_Call) RunAndReturn(run func(context.Context, string) error) *Worker_EnqueueDeleteAssetJob_Call {
	_c.Call.Return(run)
	return _c
}

// EnqueueIndexAssetJob provides a mock function with given fields: ctx, ast
func (_m *Worker) EnqueueIndexAssetJob(ctx context.Context, ast asset.Asset) error {
	ret := _m.Called(ctx, ast)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Asset) error); ok {
		r0 = rf(ctx, ast)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Worker_EnqueueIndexAssetJob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnqueueIndexAssetJob'
type Worker_EnqueueIndexAssetJob_Call struct {
	*mock.Call
}

// EnqueueIndexAssetJob is a helper method to define mock.On call
//   - ctx context.Context
//   - ast asset.Asset
func (_e *Worker_Expecter) EnqueueIndexAssetJob(ctx interface{}, ast interface{}) *Worker_EnqueueIndexAssetJob_Call {
	return &Worker_EnqueueIndexAssetJob_Call{Call: _e.mock.On("EnqueueIndexAssetJob", ctx, ast)}
}

func (_c *Worker_EnqueueIndexAssetJob_Call) Run(run func(ctx context.Context, ast asset.Asset)) *Worker_EnqueueIndexAssetJob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Asset))
	})
	return _c
}

func (_c *Worker_EnqueueIndexAssetJob_Call) Return(_a0 error) *Worker_EnqueueIndexAssetJob_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Worker_EnqueueIndexAssetJob_Call) RunAndReturn(run func(context.Context, asset.Asset) error) *Worker_EnqueueIndexAssetJob_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewWorker interface {
	mock.TestingT
	Cleanup(func())
}

// NewWorker creates a new instance of Worker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWorker(t mockConstructorTestingTNewWorker) *Worker {
	mock := &Worker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}