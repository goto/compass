// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/odpf/compass/asset"

	discovery "github.com/odpf/compass/discovery"

	mock "github.com/stretchr/testify/mock"
)

// DiscoveryRepository is an autogenerated mock type for the Repository type
type DiscoveryRepository struct {
	mock.Mock
}

type DiscoveryRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *DiscoveryRepository) EXPECT() *DiscoveryRepository_Expecter {
	return &DiscoveryRepository_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, assetID
func (_m *DiscoveryRepository) Delete(ctx context.Context, assetID string) error {
	ret := _m.Called(ctx, assetID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, assetID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DiscoveryRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type DiscoveryRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//  - ctx context.Context
//  - assetID string
func (_e *DiscoveryRepository_Expecter) Delete(ctx interface{}, assetID interface{}) *DiscoveryRepository_Delete_Call {
	return &DiscoveryRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, assetID)}
}

func (_c *DiscoveryRepository_Delete_Call) Run(run func(ctx context.Context, assetID string)) *DiscoveryRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DiscoveryRepository_Delete_Call) Return(_a0 error) *DiscoveryRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetTypes provides a mock function with given fields: _a0
func (_m *DiscoveryRepository) GetTypes(_a0 context.Context) (map[asset.Type]int, error) {
	ret := _m.Called(_a0)

	var r0 map[asset.Type]int
	if rf, ok := ret.Get(0).(func(context.Context) map[asset.Type]int); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[asset.Type]int)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DiscoveryRepository_GetTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTypes'
type DiscoveryRepository_GetTypes_Call struct {
	*mock.Call
}

// GetTypes is a helper method to define mock.On call
//  - _a0 context.Context
func (_e *DiscoveryRepository_Expecter) GetTypes(_a0 interface{}) *DiscoveryRepository_GetTypes_Call {
	return &DiscoveryRepository_GetTypes_Call{Call: _e.mock.On("GetTypes", _a0)}
}

func (_c *DiscoveryRepository_GetTypes_Call) Run(run func(_a0 context.Context)) *DiscoveryRepository_GetTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DiscoveryRepository_GetTypes_Call) Return(_a0 map[asset.Type]int, _a1 error) *DiscoveryRepository_GetTypes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Search provides a mock function with given fields: ctx, cfg
func (_m *DiscoveryRepository) Search(ctx context.Context, cfg discovery.SearchConfig) ([]discovery.SearchResult, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []discovery.SearchResult
	if rf, ok := ret.Get(0).(func(context.Context, discovery.SearchConfig) []discovery.SearchResult); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]discovery.SearchResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, discovery.SearchConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DiscoveryRepository_Search_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Search'
type DiscoveryRepository_Search_Call struct {
	*mock.Call
}

// Search is a helper method to define mock.On call
//  - ctx context.Context
//  - cfg discovery.SearchConfig
func (_e *DiscoveryRepository_Expecter) Search(ctx interface{}, cfg interface{}) *DiscoveryRepository_Search_Call {
	return &DiscoveryRepository_Search_Call{Call: _e.mock.On("Search", ctx, cfg)}
}

func (_c *DiscoveryRepository_Search_Call) Run(run func(ctx context.Context, cfg discovery.SearchConfig)) *DiscoveryRepository_Search_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(discovery.SearchConfig))
	})
	return _c
}

func (_c *DiscoveryRepository_Search_Call) Return(results []discovery.SearchResult, err error) *DiscoveryRepository_Search_Call {
	_c.Call.Return(results, err)
	return _c
}

// Suggest provides a mock function with given fields: ctx, cfg
func (_m *DiscoveryRepository) Suggest(ctx context.Context, cfg discovery.SearchConfig) ([]string, error) {
	ret := _m.Called(ctx, cfg)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, discovery.SearchConfig) []string); ok {
		r0 = rf(ctx, cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, discovery.SearchConfig) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DiscoveryRepository_Suggest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Suggest'
type DiscoveryRepository_Suggest_Call struct {
	*mock.Call
}

// Suggest is a helper method to define mock.On call
//  - ctx context.Context
//  - cfg discovery.SearchConfig
func (_e *DiscoveryRepository_Expecter) Suggest(ctx interface{}, cfg interface{}) *DiscoveryRepository_Suggest_Call {
	return &DiscoveryRepository_Suggest_Call{Call: _e.mock.On("Suggest", ctx, cfg)}
}

func (_c *DiscoveryRepository_Suggest_Call) Run(run func(ctx context.Context, cfg discovery.SearchConfig)) *DiscoveryRepository_Suggest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(discovery.SearchConfig))
	})
	return _c
}

func (_c *DiscoveryRepository_Suggest_Call) Return(suggestions []string, err error) *DiscoveryRepository_Suggest_Call {
	_c.Call.Return(suggestions, err)
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
//  - _a0 context.Context
//  - _a1 asset.Asset
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
