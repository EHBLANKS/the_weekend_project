package service

import (
	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
)

type Racing interface {
	// ListRaces will return a collection of races.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
}

// racingService implements the Racing interface.
type racingService struct {
	racesRepo db.RacesRepo
}

// NewRacingService instantiates and returns a new racingService.
func NewRacingService(racesRepo db.RacesRepo) Racing {
	return &racingService{racesRepo}
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	races, err := s.racesRepo.List(ctx, in.GetFilter())
	if err != nil {
		return nil, err
	}

	// Apply visibilitiy for filter for PR1
	switch in.GetFilter().GetVisibility() {
	case racing.Visibility_VISIBILITY_VISIBLE_ONLY:
		out := make([]*racing.Race, 0, len(races))
		for _, r := range races {
			if r.GetVisible() {
				out = append(out, r)
			}
		}
		races = out
		// NOTE: VISIBILITY_ANY || VISIBILITY_UNSPECIFIED, no filtering so far
	}

	return &racing.ListRacesResponse{Races: races}, nil
}
