package service

import (
	"context"
	"testing"

	dbmocks "racing/db/mocks"
	racing "racing/proto/racing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListRaces_ForwardsFilterAndSort(t *testing.T) {
	// Dummy races that the repo mock can return (order matters per test)
	r1 := &racing.Race{Id: 1, Visible: true}
	r2 := &racing.Race{Id: 2, Visible: false}
	r3 := &racing.Race{Id: 3, Visible: true}

	type testCase struct {
		name       string
		req        *racing.ListRacesRequest
		expectVis  racing.Visibility
		expectSort racing.SortBy
		expectDir  racing.SortDirection
		wantLimit  int32
		wantOffset int32
		repoReply  []*racing.Race
		wantIDs    []int64 // expected IDs in the same order returned by the service
	}

	tests := []testCase{
		{
			name:       "defaults to ADVERTISED_START_TIME ASC",
			req:        &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}},
			expectVis:  racing.Visibility_VISIBILITY_UNSPECIFIED,
			expectSort: racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
			expectDir:  racing.SortDirection_SORT_DIRECTION_ASCENDING,
			wantLimit:  50, // service default clamp
			wantOffset: 0,
			repoReply:  []*racing.Race{r2, r1, r3},
			wantIDs:    []int64{2, 1, 3},
		},
		{
			name: "visible only forwards visibility + defaults",
			req: &racing.ListRacesRequest{
				Filter: &racing.ListRacesRequestFilter{
					Visibility: racing.Visibility_VISIBILITY_VISIBLE_ONLY,
				},
			},
			expectVis:  racing.Visibility_VISIBILITY_VISIBLE_ONLY,
			expectSort: racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
			expectDir:  racing.SortDirection_SORT_DIRECTION_ASCENDING,
			wantLimit:  50, // service default clamp
			wantOffset: 0,
			repoReply:  []*racing.Race{r1, r3},
			wantIDs:    []int64{1, 3},
		},
		{
			name: "explicit DESC + explicit limit/offset",
			req: &racing.ListRacesRequest{
				Filter:    &racing.ListRacesRequestFilter{},
				SortBy:    racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
				Direction: racing.SortDirection_SORT_DIRECTION_DESCENDING,
				Limit:     2,
				Offset:    1,
			},
			expectVis:  racing.Visibility_VISIBILITY_UNSPECIFIED,
			expectSort: racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
			expectDir:  racing.SortDirection_SORT_DIRECTION_DESCENDING,
			wantLimit:  2,
			wantOffset: 1,
			repoReply:  []*racing.Race{r3, r2},
			wantIDs:    []int64{3, 2},
		},
	}

	for i := range tests {
		tc := tests[i] // avoid range var capture
		t.Run(tc.name, func(t *testing.T) {
			m := new(dbmocks.RacesRepo)

			// Expect service to forward ctx, filter (we check Visibility), sortBy, dir, limit, offset
			m.On(
				"List",
				mock.Anything, // ctx
				mock.MatchedBy(func(f *racing.ListRacesRequestFilter) bool {
					// if req.Filter is nil, service passes nil; handle both
					if f == nil {
						return tc.expectVis == racing.Visibility_VISIBILITY_UNSPECIFIED
					}
					return f.GetVisibility() == tc.expectVis
				}),
				tc.expectSort,
				tc.expectDir,
				tc.wantLimit,
				tc.wantOffset,
			).Return(tc.repoReply, nil).Once()

			svc := &racingService{racesRepo: m}

			resp, err := svc.ListRaces(context.Background(), tc.req)
			require.NoError(t, err)
			require.NotNil(t, resp)

			// Collect IDs in order and (for visible-only) verify visibility
			got := make([]int64, 0, len(resp.GetRaces()))
			for _, r := range resp.GetRaces() {
				got = append(got, r.GetId())
				if tc.expectVis == racing.Visibility_VISIBILITY_VISIBLE_ONLY {
					assert.True(t, r.GetVisible(), "expected only visible races")
				}
			}

			assert.Equal(t, tc.wantIDs, got)
			m.AssertExpectations(t)
		})
	}
}
