// Package db implements the data access layer for racing.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	racing "racing/proto/racing"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error
	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})
	return err
}

// List returns races filtered and ordered per request.
func (r *racesRepo) List(
	ctx context.Context,
	filter *racing.ListRacesRequestFilter,
	sortBy racing.SortBy,
	direction racing.SortDirection,
) ([]*racing.Race, error) {
	var (
		query string
		args  []any
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)
	query = r.applySort(query, sortBy, direction) // <-- ORDER BY here

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query races: %w", err)
	}
	defer rows.Close()

	races, err := r.scanRaces(rows)
	if err != nil {
		return nil, fmt.Errorf("scan races: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return races, nil
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []any) {
	var (
		clauses []string
		args    []any
	)

	if filter == nil {
		return query, args
	}

	// meeting_ids IN (...)
	if len(filter.MeetingIds) > 0 {
		placeholders := strings.Repeat("?,", len(filter.MeetingIds)-1) + "?"
		clauses = append(clauses, "meeting_id IN ("+placeholders+")")
		for _, id := range filter.MeetingIds {
			args = append(args, id)
		}
	}

	// visibility
	switch filter.GetVisibility() {
	case racing.Visibility_VISIBILITY_VISIBLE_ONLY:
		clauses = append(clauses, "visible = 1")
		// VISIBILITY_UNSPECIFIED / VISIBILITY_ANY => no clause
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	return query, args
}

func (r *racesRepo) applySort(query string, sortBy racing.SortBy, dir racing.SortDirection) string {
	// default defensively (service already sets defaults)
	if sortBy == racing.SortBy_SORT_BY_UNSPECIFIED {
		sortBy = racing.SortBy_SORT_BY_ADVERTISED_START_TIME
	}
	if dir == racing.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		dir = racing.SortDirection_SORT_DIRECTION_ASCENDING
	}

	if sortBy == racing.SortBy_SORT_BY_ADVERTISED_START_TIME {
		order := "ASC"
		if dir == racing.SortDirection_SORT_DIRECTION_DESCENDING {
			order = "DESC"
		}
		// tie-break on id for deterministic order
		query += " ORDER BY advertised_start_time " + order + ", id ASC"
	}
	return query
}

// scanRaces maps rows to proto messages.
func (r *racesRepo) scanRaces(rows *sql.Rows) ([]*racing.Race, error) {
	var out []*racing.Race
	for rows.Next() {
		var (
			id, meetingID int64
			name          string
			number        int64
			visible       bool
			start         time.Time
		)
		if err := rows.Scan(&id, &meetingID, &name, &number, &visible, &start); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		out = append(out, &racing.Race{
			Id:                  id,
			MeetingId:           meetingID,
			Name:                name,
			Number:              number,
			Visible:             visible,
			AdvertisedStartTime: timestamppb.New(start),
		})
	}
	return out, nil
}
