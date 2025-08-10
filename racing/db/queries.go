package db

const (
	racesList = "list"
)

func getRaceQueries() map[string]string {
	return map[string]string{
		racesList: `
      SELECT
        id,
        meeting_id,
        name,
        number,
        visible,
        advertised_start_time,
        CASE
          WHEN advertised_start_time < ? THEN 2 -- RACE_STATUS_CLOSED
          ELSE 1                                -- RACE_STATUS_OPEN
        END AS status
      FROM races
      `,
	}
}
