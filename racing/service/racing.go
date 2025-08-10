package service

import (
	"context"

	db "racing/db"
	racing "racing/proto/racing"
)

type Racing interface {
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
}

type racingService struct {
	racesRepo db.RacesRepo
}

func NewRacingService(racesRepo db.RacesRepo) Racing {
	return &racingService{racesRepo: racesRepo}
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	sortBy := in.GetSortBy()
	if sortBy == racing.SortBy_SORT_BY_UNSPECIFIED {
		sortBy = racing.SortBy_SORT_BY_ADVERTISED_START_TIME
	}
	dir := in.GetDirection()
	if dir == racing.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		dir = racing.SortDirection_SORT_DIRECTION_ASCENDING
	}

	races, err := s.racesRepo.List(ctx, in.GetFilter(), sortBy, dir)
	if err != nil {
		return nil, err
	}
	return &racing.ListRacesResponse{Races: races}, nil
}
