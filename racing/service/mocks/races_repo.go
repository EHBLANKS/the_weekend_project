package mocks

import (
	"context"
	"racing/proto/racing"

	"github.com/stretchr/testify/mock"
)

type RacesRepo struct {
	mock.Mock
}

func (_m *RacesRepo) Init() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *RacesRepo) List(ctx context.Context, filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	ret := _m.Called(ctx, filter)

	var r0 []*racing.Race
	if rf, ok := ret.Get(0).(func(context.Context, *racing.ListRacesRequestFilter) []*racing.Race); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*racing.Race)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *racing.ListRacesRequestFilter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
