package service

import (
	"context"

	db "sports/db"
	sports "sports/proto/sports"
)

type Sports interface {
	ListEvents(context.Context, *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
}

type sportsService struct{ repo db.EventsRepo }

func NewSportsService(repo db.EventsRepo) sports.SportsServiceServer {
	return &sportsService{repo: repo}
}

func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	sortBy := in.GetSortBy()
	if sortBy == sports.SortBy_SORT_BY_UNSPECIFIED {
		sortBy = sports.SortBy_SORT_BY_ADVERTISED_START_TIME
	}
	dir := in.GetDirection()
	if dir == sports.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		dir = sports.SortDirection_SORT_DIRECTION_ASCENDING
	}
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := in.GetOffset()
	if offset < 0 {
		offset = 0
	}

	events, err := s.repo.List(ctx, in.GetFilter(), sortBy, dir, limit, offset)
	if err != nil {
		return nil, err
	}

	next := int32(0)
	if int32(len(events)) == limit {
		next = offset + limit
	}
	return &sports.ListEventsResponse{Events: events, NextOffset: next}, nil
}
