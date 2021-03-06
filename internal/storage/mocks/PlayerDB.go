// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1 "github.com/knightpp/alias-proto/go/pkg/model/v1"
)

// PlayerDB is an autogenerated mock type for the PlayerDB type
type PlayerDB struct {
	mock.Mock
}

type PlayerDB_Expecter struct {
	mock *mock.Mock
}

func (_m *PlayerDB) EXPECT() *PlayerDB_Expecter {
	return &PlayerDB_Expecter{mock: &_m.Mock}
}

// GetPlayer provides a mock function with given fields: ctx, playerID
func (_m *PlayerDB) GetPlayer(ctx context.Context, playerID string) (*v1.Player, error) {
	ret := _m.Called(ctx, playerID)

	var r0 *v1.Player
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1.Player); ok {
		r0 = rf(ctx, playerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Player)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, playerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PlayerDB_GetPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayer'
type PlayerDB_GetPlayer_Call struct {
	*mock.Call
}

// GetPlayer is a helper method to define mock.On call
//  - ctx context.Context
//  - playerID string
func (_e *PlayerDB_Expecter) GetPlayer(ctx interface{}, playerID interface{}) *PlayerDB_GetPlayer_Call {
	return &PlayerDB_GetPlayer_Call{Call: _e.mock.On("GetPlayer", ctx, playerID)}
}

func (_c *PlayerDB_GetPlayer_Call) Run(run func(ctx context.Context, playerID string)) *PlayerDB_GetPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *PlayerDB_GetPlayer_Call) Return(_a0 *v1.Player, _a1 error) *PlayerDB_GetPlayer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// SetPlayer provides a mock function with given fields: ctx, p
func (_m *PlayerDB) SetPlayer(ctx context.Context, p *v1.Player) error {
	ret := _m.Called(ctx, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Player) error); ok {
		r0 = rf(ctx, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PlayerDB_SetPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPlayer'
type PlayerDB_SetPlayer_Call struct {
	*mock.Call
}

// SetPlayer is a helper method to define mock.On call
//  - ctx context.Context
//  - p *v1.Player
func (_e *PlayerDB_Expecter) SetPlayer(ctx interface{}, p interface{}) *PlayerDB_SetPlayer_Call {
	return &PlayerDB_SetPlayer_Call{Call: _e.mock.On("SetPlayer", ctx, p)}
}

func (_c *PlayerDB_SetPlayer_Call) Run(run func(ctx context.Context, p *v1.Player)) *PlayerDB_SetPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1.Player))
	})
	return _c
}

func (_c *PlayerDB_SetPlayer_Call) Return(_a0 error) *PlayerDB_SetPlayer_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewPlayerDB interface {
	mock.TestingT
	Cleanup(func())
}

// NewPlayerDB creates a new instance of PlayerDB. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPlayerDB(t mockConstructorTestingTNewPlayerDB) *PlayerDB {
	mock := &PlayerDB{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
