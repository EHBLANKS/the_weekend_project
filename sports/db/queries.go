package db

const (
	eventsList = "eventsList"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventsList: `
      SELECT
        id,
        competition_id,
        name,
        visible,
        advertised_start_time,
        CASE
          WHEN advertised_start_time < ? THEN 2 -- EVENT_STATUS_CLOSED
          ELSE 1                                -- EVENT_STATUS_OPEN
        END AS status
      FROM events
      `,
	}
}
