package db

const (
	racesList = "list"
	racesGet  = "get"
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
		racesGet: `
      SELECT
        id, meeting_id, name, number, visible, advertised_start_time,
        CASE WHEN advertised_start_time < ? THEN 2 ELSE 1 END AS status
      FROM races
      WHERE id = ?
      LIMIT 1
      `,
	}
}
