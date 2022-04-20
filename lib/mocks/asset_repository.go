// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/odpf/compass/asset"

	mock "github.com/stretchr/testify/mock"
)

// AssetRepository is an autogenerated mock type for the Repository type
type AssetRepository struct {
	mock.Mock
}

type AssetRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *AssetRepository) EXPECT() *AssetRepository_Expecter {
	return &AssetRepository_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, id
func (_m *AssetRepository) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssetRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type AssetRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//  - ctx context.Context
//  - id string
func (_e *AssetRepository_Expecter) Delete(ctx interface{}, id interface{}) *AssetRepository_Delete_Call {
	return &AssetRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *AssetRepository_Delete_Call) Run(run func(ctx context.Context, id string)) *AssetRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetRepository_Delete_Call) Return(_a0 error) *AssetRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

// Find provides a mock function with given fields: ctx, urn, typ, service
func (_m *AssetRepository) Find(ctx context.Context, urn string, typ asset.Type, service string) (asset.Asset, error) {
	ret := _m.Called(ctx, urn, typ, service)

	var r0 asset.Asset
	if rf, ok := ret.Get(0).(func(context.Context, string, asset.Type, string) asset.Asset); ok {
		r0 = rf(ctx, urn, typ, service)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, asset.Type, string) error); ok {
		r1 = rf(ctx, urn, typ, service)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type AssetRepository_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//  - ctx context.Context
//  - urn string
//  - typ asset.Type
//  - service string
func (_e *AssetRepository_Expecter) Find(ctx interface{}, urn interface{}, typ interface{}, service interface{}) *AssetRepository_Find_Call {
	return &AssetRepository_Find_Call{Call: _e.mock.On("Find", ctx, urn, typ, service)}
}

func (_c *AssetRepository_Find_Call) Run(run func(ctx context.Context, urn string, typ asset.Type, service string)) *AssetRepository_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(asset.Type), args[3].(string))
	})
	return _c
}

func (_c *AssetRepository_Find_Call) Return(_a0 asset.Asset, _a1 error) *AssetRepository_Find_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetAll provides a mock function with given fields: _a0, _a1
func (_m *AssetRepository) GetAll(_a0 context.Context, _a1 asset.Filter) ([]asset.Asset, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []asset.Asset
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) []asset.Asset); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAll'
type AssetRepository_GetAll_Call struct {
	*mock.Call
}

// GetAll is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 asset.Filter
func (_e *AssetRepository_Expecter) GetAll(_a0 interface{}, _a1 interface{}) *AssetRepository_GetAll_Call {
	return &AssetRepository_GetAll_Call{Call: _e.mock.On("GetAll", _a0, _a1)}
}

func (_c *AssetRepository_GetAll_Call) Run(run func(_a0 context.Context, _a1 asset.Filter)) *AssetRepository_GetAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter))
	})
	return _c
}

func (_c *AssetRepository_GetAll_Call) Return(_a0 []asset.Asset, _a1 error) *AssetRepository_GetAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *AssetRepository) GetByID(ctx context.Context, id string) (asset.Asset, error) {
	ret := _m.Called(ctx, id)

	var r0 asset.Asset
	if rf, ok := ret.Get(0).(func(context.Context, string) asset.Asset); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByID'
type AssetRepository_GetByID_Call struct {
	*mock.Call
}

// GetByID is a helper method to define mock.On call
//  - ctx context.Context
//  - id string
func (_e *AssetRepository_Expecter) GetByID(ctx interface{}, id interface{}) *AssetRepository_GetByID_Call {
	return &AssetRepository_GetByID_Call{Call: _e.mock.On("GetByID", ctx, id)}
}

func (_c *AssetRepository_GetByID_Call) Run(run func(ctx context.Context, id string)) *AssetRepository_GetByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetRepository_GetByID_Call) Return(_a0 asset.Asset, _a1 error) *AssetRepository_GetByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetByVersion provides a mock function with given fields: ctx, id, version
func (_m *AssetRepository) GetByVersion(ctx context.Context, id string, version string) (asset.Asset, error) {
	ret := _m.Called(ctx, id, version)

	var r0 asset.Asset
	if rf, ok := ret.Get(0).(func(context.Context, string, string) asset.Asset); ok {
		r0 = rf(ctx, id, version)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetByVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByVersion'
type AssetRepository_GetByVersion_Call struct {
	*mock.Call
}

// GetByVersion is a helper method to define mock.On call
//  - ctx context.Context
//  - id string
//  - version string
func (_e *AssetRepository_Expecter) GetByVersion(ctx interface{}, id interface{}, version interface{}) *AssetRepository_GetByVersion_Call {
	return &AssetRepository_GetByVersion_Call{Call: _e.mock.On("GetByVersion", ctx, id, version)}
}

func (_c *AssetRepository_GetByVersion_Call) Run(run func(ctx context.Context, id string, version string)) *AssetRepository_GetByVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *AssetRepository_GetByVersion_Call) Return(_a0 asset.Asset, _a1 error) *AssetRepository_GetByVersion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetCount provides a mock function with given fields: _a0, _a1
func (_m *AssetRepository) GetCount(_a0 context.Context, _a1 asset.Filter) (int, error) {
	ret := _m.Called(_a0, _a1)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) int); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetCount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCount'
type AssetRepository_GetCount_Call struct {
	*mock.Call
}

// GetCount is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 asset.Filter
func (_e *AssetRepository_Expecter) GetCount(_a0 interface{}, _a1 interface{}) *AssetRepository_GetCount_Call {
	return &AssetRepository_GetCount_Call{Call: _e.mock.On("GetCount", _a0, _a1)}
}

func (_c *AssetRepository_GetCount_Call) Run(run func(_a0 context.Context, _a1 asset.Filter)) *AssetRepository_GetCount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter))
	})
	return _c
}

func (_c *AssetRepository_GetCount_Call) Return(_a0 int, _a1 error) *AssetRepository_GetCount_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetVersionHistory provides a mock function with given fields: ctx, flt, id
func (_m *AssetRepository) GetVersionHistory(ctx context.Context, flt asset.Filter, id string) ([]asset.Asset, error) {
	ret := _m.Called(ctx, flt, id)

	var r0 []asset.Asset
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, string) []asset.Asset); ok {
		r0 = rf(ctx, flt, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter, string) error); ok {
		r1 = rf(ctx, flt, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetVersionHistory_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetVersionHistory'
type AssetRepository_GetVersionHistory_Call struct {
	*mock.Call
}

// GetVersionHistory is a helper method to define mock.On call
//  - ctx context.Context
//  - flt asset.Filter
//  - id string
func (_e *AssetRepository_Expecter) GetVersionHistory(ctx interface{}, flt interface{}, id interface{}) *AssetRepository_GetVersionHistory_Call {
	return &AssetRepository_GetVersionHistory_Call{Call: _e.mock.On("GetVersionHistory", ctx, flt, id)}
}

func (_c *AssetRepository_GetVersionHistory_Call) Run(run func(ctx context.Context, flt asset.Filter, id string)) *AssetRepository_GetVersionHistory_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter), args[2].(string))
	})
	return _c
}

func (_c *AssetRepository_GetVersionHistory_Call) Return(_a0 []asset.Asset, _a1 error) *AssetRepository_GetVersionHistory_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Upsert provides a mock function with given fields: ctx, ast
func (_m *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) (string, error) {
	ret := _m.Called(ctx, ast)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset) string); ok {
		r0 = rf(ctx, ast)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *asset.Asset) error); ok {
		r1 = rf(ctx, ast)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type AssetRepository_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//  - ctx context.Context
//  - ast *asset.Asset
func (_e *AssetRepository_Expecter) Upsert(ctx interface{}, ast interface{}) *AssetRepository_Upsert_Call {
	return &AssetRepository_Upsert_Call{Call: _e.mock.On("Upsert", ctx, ast)}
}

func (_c *AssetRepository_Upsert_Call) Run(run func(ctx context.Context, ast *asset.Asset)) *AssetRepository_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*asset.Asset))
	})
	return _c
}

func (_c *AssetRepository_Upsert_Call) Return(_a0 string, _a1 error) *AssetRepository_Upsert_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
