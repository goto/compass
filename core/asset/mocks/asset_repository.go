// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/goto/compass/core/asset"

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

// AddProbe provides a mock function with given fields: ctx, assetURN, probe
func (_m *AssetRepository) AddProbe(ctx context.Context, assetURN string, probe *asset.Probe) error {
	ret := _m.Called(ctx, assetURN, probe)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *asset.Probe) error); ok {
		r0 = rf(ctx, assetURN, probe)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssetRepository_AddProbe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddProbe'
type AssetRepository_AddProbe_Call struct {
	*mock.Call
}

// AddProbe is a helper method to define mock.On call
//   - ctx context.Context
//   - assetURN string
//   - probe *asset.Probe
func (_e *AssetRepository_Expecter) AddProbe(ctx interface{}, assetURN interface{}, probe interface{}) *AssetRepository_AddProbe_Call {
	return &AssetRepository_AddProbe_Call{Call: _e.mock.On("AddProbe", ctx, assetURN, probe)}
}

func (_c *AssetRepository_AddProbe_Call) Run(run func(ctx context.Context, assetURN string, probe *asset.Probe)) *AssetRepository_AddProbe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*asset.Probe))
	})
	return _c
}

func (_c *AssetRepository_AddProbe_Call) Return(_a0 error) *AssetRepository_AddProbe_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssetRepository_AddProbe_Call) RunAndReturn(run func(context.Context, string, *asset.Probe) error) *AssetRepository_AddProbe_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteByID provides a mock function with given fields: ctx, id
func (_m *AssetRepository) DeleteByID(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssetRepository_DeleteByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteByID'
type AssetRepository_DeleteByID_Call struct {
	*mock.Call
}

// DeleteByID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *AssetRepository_Expecter) DeleteByID(ctx interface{}, id interface{}) *AssetRepository_DeleteByID_Call {
	return &AssetRepository_DeleteByID_Call{Call: _e.mock.On("DeleteByID", ctx, id)}
}

func (_c *AssetRepository_DeleteByID_Call) Run(run func(ctx context.Context, id string)) *AssetRepository_DeleteByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetRepository_DeleteByID_Call) Return(_a0 error) *AssetRepository_DeleteByID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssetRepository_DeleteByID_Call) RunAndReturn(run func(context.Context, string) error) *AssetRepository_DeleteByID_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteByURN provides a mock function with given fields: ctx, urn
func (_m *AssetRepository) DeleteByURN(ctx context.Context, urn string) error {
	ret := _m.Called(ctx, urn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, urn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssetRepository_DeleteByURN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteByURN'
type AssetRepository_DeleteByURN_Call struct {
	*mock.Call
}

// DeleteByURN is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
func (_e *AssetRepository_Expecter) DeleteByURN(ctx interface{}, urn interface{}) *AssetRepository_DeleteByURN_Call {
	return &AssetRepository_DeleteByURN_Call{Call: _e.mock.On("DeleteByURN", ctx, urn)}
}

func (_c *AssetRepository_DeleteByURN_Call) Run(run func(ctx context.Context, urn string)) *AssetRepository_DeleteByURN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetRepository_DeleteByURN_Call) Return(_a0 error) *AssetRepository_DeleteByURN_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AssetRepository_DeleteByURN_Call) RunAndReturn(run func(context.Context, string) error) *AssetRepository_DeleteByURN_Call {
	_c.Call.Return(run)
	return _c
}

// GetAll provides a mock function with given fields: _a0, _a1
func (_m *AssetRepository) GetAll(_a0 context.Context, _a1 asset.Filter) ([]asset.Asset, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) ([]asset.Asset, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) []asset.Asset); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

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
//   - _a0 context.Context
//   - _a1 asset.Filter
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

func (_c *AssetRepository_GetAll_Call) RunAndReturn(run func(context.Context, asset.Filter) ([]asset.Asset, error)) *AssetRepository_GetAll_Call {
	_c.Call.Return(run)
	return _c
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *AssetRepository) GetByID(ctx context.Context, id string) (asset.Asset, error) {
	ret := _m.Called(ctx, id)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (asset.Asset, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) asset.Asset); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

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
//   - ctx context.Context
//   - id string
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

func (_c *AssetRepository_GetByID_Call) RunAndReturn(run func(context.Context, string) (asset.Asset, error)) *AssetRepository_GetByID_Call {
	_c.Call.Return(run)
	return _c
}

// GetByURN provides a mock function with given fields: ctx, urn
func (_m *AssetRepository) GetByURN(ctx context.Context, urn string) (asset.Asset, error) {
	ret := _m.Called(ctx, urn)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (asset.Asset, error)); ok {
		return rf(ctx, urn)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) asset.Asset); ok {
		r0 = rf(ctx, urn)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, urn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetByURN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByURN'
type AssetRepository_GetByURN_Call struct {
	*mock.Call
}

// GetByURN is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
func (_e *AssetRepository_Expecter) GetByURN(ctx interface{}, urn interface{}) *AssetRepository_GetByURN_Call {
	return &AssetRepository_GetByURN_Call{Call: _e.mock.On("GetByURN", ctx, urn)}
}

func (_c *AssetRepository_GetByURN_Call) Run(run func(ctx context.Context, urn string)) *AssetRepository_GetByURN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetRepository_GetByURN_Call) Return(_a0 asset.Asset, _a1 error) *AssetRepository_GetByURN_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetRepository_GetByURN_Call) RunAndReturn(run func(context.Context, string) (asset.Asset, error)) *AssetRepository_GetByURN_Call {
	_c.Call.Return(run)
	return _c
}

// GetByVersionWithID provides a mock function with given fields: ctx, id, version
func (_m *AssetRepository) GetByVersionWithID(ctx context.Context, id string, version string) (asset.Asset, error) {
	ret := _m.Called(ctx, id, version)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (asset.Asset, error)); ok {
		return rf(ctx, id, version)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) asset.Asset); ok {
		r0 = rf(ctx, id, version)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetByVersionWithID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByVersionWithID'
type AssetRepository_GetByVersionWithID_Call struct {
	*mock.Call
}

// GetByVersionWithID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - version string
func (_e *AssetRepository_Expecter) GetByVersionWithID(ctx interface{}, id interface{}, version interface{}) *AssetRepository_GetByVersionWithID_Call {
	return &AssetRepository_GetByVersionWithID_Call{Call: _e.mock.On("GetByVersionWithID", ctx, id, version)}
}

func (_c *AssetRepository_GetByVersionWithID_Call) Run(run func(ctx context.Context, id string, version string)) *AssetRepository_GetByVersionWithID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *AssetRepository_GetByVersionWithID_Call) Return(_a0 asset.Asset, _a1 error) *AssetRepository_GetByVersionWithID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetRepository_GetByVersionWithID_Call) RunAndReturn(run func(context.Context, string, string) (asset.Asset, error)) *AssetRepository_GetByVersionWithID_Call {
	_c.Call.Return(run)
	return _c
}

// GetByVersionWithURN provides a mock function with given fields: ctx, urn, version
func (_m *AssetRepository) GetByVersionWithURN(ctx context.Context, urn string, version string) (asset.Asset, error) {
	ret := _m.Called(ctx, urn, version)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (asset.Asset, error)); ok {
		return rf(ctx, urn, version)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) asset.Asset); ok {
		r0 = rf(ctx, urn, version)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, urn, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetByVersionWithURN_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByVersionWithURN'
type AssetRepository_GetByVersionWithURN_Call struct {
	*mock.Call
}

// GetByVersionWithURN is a helper method to define mock.On call
//   - ctx context.Context
//   - urn string
//   - version string
func (_e *AssetRepository_Expecter) GetByVersionWithURN(ctx interface{}, urn interface{}, version interface{}) *AssetRepository_GetByVersionWithURN_Call {
	return &AssetRepository_GetByVersionWithURN_Call{Call: _e.mock.On("GetByVersionWithURN", ctx, urn, version)}
}

func (_c *AssetRepository_GetByVersionWithURN_Call) Run(run func(ctx context.Context, urn string, version string)) *AssetRepository_GetByVersionWithURN_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *AssetRepository_GetByVersionWithURN_Call) Return(_a0 asset.Asset, _a1 error) *AssetRepository_GetByVersionWithURN_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetRepository_GetByVersionWithURN_Call) RunAndReturn(run func(context.Context, string, string) (asset.Asset, error)) *AssetRepository_GetByVersionWithURN_Call {
	_c.Call.Return(run)
	return _c
}

// GetCount provides a mock function with given fields: _a0, _a1
func (_m *AssetRepository) GetCount(_a0 context.Context, _a1 asset.Filter) (int, error) {
	ret := _m.Called(_a0, _a1)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) (int, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) int); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int)
	}

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
//   - _a0 context.Context
//   - _a1 asset.Filter
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

func (_c *AssetRepository_GetCount_Call) RunAndReturn(run func(context.Context, asset.Filter) (int, error)) *AssetRepository_GetCount_Call {
	_c.Call.Return(run)
	return _c
}

// GetProbes provides a mock function with given fields: ctx, assetURN
func (_m *AssetRepository) GetProbes(ctx context.Context, assetURN string) ([]asset.Probe, error) {
	ret := _m.Called(ctx, assetURN)

	var r0 []asset.Probe
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]asset.Probe, error)); ok {
		return rf(ctx, assetURN)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []asset.Probe); ok {
		r0 = rf(ctx, assetURN)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Probe)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, assetURN)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetProbes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProbes'
type AssetRepository_GetProbes_Call struct {
	*mock.Call
}

// GetProbes is a helper method to define mock.On call
//   - ctx context.Context
//   - assetURN string
func (_e *AssetRepository_Expecter) GetProbes(ctx interface{}, assetURN interface{}) *AssetRepository_GetProbes_Call {
	return &AssetRepository_GetProbes_Call{Call: _e.mock.On("GetProbes", ctx, assetURN)}
}

func (_c *AssetRepository_GetProbes_Call) Run(run func(ctx context.Context, assetURN string)) *AssetRepository_GetProbes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AssetRepository_GetProbes_Call) Return(_a0 []asset.Probe, _a1 error) *AssetRepository_GetProbes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetRepository_GetProbes_Call) RunAndReturn(run func(context.Context, string) ([]asset.Probe, error)) *AssetRepository_GetProbes_Call {
	_c.Call.Return(run)
	return _c
}

// GetProbesWithFilter provides a mock function with given fields: ctx, flt
func (_m *AssetRepository) GetProbesWithFilter(ctx context.Context, flt asset.ProbesFilter) (map[string][]asset.Probe, error) {
	ret := _m.Called(ctx, flt)

	var r0 map[string][]asset.Probe
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.ProbesFilter) (map[string][]asset.Probe, error)); ok {
		return rf(ctx, flt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.ProbesFilter) map[string][]asset.Probe); ok {
		r0 = rf(ctx, flt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]asset.Probe)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.ProbesFilter) error); ok {
		r1 = rf(ctx, flt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetProbesWithFilter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProbesWithFilter'
type AssetRepository_GetProbesWithFilter_Call struct {
	*mock.Call
}

// GetProbesWithFilter is a helper method to define mock.On call
//   - ctx context.Context
//   - flt asset.ProbesFilter
func (_e *AssetRepository_Expecter) GetProbesWithFilter(ctx interface{}, flt interface{}) *AssetRepository_GetProbesWithFilter_Call {
	return &AssetRepository_GetProbesWithFilter_Call{Call: _e.mock.On("GetProbesWithFilter", ctx, flt)}
}

func (_c *AssetRepository_GetProbesWithFilter_Call) Run(run func(ctx context.Context, flt asset.ProbesFilter)) *AssetRepository_GetProbesWithFilter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.ProbesFilter))
	})
	return _c
}

func (_c *AssetRepository_GetProbesWithFilter_Call) Return(_a0 map[string][]asset.Probe, _a1 error) *AssetRepository_GetProbesWithFilter_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetRepository_GetProbesWithFilter_Call) RunAndReturn(run func(context.Context, asset.ProbesFilter) (map[string][]asset.Probe, error)) *AssetRepository_GetProbesWithFilter_Call {
	_c.Call.Return(run)
	return _c
}

// GetTypes provides a mock function with given fields: ctx, flt
func (_m *AssetRepository) GetTypes(ctx context.Context, flt asset.Filter) (map[asset.Type]int, error) {
	ret := _m.Called(ctx, flt)

	var r0 map[asset.Type]int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) (map[asset.Type]int, error)); ok {
		return rf(ctx, flt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter) map[asset.Type]int); ok {
		r0 = rf(ctx, flt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[asset.Type]int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, asset.Filter) error); ok {
		r1 = rf(ctx, flt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AssetRepository_GetTypes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTypes'
type AssetRepository_GetTypes_Call struct {
	*mock.Call
}

// GetTypes is a helper method to define mock.On call
//   - ctx context.Context
//   - flt asset.Filter
func (_e *AssetRepository_Expecter) GetTypes(ctx interface{}, flt interface{}) *AssetRepository_GetTypes_Call {
	return &AssetRepository_GetTypes_Call{Call: _e.mock.On("GetTypes", ctx, flt)}
}

func (_c *AssetRepository_GetTypes_Call) Run(run func(ctx context.Context, flt asset.Filter)) *AssetRepository_GetTypes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(asset.Filter))
	})
	return _c
}

func (_c *AssetRepository_GetTypes_Call) Return(_a0 map[asset.Type]int, _a1 error) *AssetRepository_GetTypes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AssetRepository_GetTypes_Call) RunAndReturn(run func(context.Context, asset.Filter) (map[asset.Type]int, error)) *AssetRepository_GetTypes_Call {
	_c.Call.Return(run)
	return _c
}

// GetVersionHistory provides a mock function with given fields: ctx, flt, id
func (_m *AssetRepository) GetVersionHistory(ctx context.Context, flt asset.Filter, id string) ([]asset.Asset, error) {
	ret := _m.Called(ctx, flt, id)

	var r0 []asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, string) ([]asset.Asset, error)); ok {
		return rf(ctx, flt, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, asset.Filter, string) []asset.Asset); ok {
		r0 = rf(ctx, flt, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

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
//   - ctx context.Context
//   - flt asset.Filter
//   - id string
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

func (_c *AssetRepository_GetVersionHistory_Call) RunAndReturn(run func(context.Context, asset.Filter, string) ([]asset.Asset, error)) *AssetRepository_GetVersionHistory_Call {
	_c.Call.Return(run)
	return _c
}

// Upsert provides a mock function with given fields: ctx, ast
func (_m *AssetRepository) Upsert(ctx context.Context, ast *asset.Asset) (string, error) {
	ret := _m.Called(ctx, ast)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset) (string, error)); ok {
		return rf(ctx, ast)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *asset.Asset) string); ok {
		r0 = rf(ctx, ast)
	} else {
		r0 = ret.Get(0).(string)
	}

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
//   - ctx context.Context
//   - ast *asset.Asset
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

func (_c *AssetRepository_Upsert_Call) RunAndReturn(run func(context.Context, *asset.Asset) (string, error)) *AssetRepository_Upsert_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewAssetRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewAssetRepository creates a new instance of AssetRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAssetRepository(t mockConstructorTestingTNewAssetRepository) *AssetRepository {
	mock := &AssetRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
