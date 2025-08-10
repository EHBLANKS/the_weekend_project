package service

import (
	"context"
	"testing"

	dbmocks "sports/db/mocks"

	sportspb "sports/proto/sports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListEvents_DefaultsAndNextOffset(t *testing.T) {
	m := new(dbmocks.EventsRepo)

	// Build a full "page" (50) to trigger next_offset
	evs := make([]*sportspb.Event, 50)
	for i := range evs {
		evs[i] = &sportspb.Event{Id: int64(i + 1)}
	}

	m.On("List",
		mock.Anything,
		(*sportspb.ListEventsRequestFilter)(nil),
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_ASCENDING,
		int32(50), int32(0),
	).Return(evs, nil).Once()

	svc := &sportsService{repo: m}
	resp, err := svc.ListEvents(context.Background(), &sportspb.ListEventsRequest{})
	require.NoError(t, err)
	assert.Equal(t, int32(50), resp.GetNextOffset()) // full page -> next = 50
	m.AssertExpectations(t)
}

func TestListEvents_ExplicitLimitOffset_NoNextWhenPartial(t *testing.T) {
	m := new(dbmocks.EventsRepo)

	req := &sportspb.ListEventsRequest{
		Filter:    &sportspb.ListEventsRequestFilter{},
		SortBy:    sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		Direction: sportspb.SortDirection_SORT_DIRECTION_DESCENDING,
		Limit:     3,
		Offset:    6,
	}

	m.On("List",
		mock.Anything,
		req.GetFilter(),
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_DESCENDING,
		int32(3), int32(6),
	).Return([]*sportspb.Event{
		{Id: 7}, {Id: 6}, // only 2 => partial page => next_offset = 0
	}, nil).Once()

	svc := &sportsService{repo: m}
	resp, err := svc.ListEvents(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetNextOffset())
	m.AssertExpectations(t)
}

func TestListEvents_VisibleOnly_ForwardsFilter(t *testing.T) {
	m := new(dbmocks.EventsRepo)

	req := &sportspb.ListEventsRequest{
		Filter: &sportspb.ListEventsRequestFilter{Visibility: sportspb.Visibility_VISIBILITY_VISIBLE_ONLY},
		Limit:  5,
		Offset: 0,
	}
	m.On("List",
		mock.Anything,
		mock.MatchedBy(func(f *sportspb.ListEventsRequestFilter) bool {
			return f != nil && f.GetVisibility() == sportspb.Visibility_VISIBILITY_VISIBLE_ONLY
		}),
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_ASCENDING,
		int32(5), int32(0),
	).Return([]*sportspb.Event{
		{Id: 10, Visible: true},
	}, nil).Once()

	svc := &sportsService{repo: m}
	resp, err := svc.ListEvents(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.GetEvents(), 1)
	assert.True(t, resp.GetEvents()[0].GetVisible())

	m.AssertExpectations(t)
}
