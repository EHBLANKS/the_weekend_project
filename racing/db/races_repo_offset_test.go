package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	racing "racing/proto/racing"

	_ "modernc.org/sqlite" // CGO-free sqlite; driver name: "sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:offpag.db?mode=memory&cache=shared")
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

func seedRaces(t *testing.T, db *sql.DB) {
	t.Helper()
	now := time.Date(2021, 2, 28, 7, 0, 0, 0, time.UTC)

	ins := `INSERT INTO races (id, meeting_id, name, number, visible, advertised_start_time)
          VALUES (?,?,?,?,?,?)`

	// times strictly increasing; include some visible=false
	rows := []struct {
		id, meet, num int64
		name          string
		vis           bool
		start         time.Time
	}{
		{23, 6, 10, "A", false, now.Add(-20 * time.Minute)},
		{75, 10, 11, "B", false, now.Add(-10 * time.Minute)},
		{76, 7, 9, "C", true, now.Add(-1 * time.Minute)},
		{80, 8, 7, "D", true, now.Add(+12 * time.Minute)},
		{83, 8, 10, "E", true, now.Add(+49 * time.Minute)},
	}
	for _, r := range rows {
		if _, err := db.Exec(ins, r.id, r.meet, r.name, r.num, r.vis, r.start); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func TestRepo_List_Offset_ASC(t *testing.T) {
	sqlDB := newTestDB(t)
	seedRaces(t, sqlDB)
	repo := &racesRepo{db: sqlDB}

	// page 1: limit=2, offset=0
	got1, err := repo.List(context.Background(),
		&racing.ListRacesRequestFilter{},
		racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
		racing.SortDirection_SORT_DIRECTION_ASCENDING,
		2, 0,
	)
	if err != nil {
		t.Fatalf("List p1: %v", err)
	}
	if len(got1) != 2 {
		t.Fatalf("want 2, got %d", len(got1))
	}
	if got1[0].GetId() != 23 || got1[1].GetId() != 75 {
		t.Fatalf("unexpected ids p1: %d,%d", got1[0].GetId(), got1[1].GetId())
	}
	// Status must be populated (DB-derived)
	for _, r := range got1 {
		if r.GetStatus() == racing.RaceStatus_RACE_STATUS_UNSPECIFIED {
			t.Fatalf("status unspecified for id=%d", r.GetId())
		}
	}

	// page 2: limit=2, offset=2
	got2, err := repo.List(context.Background(),
		&racing.ListRacesRequestFilter{},
		racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
		racing.SortDirection_SORT_DIRECTION_ASCENDING,
		2, 2,
	)
	if err != nil {
		t.Fatalf("List p2: %v", err)
	}
	if len(got2) != 2 {
		t.Fatalf("want 2, got %d", len(got2))
	}
	if got2[0].GetId() != 76 || got2[1].GetId() != 80 {
		t.Fatalf("unexpected ids p2: %d,%d", got2[0].GetId(), got2[1].GetId())
	}

	// page 3: limit=2, offset=4 (tail)
	got3, err := repo.List(context.Background(),
		&racing.ListRacesRequestFilter{},
		racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
		racing.SortDirection_SORT_DIRECTION_ASCENDING,
		2, 4,
	)
	if err != nil {
		t.Fatalf("List p3: %v", err)
	}
	if len(got3) != 1 {
		t.Fatalf("want 1, got %d", len(got3))
	}
	if got3[0].GetId() != 83 {
		t.Fatalf("unexpected id p3: %d", got3[0].GetId())
	}
}

func TestRepo_List_Offset_DESC(t *testing.T) {
	sqlDB := newTestDB(t)
	seedRaces(t, sqlDB)
	repo := &racesRepo{db: sqlDB}

	got, err := repo.List(context.Background(),
		&racing.ListRacesRequestFilter{},
		racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
		racing.SortDirection_SORT_DIRECTION_DESCENDING,
		3, 0,
	)
	if err != nil {
		t.Fatalf("List desc: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
	// DESC → highest time first (E, D, C)
	if got[0].GetId() != 83 || got[1].GetId() != 80 || got[2].GetId() != 76 {
		t.Fatalf("unexpected ids desc: %d,%d,%d", got[0].GetId(), got[1].GetId(), got[2].GetId())
	}
}

func TestRepo_List_VisibleOnly_WithOffset(t *testing.T) {
	sqlDB := newTestDB(t)
	seedRaces(t, sqlDB)
	repo := &racesRepo{db: sqlDB}

	got, err := repo.List(context.Background(),
		&racing.ListRacesRequestFilter{Visibility: racing.Visibility_VISIBILITY_VISIBLE_ONLY},
		racing.SortBy_SORT_BY_ADVERTISED_START_TIME,
		racing.SortDirection_SORT_DIRECTION_ASCENDING,
		10, 0,
	)
	if err != nil {
		t.Fatalf("List visible: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("want >0 rows")
	}
	for _, r := range got {
		if !r.GetVisible() {
			t.Fatalf("expected only visible races; got id=%d visible=%v", r.GetId(), r.GetVisible())
		}
		if r.GetStatus() == racing.RaceStatus_RACE_STATUS_UNSPECIFIED {
			t.Fatalf("status unspecified for id=%d", r.GetId())
		}
	}
}
