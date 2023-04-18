// Code generated by mockery v2.25.1. DO NOT EDIT.

package mocks

import (
	context "context"

	asset "github.com/goto/compass/core/asset"

	mock "github.com/stretchr/testify/mock"

	star "github.com/goto/compass/core/star"

	user "github.com/goto/compass/core/user"
)

// StarRepository is an autogenerated mock type for the Repository type
type StarRepository struct {
	mock.Mock
}

type StarRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *StarRepository) EXPECT() *StarRepository_Expecter {
	return &StarRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, userID, assetID
func (_m *StarRepository) Create(ctx context.Context, userID string, assetID string) (string, error) {
	ret := _m.Called(ctx, userID, assetID)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, error)); ok {
		return rf(ctx, userID, assetID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, userID, assetID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, userID, assetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StarRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type StarRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - assetID string
func (_e *StarRepository_Expecter) Create(ctx interface{}, userID interface{}, assetID interface{}) *StarRepository_Create_Call {
	return &StarRepository_Create_Call{Call: _e.mock.On("Create", ctx, userID, assetID)}
}

func (_c *StarRepository_Create_Call) Run(run func(ctx context.Context, userID string, assetID string)) *StarRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *StarRepository_Create_Call) Return(_a0 string, _a1 error) *StarRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StarRepository_Create_Call) RunAndReturn(run func(context.Context, string, string) (string, error)) *StarRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, userID, assetID
func (_m *StarRepository) Delete(ctx context.Context, userID string, assetID string) error {
	ret := _m.Called(ctx, userID, assetID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, userID, assetID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StarRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type StarRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - assetID string
func (_e *StarRepository_Expecter) Delete(ctx interface{}, userID interface{}, assetID interface{}) *StarRepository_Delete_Call {
	return &StarRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, userID, assetID)}
}

func (_c *StarRepository_Delete_Call) Run(run func(ctx context.Context, userID string, assetID string)) *StarRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *StarRepository_Delete_Call) Return(_a0 error) *StarRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StarRepository_Delete_Call) RunAndReturn(run func(context.Context, string, string) error) *StarRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// GetAllAssetsByUserID provides a mock function with given fields: ctx, flt, userID
func (_m *StarRepository) GetAllAssetsByUserID(ctx context.Context, flt star.Filter, userID string) ([]asset.Asset, error) {
	ret := _m.Called(ctx, flt, userID)

	var r0 []asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, star.Filter, string) ([]asset.Asset, error)); ok {
		return rf(ctx, flt, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, star.Filter, string) []asset.Asset); ok {
		r0 = rf(ctx, flt, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]asset.Asset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, star.Filter, string) error); ok {
		r1 = rf(ctx, flt, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StarRepository_GetAllAssetsByUserID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllAssetsByUserID'
type StarRepository_GetAllAssetsByUserID_Call struct {
	*mock.Call
}

// GetAllAssetsByUserID is a helper method to define mock.On call
//   - ctx context.Context
//   - flt star.Filter
//   - userID string
func (_e *StarRepository_Expecter) GetAllAssetsByUserID(ctx interface{}, flt interface{}, userID interface{}) *StarRepository_GetAllAssetsByUserID_Call {
	return &StarRepository_GetAllAssetsByUserID_Call{Call: _e.mock.On("GetAllAssetsByUserID", ctx, flt, userID)}
}

func (_c *StarRepository_GetAllAssetsByUserID_Call) Run(run func(ctx context.Context, flt star.Filter, userID string)) *StarRepository_GetAllAssetsByUserID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(star.Filter), args[2].(string))
	})
	return _c
}

func (_c *StarRepository_GetAllAssetsByUserID_Call) Return(_a0 []asset.Asset, _a1 error) *StarRepository_GetAllAssetsByUserID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StarRepository_GetAllAssetsByUserID_Call) RunAndReturn(run func(context.Context, star.Filter, string) ([]asset.Asset, error)) *StarRepository_GetAllAssetsByUserID_Call {
	_c.Call.Return(run)
	return _c
}

// GetAssetByUserID provides a mock function with given fields: ctx, userID, assetID
func (_m *StarRepository) GetAssetByUserID(ctx context.Context, userID string, assetID string) (asset.Asset, error) {
	ret := _m.Called(ctx, userID, assetID)

	var r0 asset.Asset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (asset.Asset, error)); ok {
		return rf(ctx, userID, assetID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) asset.Asset); ok {
		r0 = rf(ctx, userID, assetID)
	} else {
		r0 = ret.Get(0).(asset.Asset)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, userID, assetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StarRepository_GetAssetByUserID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAssetByUserID'
type StarRepository_GetAssetByUserID_Call struct {
	*mock.Call
}

// GetAssetByUserID is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - assetID string
func (_e *StarRepository_Expecter) GetAssetByUserID(ctx interface{}, userID interface{}, assetID interface{}) *StarRepository_GetAssetByUserID_Call {
	return &StarRepository_GetAssetByUserID_Call{Call: _e.mock.On("GetAssetByUserID", ctx, userID, assetID)}
}

func (_c *StarRepository_GetAssetByUserID_Call) Run(run func(ctx context.Context, userID string, assetID string)) *StarRepository_GetAssetByUserID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *StarRepository_GetAssetByUserID_Call) Return(_a0 asset.Asset, _a1 error) *StarRepository_GetAssetByUserID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StarRepository_GetAssetByUserID_Call) RunAndReturn(run func(context.Context, string, string) (asset.Asset, error)) *StarRepository_GetAssetByUserID_Call {
	_c.Call.Return(run)
	return _c
}

// GetStargazers provides a mock function with given fields: ctx, flt, assetID
func (_m *StarRepository) GetStargazers(ctx context.Context, flt star.Filter, assetID string) ([]user.User, error) {
	ret := _m.Called(ctx, flt, assetID)

	var r0 []user.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, star.Filter, string) ([]user.User, error)); ok {
		return rf(ctx, flt, assetID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, star.Filter, string) []user.User); ok {
		r0 = rf(ctx, flt, assetID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]user.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, star.Filter, string) error); ok {
		r1 = rf(ctx, flt, assetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StarRepository_GetStargazers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStargazers'
type StarRepository_GetStargazers_Call struct {
	*mock.Call
}

// GetStargazers is a helper method to define mock.On call
//   - ctx context.Context
//   - flt star.Filter
//   - assetID string
func (_e *StarRepository_Expecter) GetStargazers(ctx interface{}, flt interface{}, assetID interface{}) *StarRepository_GetStargazers_Call {
	return &StarRepository_GetStargazers_Call{Call: _e.mock.On("GetStargazers", ctx, flt, assetID)}
}

func (_c *StarRepository_GetStargazers_Call) Run(run func(ctx context.Context, flt star.Filter, assetID string)) *StarRepository_GetStargazers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(star.Filter), args[2].(string))
	})
	return _c
}

func (_c *StarRepository_GetStargazers_Call) Return(_a0 []user.User, _a1 error) *StarRepository_GetStargazers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StarRepository_GetStargazers_Call) RunAndReturn(run func(context.Context, star.Filter, string) ([]user.User, error)) *StarRepository_GetStargazers_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewStarRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewStarRepository creates a new instance of StarRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStarRepository(t mockConstructorTestingTNewStarRepository) *StarRepository {
	mock := &StarRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
