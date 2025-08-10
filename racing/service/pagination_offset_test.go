package service

import (
	"context"
	"testing"

	dbmocks "racing/db/mocks"
	racing "racing/proto/racing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRaces_OffsetLimit_AndNextOffset(t *testing.T) {
	tests := []struct {
		name        string
		req         *racing.ListRacesRequest
		mockReturnN int   // how many races repo returns
		wantLimit   int32 // what service should pass to repo (after clamping)
		wantOffset  int32
		wantNext    int32 // expected NextOffset
	}{
		{
			name:        "defaults clamp: limit=0->50, offset<0->0; next when full page",
			req:         &racing.ListRacesRequest{Limit: 0, Offset: -10}, // default clamp
			mockReturnN: 50,                                              // repo returns full page
			wantLimit:   50, wantOffset: 0, wantNext: 50,
		},
		{
			name:        "explicit limit/offset; next when full page",
			req:         &racing.ListRacesRequest{Limit: 2, Offset: 4},
			mockReturnN: 2,
			wantLimit:   2, wantOffset: 4, wantNext: 6,
		},
		{
			name:        "no next when partial page",
			req:         &racing.ListRacesRequest{Limit: 3, Offset: 6},
			mockReturnN: 2,
			wantLimit:   3, wantOffset: 6, wantNext: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(dbmocks.RacesRepo)

			// Build the slice of mock races with tc.mockReturnN length
			out := make([]*racing.Race, 0, tc.mockReturnN)
			for i := 0; i < tc.mockReturnN; i++ {
				out = append(out, &racing.Race{Id: int64(100 + i)})
			}

			// Expect List(ctx, filter, sortBy, dir, limit, offset)
			m.
				On("List",
					mock.Anything,
					tc.req.GetFilter(),
					racing.SortBy_SORT_BY_ADVERTISED_START_TIME,   // defaulted by service
					racing.SortDirection_SORT_DIRECTION_ASCENDING, // defaulted by service
					tc.wantLimit,
					tc.wantOffset,
				).
				Return(out, nil).
				Once()

			svc := &racingService{racesRepo: m}

			resp, err := svc.ListRaces(context.Background(), tc.req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)

			// NextOffset semantics
			assert.Equal(t, tc.wantNext, resp.GetNextOffset())

			m.AssertExpectations(t)
		})
	}
}
