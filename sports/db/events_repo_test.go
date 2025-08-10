package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sportspb "sports/proto/sports"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:sports-test.db?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := EnsureSchemaAndSeed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

func seedEventsExact(t *testing.T, db *sql.DB) {
	t.Helper()
	_, _ = db.Exec(`DELETE FROM events`)
	now := time.Date(2021, 2, 28, 7, 0, 0, 0, time.UTC)

	ins := `INSERT INTO events (id, competition_id, name, visible, advertised_start_time) VALUES (?,?,?,?,?)`
	// strictly increasing times so sort/order expectations are deterministic
	_, _ = db.Exec(ins, 201, 10, "A", true, now.Add(-20*time.Minute))  // CLOSED
	_, _ = db.Exec(ins, 202, 10, "B", false, now.Add(-10*time.Minute)) // CLOSED
	_, _ = db.Exec(ins, 203, 11, "C", true, now.Add(-1*time.Minute))   // CLOSED
	_, _ = db.Exec(ins, 204, 11, "D", true, now.Add(+12*time.Minute))  // OPEN
	_, _ = db.Exec(ins, 205, 12, "E", false, now.Add(+49*time.Minute)) // OPEN
}

func TestEventsRepo_List_ASC_OffsetPaging(t *testing.T) {
	sqlDB := newTestDB(t)
	seedEventsExact(t, sqlDB)
	repo := NewEventsRepo(sqlDB)

	// page 1
	ev1, err := repo.List(context.Background(),
		&sportspb.ListEventsRequestFilter{},
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_ASCENDING,
		2, 0,
	)
	if err != nil {
		t.Fatalf("List p1: %v", err)
	}
	if len(ev1) != 2 {
		t.Fatalf("want 2, got %d", len(ev1))
	}
	if ev1[0].GetId() != 201 || ev1[1].GetId() != 202 {
		t.Fatalf("ids p1 = %v,%v", ev1[0].GetId(), ev1[1].GetId())
	}

	// page 2
	ev2, err := repo.List(context.Background(),
		&sportspb.ListEventsRequestFilter{},
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_ASCENDING,
		2, 2,
	)
	if err != nil {
		t.Fatalf("List p2: %v", err)
	}
	if len(ev2) != 2 {
		t.Fatalf("want 2, got %d", len(ev2))
	}
	if ev2[0].GetId() != 203 || ev2[1].GetId() != 204 {
		t.Fatalf("ids p2 = %v,%v", ev2[0].GetId(), ev2[1].GetId())
	}
}

func TestEventsRepo_List_DESC_Order(t *testing.T) {
	sqlDB := newTestDB(t)
	seedEventsExact(t, sqlDB)
	repo := NewEventsRepo(sqlDB)

	ev, err := repo.List(context.Background(),
		&sportspb.ListEventsRequestFilter{},
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_DESCENDING,
		3, 0,
	)
	if err != nil {
		t.Fatalf("List desc: %v", err)
	}
	if len(ev) != 3 {
		t.Fatalf("want 3, got %d", len(ev))
	}
	// latest first
	if ev[0].GetId() != 205 || ev[1].GetId() != 204 || ev[2].GetId() != 203 {
		t.Fatalf("ids desc = %v,%v,%v", ev[0].GetId(), ev[1].GetId(), ev[2].GetId())
	}
}

func TestEventsRepo_List_VisibleOnly_AndStatusDerived(t *testing.T) {
	sqlDB := newTestDB(t)
	seedEventsExact(t, sqlDB)
	repo := NewEventsRepo(sqlDB)

	ev, err := repo.List(context.Background(),
		&sportspb.ListEventsRequestFilter{Visibility: sportspb.Visibility_VISIBILITY_VISIBLE_ONLY},
		sportspb.SortBy_SORT_BY_ADVERTISED_START_TIME,
		sportspb.SortDirection_SORT_DIRECTION_ASCENDING,
		10, 0,
	)
	if err != nil {
		t.Fatalf("List visible: %v", err)
	}
	if len(ev) == 0 {
		t.Fatalf("want > 0")
	}
	for _, e := range ev {
		if !e.GetVisible() {
			t.Fatalf("expected only visible; id=%d visible=%v", e.GetId(), e.GetVisible())
		}
		if e.GetStatus() == sportspb.EventStatus_EVENT_STATUS_UNSPECIFIED {
			t.Fatalf("status not derived for id=%d", e.GetId())
		}
	}
}
