// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"

	gamesvc "github.com/knightpp/alias-proto/go/game_service"
	mock "github.com/stretchr/testify/mock"
)

// Player is an autogenerated mock type for the Player type
type Player struct {
	mock.Mock
}

type Player_Expecter struct {
	mock *mock.Mock
}

func (_m *Player) EXPECT() *Player_Expecter {
	return &Player_Expecter{mock: &_m.Mock}
}

// GetPlayer provides a mock function with given fields: ctx, token
func (_m *Player) GetPlayer(ctx context.Context, token string) (*gamesvc.Player, error) {
	ret := _m.Called(ctx, token)

	var r0 *gamesvc.Player
	if rf, ok := ret.Get(0).(func(context.Context, string) *gamesvc.Player); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gamesvc.Player)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Player_GetPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayer'
type Player_GetPlayer_Call struct {
	*mock.Call
}

// GetPlayer is a helper method to define mock.On call
//   - ctx context.Context
//   - token string
func (_e *Player_Expecter) GetPlayer(ctx interface{}, token interface{}) *Player_GetPlayer_Call {
	return &Player_GetPlayer_Call{Call: _e.mock.On("GetPlayer", ctx, token)}
}

func (_c *Player_GetPlayer_Call) Run(run func(ctx context.Context, token string)) *Player_GetPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Player_GetPlayer_Call) Return(_a0 *gamesvc.Player, _a1 error) *Player_GetPlayer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// SetPlayer provides a mock function with given fields: ctx, token, p
func (_m *Player) SetPlayer(ctx context.Context, token string, p *gamesvc.Player) error {
	ret := _m.Called(ctx, token, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *gamesvc.Player) error); ok {
		r0 = rf(ctx, token, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Player_SetPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetPlayer'
type Player_SetPlayer_Call struct {
	*mock.Call
}

// SetPlayer is a helper method to define mock.On call
//   - ctx context.Context
//   - token string
//   - p *gamesvc.Player
func (_e *Player_Expecter) SetPlayer(ctx interface{}, token interface{}, p interface{}) *Player_SetPlayer_Call {
	return &Player_SetPlayer_Call{Call: _e.mock.On("SetPlayer", ctx, token, p)}
}

func (_c *Player_SetPlayer_Call) Run(run func(ctx context.Context, token string, p *gamesvc.Player)) *Player_SetPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*gamesvc.Player))
	})
	return _c
}

func (_c *Player_SetPlayer_Call) Return(_a0 error) *Player_SetPlayer_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewPlayer interface {
	mock.TestingT
	Cleanup(func())
}

// NewPlayer creates a new instance of Player. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPlayer(t mockConstructorTestingTNewPlayer) *Player {
	mock := &Player{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
