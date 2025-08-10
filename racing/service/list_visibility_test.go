package service

import (
	"context"
	"testing"
	"time"

	dbmocks "racing/db/mocks"
	racing "racing/proto/racing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestListRaces_ForwardsFilterAndSort(t *testing.T) {
	now := time.Now()

	r1 := &racing.Race{Id: 1, Visible: true, AdvertisedStartTime: timestamppb.New(now.Add(1 * time.Minute))}
	r2 := &racing.Race{Id: 2, Visible: false, AdvertisedStartTime: timestamppb.New(now.Add(2 * time.Minute))}
	r3 := &racing.Race{Id: 3, Visible: true, AdvertisedStartTime: timestamppb.New(now.Add(3 * time.Minute))}

	type testCase struct {
		name       string
		req        *racing.ListRacesRequest
		expectVis  racing.Visibility
		expectSort racing.SortBy
		expectDir  racing.SortDirection
		repoReply  []*racing.Race
		wantIDs    []int64
	}

	tests := []testCase{
		{
			name:       "defaults to ADVERTISED_START_TIME ASC",
			req:        &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}},
			expectVis:  racing.Visibility_VISIBILITY_UNSPECIFIED,
			expectSort: racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
			expectDir:  racing.SortDirection_SORT_DIRECTION_ASCENDING,
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
			repoReply:  []*racing.Race{r1, r3},
			wantIDs:    []int64{1, 3},
		},
		{
			name: "explicit DESC is forwarded",
			req: &racing.ListRacesRequest{
				Filter:    &racing.ListRacesRequestFilter{},
				SortBy:    racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
				Direction: racing.SortDirection_SORT_DIRECTION_DESCENDING,
			},
			expectVis:  racing.Visibility_VISIBILITY_UNSPECIFIED,
			expectSort: racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
			expectDir:  racing.SortDirection_SORT_DIRECTION_DESCENDING,
			repoReply:  []*racing.Race{r3, r2, r1},
			wantIDs:    []int64{3, 2, 1},
		},
	}

	for i := range tests {
		tc := tests[i] // avoid range var capture
		t.Run(tc.name, func(t *testing.T) {
			m := new(dbmocks.RacesRepo)

			// Expect service to forward ctx, filter (we check Visibility), sortBy, direction
			m.On(
				"List",
				mock.Anything, // ctx
				mock.MatchedBy(func(f *racing.ListRacesRequestFilter) bool {
					return f.GetVisibility() == tc.expectVis
				}),
				tc.expectSort,
				tc.expectDir,
			).Return(tc.repoReply, nil).Once()

			svc := &racingService{racesRepo: m}

			resp, err := svc.ListRaces(context.Background(), tc.req)
			assert.NoError(t, err)

			var got []int64
			for _, r := range resp.GetRaces() {
				got = append(got, r.GetId())
				if tc.expectVis == racing.Visibility_VISIBILITY_VISIBLE_ONLY {
					assert.True(t, r.GetVisible(), "expected only visible races")
				}
			}
			assert.ElementsMatch(t, tc.wantIDs, got)

			m.AssertExpectations(t)
		})
	}
}
