package db

import (
	"context"
	sports "sports/proto/sports"
)

//go:generate mockery --name EventsRepo --dir . --output ./mocks --outpkg dbmocks --filename sports_repo_mock.go
type EventsRepo interface {
	List(
		ctx context.Context,
		filter *sports.ListEventsRequestFilter,
		sortBy sports.SortBy,
		dir sports.SortDirection,
		limit, offset int32,
	) ([]*sports.Event, error)
}
