// Code generated by mockery v2.12.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	tag "github.com/goto/compass/core/tag"

	testing "testing"
)

// TagTemplateService is an autogenerated mock type for the TagTemplateService type
type TagTemplateService struct {
	mock.Mock
}

type TagTemplateService_Expecter struct {
	mock *mock.Mock
}

func (_m *TagTemplateService) EXPECT() *TagTemplateService_Expecter {
	return &TagTemplateService_Expecter{mock: &_m.Mock}
}

// CreateTemplate provides a mock function with given fields: ctx, template
func (_m *TagTemplateService) CreateTemplate(ctx context.Context, template *tag.Template) error {
	ret := _m.Called(ctx, template)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *tag.Template) error); ok {
		r0 = rf(ctx, template)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TagTemplateService_CreateTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTemplate'
type TagTemplateService_CreateTemplate_Call struct {
	*mock.Call
}

// CreateTemplate is a helper method to define mock.On call
//  - ctx context.Context
//  - template *tag.Template
func (_e *TagTemplateService_Expecter) CreateTemplate(ctx interface{}, template interface{}) *TagTemplateService_CreateTemplate_Call {
	return &TagTemplateService_CreateTemplate_Call{Call: _e.mock.On("CreateTemplate", ctx, template)}
}

func (_c *TagTemplateService_CreateTemplate_Call) Run(run func(ctx context.Context, template *tag.Template)) *TagTemplateService_CreateTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*tag.Template))
	})
	return _c
}

func (_c *TagTemplateService_CreateTemplate_Call) Return(_a0 error) *TagTemplateService_CreateTemplate_Call {
	_c.Call.Return(_a0)
	return _c
}

// DeleteTemplate provides a mock function with given fields: ctx, urn
func (_m *TagTemplateService) DeleteTemplate(ctx context.Context, urn string) error {
	ret := _m.Called(ctx, urn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, urn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TagTemplateService_DeleteTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteTemplate'
type TagTemplateService_DeleteTemplate_Call struct {
	*mock.Call
}

// DeleteTemplate is a helper method to define mock.On call
//  - ctx context.Context
//  - urn string
func (_e *TagTemplateService_Expecter) DeleteTemplate(ctx interface{}, urn interface{}) *TagTemplateService_DeleteTemplate_Call {
	return &TagTemplateService_DeleteTemplate_Call{Call: _e.mock.On("DeleteTemplate", ctx, urn)}
}

func (_c *TagTemplateService_DeleteTemplate_Call) Run(run func(ctx context.Context, urn string)) *TagTemplateService_DeleteTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *TagTemplateService_DeleteTemplate_Call) Return(_a0 error) *TagTemplateService_DeleteTemplate_Call {
	_c.Call.Return(_a0)
	return _c
}

// GetTemplate provides a mock function with given fields: ctx, urn
func (_m *TagTemplateService) GetTemplate(ctx context.Context, urn string) (tag.Template, error) {
	ret := _m.Called(ctx, urn)

	var r0 tag.Template
	if rf, ok := ret.Get(0).(func(context.Context, string) tag.Template); ok {
		r0 = rf(ctx, urn)
	} else {
		r0 = ret.Get(0).(tag.Template)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, urn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TagTemplateService_GetTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTemplate'
type TagTemplateService_GetTemplate_Call struct {
	*mock.Call
}

// GetTemplate is a helper method to define mock.On call
//  - ctx context.Context
//  - urn string
func (_e *TagTemplateService_Expecter) GetTemplate(ctx interface{}, urn interface{}) *TagTemplateService_GetTemplate_Call {
	return &TagTemplateService_GetTemplate_Call{Call: _e.mock.On("GetTemplate", ctx, urn)}
}

func (_c *TagTemplateService_GetTemplate_Call) Run(run func(ctx context.Context, urn string)) *TagTemplateService_GetTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *TagTemplateService_GetTemplate_Call) Return(_a0 tag.Template, _a1 error) *TagTemplateService_GetTemplate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// GetTemplates provides a mock function with given fields: ctx, templateURN
func (_m *TagTemplateService) GetTemplates(ctx context.Context, templateURN string) ([]tag.Template, error) {
	ret := _m.Called(ctx, templateURN)

	var r0 []tag.Template
	if rf, ok := ret.Get(0).(func(context.Context, string) []tag.Template); ok {
		r0 = rf(ctx, templateURN)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]tag.Template)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, templateURN)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TagTemplateService_GetTemplates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTemplates'
type TagTemplateService_GetTemplates_Call struct {
	*mock.Call
}

// GetTemplates is a helper method to define mock.On call
//  - ctx context.Context
//  - templateURN string
func (_e *TagTemplateService_Expecter) GetTemplates(ctx interface{}, templateURN interface{}) *TagTemplateService_GetTemplates_Call {
	return &TagTemplateService_GetTemplates_Call{Call: _e.mock.On("GetTemplates", ctx, templateURN)}
}

func (_c *TagTemplateService_GetTemplates_Call) Run(run func(ctx context.Context, templateURN string)) *TagTemplateService_GetTemplates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *TagTemplateService_GetTemplates_Call) Return(_a0 []tag.Template, _a1 error) *TagTemplateService_GetTemplates_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// UpdateTemplate provides a mock function with given fields: ctx, templateURN, template
func (_m *TagTemplateService) UpdateTemplate(ctx context.Context, templateURN string, template *tag.Template) error {
	ret := _m.Called(ctx, templateURN, template)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *tag.Template) error); ok {
		r0 = rf(ctx, templateURN, template)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TagTemplateService_UpdateTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateTemplate'
type TagTemplateService_UpdateTemplate_Call struct {
	*mock.Call
}

// UpdateTemplate is a helper method to define mock.On call
//  - ctx context.Context
//  - templateURN string
//  - template *tag.Template
func (_e *TagTemplateService_Expecter) UpdateTemplate(ctx interface{}, templateURN interface{}, template interface{}) *TagTemplateService_UpdateTemplate_Call {
	return &TagTemplateService_UpdateTemplate_Call{Call: _e.mock.On("UpdateTemplate", ctx, templateURN, template)}
}

func (_c *TagTemplateService_UpdateTemplate_Call) Run(run func(ctx context.Context, templateURN string, template *tag.Template)) *TagTemplateService_UpdateTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*tag.Template))
	})
	return _c
}

func (_c *TagTemplateService_UpdateTemplate_Call) Return(_a0 error) *TagTemplateService_UpdateTemplate_Call {
	_c.Call.Return(_a0)
	return _c
}

// Validate provides a mock function with given fields: template
func (_m *TagTemplateService) Validate(template tag.Template) error {
	ret := _m.Called(template)

	var r0 error
	if rf, ok := ret.Get(0).(func(tag.Template) error); ok {
		r0 = rf(template)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TagTemplateService_Validate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Validate'
type TagTemplateService_Validate_Call struct {
	*mock.Call
}

// Validate is a helper method to define mock.On call
//  - template tag.Template
func (_e *TagTemplateService_Expecter) Validate(template interface{}) *TagTemplateService_Validate_Call {
	return &TagTemplateService_Validate_Call{Call: _e.mock.On("Validate", template)}
}

func (_c *TagTemplateService_Validate_Call) Run(run func(template tag.Template)) *TagTemplateService_Validate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(tag.Template))
	})
	return _c
}

func (_c *TagTemplateService_Validate_Call) Return(_a0 error) *TagTemplateService_Validate_Call {
	_c.Call.Return(_a0)
	return _c
}

// NewTagTemplateService creates a new instance of TagTemplateService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewTagTemplateService(t testing.TB) *TagTemplateService {
	mock := &TagTemplateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
