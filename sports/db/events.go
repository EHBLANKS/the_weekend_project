package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sports "sports/proto/sports"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type eventsRepo struct{ db *sql.DB }

func NewEventsRepo(db *sql.DB) EventsRepo { return &eventsRepo{db: db} }

// List implements offset/limit, DB-derived status, and whitelisted ORDER BY.
func (r *eventsRepo) List(
	ctx context.Context,
	filter *sports.ListEventsRequestFilter,
	sortBy sports.SortBy,
	dir sports.SortDirection,
	limit, offset int32,
) ([]*sports.Event, error) {
	now := time.Now().UTC()
	query := getEventQueries()[eventsList]

	// Filters
	query, filterArgs := r.applyFilter(query, filter)

	// ORDER BY
	query = r.applySort(query, sortBy, dir)

	// LIMIT/OFFSET
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query += " LIMIT ? OFFSET ?"

	// Bind: NOW for CASE, then filter args, then limit/offset
	args := make([]any, 0, 1+len(filterArgs)+2)
	args = append(args, now)
	args = append(args, filterArgs...)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()
	return r.scanEvents(rows)
}

func (r *eventsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []any) {
	var clauses []string
	var args []any
	if filter == nil {
		return query, args
	}
	if len(filter.CompetitionIds) > 0 {
		ph := strings.Repeat("?,", len(filter.CompetitionIds)-1) + "?"
		clauses = append(clauses, "competition_id IN ("+ph+")")
		for _, id := range filter.CompetitionIds {
			args = append(args, id)
		}
	}
	if filter.GetVisibility() == sports.Visibility_VISIBILITY_VISIBLE_ONLY {
		clauses = append(clauses, "visible = 1")
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	return query, args
}

func (r *eventsRepo) applySort(query string, sortBy sports.SortBy, dir sports.SortDirection) string {
	if sortBy == sports.SortBy_SORT_BY_UNSPECIFIED {
		sortBy = sports.SortBy_SORT_BY_ADVERTISED_START_TIME
	}
	if dir == sports.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		dir = sports.SortDirection_SORT_DIRECTION_ASCENDING
	}
	if sortBy == sports.SortBy_SORT_BY_ADVERTISED_START_TIME {
		order := "ASC"
		if dir == sports.SortDirection_SORT_DIRECTION_DESCENDING {
			order = "DESC"
		}
		query += " ORDER BY advertised_start_time " + order + ", id " + order
	}
	return query
}

func (r *eventsRepo) scanEvents(rows *sql.Rows) ([]*sports.Event, error) {
	var out []*sports.Event
	for rows.Next() {
		var (
			id, compID int64
			name       string
			visible    bool
			start      time.Time
			statusInt  int64
		)
		if err := rows.Scan(&id, &compID, &name, &visible, &start, &statusInt); err != nil {
			return nil, err
		}
		status := sports.EventStatus_EVENT_STATUS_UNSPECIFIED
		switch statusInt {
		case 1:
			status = sports.EventStatus_EVENT_STATUS_OPEN
		case 2:
			status = sports.EventStatus_EVENT_STATUS_CLOSED
		}
		out = append(out, &sports.Event{
			Id:                  id,
			CompetitionId:       compID,
			Name:                name,
			Visible:             visible,
			AdvertisedStartTime: timestamppb.New(start),
			Status:              status,
		})
	}
	return out, nil
}
