package db

import (
	"context"
	"testing"
	"time"

	racing "racing/proto/racing"

	_ "modernc.org/sqlite"
)

func TestRepo_Get_FindsRowAndStatus(t *testing.T) {
	db := newTestDB(t)
	// reuse your existing newTestDB/seed helpers if present
	now := time.Now().UTC()
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS races(
        id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, number INTEGER, visible BOOLEAN, advertised_start_time TIMESTAMP
    )`)
	ins := `INSERT INTO races (id,meeting_id,name,number,visible,advertised_start_time) VALUES (?,?,?,?,?,?)`
	_, _ = db.Exec(ins, 75, 1, "Test", 7, 1, now.Add(time.Minute))

	repo := &racesRepo{db: db}
	got, err := repo.Get(context.Background(), 75)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GetId() != 75 {
		t.Fatalf("want id=75, got %d", got.GetId())
	}
	if got.GetStatus() != racing.RaceStatus_RACE_STATUS_OPEN {
		t.Fatalf("want OPEN, got %v", got.GetStatus())
	}
}

func TestRepo_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := &racesRepo{db: db}
	_, err := repo.Get(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error for missing row")
	}
}
