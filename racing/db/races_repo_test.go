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

	seedData := []struct {
		id, meet, num int64
		name          string
		vis           bool
	}{
		{id: 1, meet: 3, num: 1, name: "Alpha", vis: true},
		{id: 2, meet: 3, num: 2, name: "Bravo", vis: false},
		{id: 3, meet: 5, num: 3, name: "Charlie", vis: true},
	}

	for _, r := range seedData {
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.List(context.Background(), tc.filter, racing.SortBy_SORT_BY_ADVERTISED_START_TIME, racing.SortDirection_SORT_DIRECTION_ASCENDING)
			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}

			if len(got) < tc.wantMin {
				t.Errorf("got %d rows, want at least %d", len(got), tc.wantMin)
			}

			if tc.allVisible {
				for _, r := range got {
					if !r.Visible {
						t.Errorf("expected all visible, got race ID %d with visible=%v", r.Id, r.Visible)
					}
				}
			}
		})
	}
}
