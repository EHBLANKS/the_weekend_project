package db

import (
	"context"

	racing "racing/proto/racing"
)

//go:generate mockery --name RacesRepo --dir . --output ./mocks --outpkg dbmocks --filename races_repo_mock.go
type RacesRepo interface {
	List(ctx context.Context,
		filter *racing.ListRacesRequestFilter,
		sortBy racing.SortBy,
		direction racing.SortDirection,
	) ([]*racing.Race, error)
}
