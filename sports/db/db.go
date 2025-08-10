package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}

func EnsureSchemaAndSeed(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY,
  competition_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  visible BOOLEAN NOT NULL,
  advertised_start_time TIMESTAMP NOT NULL
);
`)
	if err != nil {
		return err
	}

	// minimal seed if table empty
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&n)
	if n > 0 {
		return nil
	}

	now := time.Now().UTC()
	ins := `INSERT INTO events (id,competition_id,name,visible,advertised_start_time) VALUES (?,?,?,?,?)`
	_, _ = db.Exec(ins, 201, 10, "Chelsea vs Spurs", true, now.Add(15*time.Minute))
	_, _ = db.Exec(ins, 202, 10, "Arsenal vs City", true, now.Add(90*time.Minute))
	_, _ = db.Exec(ins, 203, 11, "Yankees @ Red Sox", false, now.Add(-30*time.Minute))
	return nil
}
