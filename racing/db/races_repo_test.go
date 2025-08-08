package db

import (
	"context"
	"database/sql"
	"racing/proto/racing"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestSQLite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:repotest.db?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	const schema = `
  CREATE TABLE races (
    id INTEGER PRIMARY KEY,
    meeting_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    number INTEGER NOT NULL,
    visible BOOLEAN NOT NULL,
    advertised_start_time TIMESTAMP NOT NULL
  );`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func seedRepo(t *testing.T, db *sql.DB) {
	t.Helper()
	now := time.Now().UTC()
	const ins = `INSERT INTO races (id,meeting_id,name,number,visible,advertised_start_time)
	             VALUES (?,?,?,?,?,?)`
	for _, r := range []struct {
		id, meet, num int64
		name          string
		vis           bool
	}{
		{1, 3, 1, "Alpha", true},
		{2, 3, 2, "Bravo", false},
		{3, 5, 3, "Charlie", true},
	} {
		if _, err := db.Exec(ins, r.id, r.meet, r.name, r.num, r.vis, now); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func TestRacesRepo_List_Visibility(t *testing.T) {
	sqlDB := newTestSQLite(t)
	seedRepo(t, sqlDB)
	repo := &racesRepo{db: sqlDB}

	tests := []struct {
		name       string
		filter     *racing.ListRacesRequestFilter
		wantMin    int  // minimum expected rows
		allVisible bool // if true, assert every row is visible
	}{
		{
			name:    "unspecified -> all",
			filter:  &racing.ListRacesRequestFilter{},
			wantMin: 3,
		},
		{
			name:       "visible only",
			filter:     &racing.ListRacesRequestFilter{Visibility: racing.Visibility_VISIBILITY_VISIBLE_ONLY},
			wantMin:    1,
			allVisible: true,
		},
		{
			name:       "meeting_ids + visible only",
			filter:     &racing.ListRacesRequestFilter{MeetingIds: []int64{3, 5}, Visibility: racing.Visibility_VISIBILITY_VISIBLE_ONLY},
			wantMin:    2, // Alpha(3, vis) + Charlie(5, vis)
			allVisible: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// If your List signature is List(ctx, filter)
			got, err := repo.List(context.Background(), tc.filter)
			if err != nil {
				t.Fatalf("List error: %v", err)
			}
			if len(got) < tc.wantMin {
				t.Fatalf("got %d races, want >= %d", len(got), tc.wantMin)
			}
			if tc.allVisible {
				for _, r := range got {
					if !r.GetVisible() {
						t.Fatalf("expected visible=true, got id=%d visible=false", r.GetId())
					}
				}
			}
		})
	}
}
