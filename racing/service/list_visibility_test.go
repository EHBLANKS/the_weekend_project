package service

import (
	"context"
	"racing/proto/racing"
	"racing/service/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListRaces(t *testing.T) {
	cases := []struct {
		name       string
		visibility racing.Visibility
		want       int
	}{
		{"visible-only", racing.Visibility_VISIBILITY_VISIBLE_ONLY, 1},
		{"all", racing.Visibility_VISIBILITY_UNSPECIFIED, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo := new(mocks.RacesRepo)
			repo.On("List", ctx, mock.Anything).
				Return([]*racing.Race{{Visible: true}, {Visible: false}}, nil)

			svc := NewRacingService(repo)
			resp, err := svc.ListRaces(ctx, &racing.ListRacesRequest{
				Filter: &racing.ListRacesRequestFilter{Visibility: tc.visibility},
			})

			require.NoError(t, err)
			require.Len(t, resp.Races, tc.want)
			repo.AssertExpectations(t)
		})
	}
}
