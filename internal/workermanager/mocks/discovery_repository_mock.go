// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/goto/compass/core/asset"

	mock "github.com/stretchr/testify/mock"
)

// DiscoveryRepository is an autogenerated mock type for the DiscoveryRepository type
type DiscoveryRepository struct {
	mock.Mock
}

type DiscoveryRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *DiscoveryRepository) EXPECT() *DiscoveryRepository_Expecter {
	return &DiscoveryRepository_Expecter{mock: &_m.Mock}
}

// DeleteByURN provides a mock function with given fields: ctx, assetURN
func (_m *DiscoveryRepository) DeleteByURN(ctx context.Context, assetURN string) error {
	ret := _m.Called(ctx, assetURN)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, assetURN)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DiscoveryRepository_DeleteByURN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteByURN'
type DiscoveryRepository_DeleteByURN_Call struct {
	*mock.Call
}

// DeleteByURN is a helper method to define mock.On call
//   - ctx context.Context
//   - assetURN string
func (_e *DiscoveryRepository_Expecter) DeleteByURN(ctx interface{}, assetURN interface{}) *DiscoveryRepository_DeleteByURN_Call {
	return &DiscoveryRepository_DeleteByURN_Call{Call: _e.mock.On("DeleteByURN", ctx, assetURN)}
}

func (_c *DiscoveryRepository_DeleteByURN_Call) Run(run func(ctx context.Context, assetURN string)) *DiscoveryRepository_DeleteByURN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DiscoveryRepository_DeleteByURN_Call) Return(_a0 error) *DiscoveryRepository_DeleteByURN_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DiscoveryRepository_DeleteByURN_Call) RunAndReturn(run func(context.Context, string) error) *DiscoveryRepository_DeleteByURN_Call {
	_c.Call.Return(run)
	return _c
}

// SyncAssets provides a mock function with given fields: ctx, indexName
func (_m *DiscoveryRepository) SyncAssets(ctx context.Context, indexName string) (func() error, error) {
	ret := _m.Called(ctx, indexName)

	var r0 func() error
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (func() error, error)); ok {
		return rf(ctx, indexName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) func() error); ok {
		r0 = rf(ctx, indexName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func() error)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, indexName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DiscoveryRepository_SyncAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SyncAssets'
type DiscoveryRepository_SyncAssets_Call struct {
	*mock.Call
}

// SyncAssets is a helper method to define mock.On call
//   - ctx context.Context
//   - indexName string
func (_e *DiscoveryRepository_Expecter) SyncAssets(ctx interface{}, indexName interface{}) *DiscoveryRepository_SyncAssets_Call {
	return &DiscoveryRepository_SyncAssets_Call{Call: _e.mock.On("SyncAssets", ctx, indexName)}
}

func (_c *DiscoveryRepository_SyncAssets_Call) Run(run func(ctx context.Context, indexName string)) *DiscoveryRepository_SyncAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DiscoveryRepository_SyncAssets_Call) Return(cleanup func() error, err error) *DiscoveryRepository_SyncAssets_Call {
	_c.Call.Return(cleanup, err)
	return _c
}

func (_c *DiscoveryRepository_SyncAssets_Call) RunAndReturn(run func(context.Context, string) (func() error, error)) *DiscoveryRepository_SyncAssets_Call {
	_c.Call.Return(run)
	return _c
}

// Upsert provides a mock function with given fields: _a0, _a1
func (_m *DiscoveryRepository) Upsert(_a0 context.Context, _a1 asset.Asset) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Asset) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DiscoveryRepository_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type DiscoveryRepository_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 asset.Asset
func (_e *DiscoveryRepository_Expecter) Upsert(_a0 interface{}, _a1 interface{}) *DiscoveryRepository_Upsert_Call {
	return &DiscoveryRepository_Upsert_Call{Call: _e.mock.On("Upsert", _a0, _a1)}
}

func (_c *DiscoveryRepository_Upsert_Call) Run(run func(_a0 context.Context, _a1 asset.Asset)) *DiscoveryRepository_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Asset))
	})
	return _c
}

func (_c *DiscoveryRepository_Upsert_Call) Return(_a0 error) *DiscoveryRepository_Upsert_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DiscoveryRepository_Upsert_Call) RunAndReturn(run func(context.Context, asset.Asset) error) *DiscoveryRepository_Upsert_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewDiscoveryRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewDiscoveryRepository creates a new instance of DiscoveryRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDiscoveryRepository(t mockConstructorTestingTNewDiscoveryRepository) *DiscoveryRepository {
	mock := &DiscoveryRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
