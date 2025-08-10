package service

import (
	"context"
	"racing/models"

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
	// Defaults for sort
	sortBy := in.GetSortBy()
	if sortBy == racing.SortBy_SORT_BY_UNSPECIFIED {
		sortBy = racing.SortBy_SORT_BY_ADVERTISED_START_TIME
	}
	dir := in.GetDirection()
	if dir == racing.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		dir = racing.SortDirection_SORT_DIRECTION_ASCENDING
	}

	// Clamp pagination
	limit := in.GetLimit()
	if limit <= 0 {
		limit = models.DefaultPageSize
	}
	if limit > models.MaxPageSize {
		limit = models.MaxPageSize
	}
	offset := in.GetOffset()
	if offset < 0 {
		offset = 0
	}

	races, err := s.racesRepo.List(ctx, in.GetFilter(), sortBy, dir, limit, offset)
	if err != nil {
		return nil, err
	}

	// naive next_offset (0 when last page)
	next := int32(0)
	if int32(len(races)) == limit {
		next = offset + limit
	}

	return &racing.ListRacesResponse{
		Races:      races,
		NextOffset: next,
	}, nil
}
