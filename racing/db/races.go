package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(ctx context.Context, filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
}

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

func (r *racesRepo) List(ctx context.Context, filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		query string
		args  []any
	)

	query = getRaceQueries()[racesList]
	query, args = r.applyFilter(query, filter)

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
		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	// PR1: visibility filter
	switch filter.GetVisibility() {
	case racing.Visibility_VISIBILITY_VISIBLE_ONLY:
		clauses = append(clauses, "visible = 1")
		// VISIBILITY_ANY / UNSPECIFIED => no clause
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	return query, args
}

func (m *racesRepo) scanRaces(rows *sql.Rows) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var (
			race            racing.Race
			advertisedStart time.Time
		)

		if err := rows.Scan(
			&race.Id,
			&race.MeetingId,
			&race.Name,
			&race.Number,
			&race.Visible,
			&advertisedStart,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}

		race.AdvertisedStartTime = timestamppb.New(advertisedStart)
		races = append(races, &race)
	}
	return races, nil
}
