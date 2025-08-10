# Entain BE Technical Test — Implementation (PR1–PR5)

This repo contains a small multi-service Go backend demonstrating gRPC + gRPC-Gateway, Protobuf APIs, SQLite persistence, and practical API design.

It implements the five tasks from the assignment:

- **PR1**: Visibility filter on ListRaces
- **PR2**: Stable ordering (with enums)
- **PR3**: Derived status (OPEN/CLOSED) + offset/limit pagination
- **PR4**: GetRace by ID
- **PR5**: New Sports service mirroring Racing (ListEvents)

---

## Structure

```
.
├── README.md
├── go.work                    # workspace for multi-module build
├── api/                       # REST gateway (gRPC-Gateway)
│   ├── main.go
│   └── proto/
│       ├── google/api/{annotations.proto,http.proto}
│       ├── racing/…           # generated stubs (incl. HTTP handlers)
│       └── sports/…           # generated stubs (incl. HTTP handlers)
├── racing/                    # Racing microservice
│   ├── main.go
│   ├── db/
│   │   ├── races.go           # repo + SQL + status derivation
│   │   ├── queries.go
│   │   └── racing.db          # sample DB
│   ├── proto/racing/…         # proto + generated stubs
│   └── service/…              # business logic (sorting, paging)
└── sports/                    # Sports microservice
    ├── main.go
    ├── db/
    │   ├── db.go              # open + EnsureSchemaAndSeed()
    │   └── events.go          # repo + SQL + status derivation
    └── proto/sports/…         # proto + generated stubs
```

Workspace (root `go.work`) includes `./api`, `./racing`, `./sports`.

---

## Prerequisites

- **Go 1.23+** (workspace may pin toolchain go1.24.x)
- **protoc** (Protocol Buffers compiler)
- **Generators**:

  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
  ```

- **(Dev) mockery** for mocks:

  ```bash
  go install github.com/vektra/mockery/v2@latest
  ```

---

## Build & Run

### 1. Generate stubs

```bash
(cd racing && go generate ./...)
(cd sports && go generate ./...)
(cd api && go generate ./...)
```

Generation writes `.pb.go` files next to `.proto` (paths=source_relative).
The api module also generates the HTTP gateway handlers.

### 2. Run services (in three terminals)

```bash
# Racing gRPC (localhost:9000)
cd racing && go build && ./racing

# Sports gRPC (localhost:9100)
cd sports && go build && ./sports

# API Gateway HTTP (localhost:8000)
cd api && go build && ./api
```

### 3. Smoke tests (examples with jq)

```bash
# Racing — list visible only, ASC, page 1 (3 per page)
curl -s localhost:8000/v1/list-races \
  -H 'content-type: application/json' \
  -d '{"filter":{"visibility":"VISIBILITY_VISIBLE_ONLY"},"sortBy":"SORT_BY_ADVERTISED_START_TIME","direction":"SORT_DIRECTION_ASCENDING","limit":3,"offset":0}' \
  | jq '{ids:[.races[].id], next:.nextOffset}'

# Racing — page 2
curl -s localhost:8000/v1/list-races \
  -H 'content-type: application/json' \
  -d '{"filter":{"visibility":"VISIBILITY_VISIBLE_ONLY"},"sortBy":"SORT_BY_ADVERTISED_START_TIME","direction":"SORT_DIRECTION_ASCENDING","limit":3,"offset":3}' \
  | jq '{ids:[.races[].id], next:.nextOffset}'

# Racing — get single
curl -s localhost:8000/v1/races/75 | jq

# Sports — list all, ASC
curl -s localhost:8000/v1/sports/list-events \
  -H 'content-type: application/json' \
  -d '{"filter":{},"sortBy":"SORT_BY_ADVERTISED_START_TIME","direction":"SORT_DIRECTION_ASCENDING","limit":5,"offset":0}' \
  | jq '{ids:[.events[].id], next:.nextOffset}'
```

---

## API

### Common types

- **SortBy**
  - `SORT_BY_ADVERTISED_START_TIME` (default)
- **SortDirection**
  - `SORT_DIRECTION_ASCENDING` (default)
  - `SORT_DIRECTION_DESCENDING`
- **Visibility**
  - `VISIBILITY_UNSPECIFIED` (default = no filter)
  - `VISIBILITY_VISIBLE_ONLY`
- **Status derivation** (DB-level)
  - `< now (UTC)` → `CLOSED`
  - `≥ now (UTC)` → `OPEN`

### Racing

#### ListRaces — `POST /v1/list-races`

**Request:**

```json
{
  "filter": {
    "meetingIds": [6, 10],
    "visibility": "VISIBILITY_VISIBLE_ONLY"
  },
  "sortBy": "SORT_BY_ADVERTISED_START_TIME",
  "direction": "SORT_DIRECTION_ASCENDING",
  "limit": 50,
  "offset": 0
}
```

**Response:**

```json
{
  "races": [
    {
      "id": "75",
      "meetingId": "10",
      "name": "Hawaii griffins",
      "number": "11",
      "visible": false,
      "advertisedStartTime": "2021-02-28T07:00:50Z",
      "status": "RACE_STATUS_CLOSED"
    }
  ],
  "nextOffset": 50
}
```

**Notes:**

- Stable ordering: `ORDER BY advertised_start_time, id` (id is the tiebreaker).
- Pagination: offset/limit. `nextOffset = offset + limit` only when `len(races) == limit`, else `0`.

#### GetRace — `GET /v1/races/{id}`

### Sports

#### ListEvents — `POST /v1/sports/list-events`

**Request:**

```json
{
  "filter": {
    "competitionIds": [10, 11],
    "visibility": "VISIBILITY_VISIBLE_ONLY"
  },
  "sortBy": "SORT_BY_ADVERTISED_START_TIME",
  "direction": "SORT_DIRECTION_ASCENDING",
  "limit": 5,
  "offset": 0
}
```

**Response:**

```json
{
  "events": [
    {
      "id": "201",
      "competitionId": "10",
      "name": "Chelsea vs Spurs",
      "visible": true,
      "advertisedStartTime": "2021-02-28T07:15:00Z",
      "status": "EVENT_STATUS_OPEN"
    }
  ],
  "nextOffset": 5
}
```

---

## Design Notes

- **PR1** – **Visibility filter**: Implemented at SQL level (`visible = 1`), exposed via `ListRacesRequest.filter.visibility`.
- **PR2** – **Ordering**: Whitelisted, stable ordering: `advertised_start_time` then `id`. Client controls ASC/DESC.
- **PR3** – **Status + Pagination**: `status` derived in SQL with a `CASE` using a bound `now` param (UTC). Pagination uses offset/limit with `nextOffset` semantics:
  - Defaults: `limit=50` (max 200), `offset=0`.
  - `nextOffset = offset + limit` only if `len(results) == limit`; otherwise `0`.
- **PR4** – **Get by ID**: `GetRace(id)` with validation and NotFound mapping.
- **PR5** – **Sports service**: Separate microservice mirroring the Racing API (`ListEvents`), with visibility & competition filters, sorting, and pagination. REST exposed via the same API Gateway.
- **Security & robustness**: Parameterized SQL; whitelisted ORDER BY; input clamps; gRPC status → HTTP mapping via gRPC-Gateway.
- **Why DB-level for status & ordering**: Offloads work to the DB (indexable), reduces memory pressure, and gives consistent semantics.

---

## Running Tests

### Racing

```bash
cd racing
go test ./... -v
```

### Sports

```bash
cd sports
go test ./... -v
```

- **Repo (integration) tests**: SQLite queries, ordering, visibility, paging, and status derivation.
- **Service (unit) tests**: Mocks verify forwarding of filters, sort, pagination, and `nextOffset` behavior; Racing also tests `GetRace` success/invalid/not found.

### Regenerate mocks with

```bash
cd racing/db && mockery --name RacesRepo --dir . --output ./mocks --outpkg dbmocks --filename races_repo_mock.go
cd sports/db && mockery --name EventsRepo --dir . --output ./mocks --outpkg dbmocks --filename events_repo_mock.go
```

---

## Dev Tips

- Keep `go.work` at repo root so tools understand the multi-module layout:

  ```go
  go 1.23.0
  toolchain go1.24.x

  use (
      ./api
      ./racing
      ./sports
  )
  ```

- **Formatting & linting**:

  ```bash
  go fmt ./...
  go vet ./...
  go mod tidy
  ```

- **Suggested DB indexes** (if this were production):

  ```sql
  CREATE INDEX IF NOT EXISTS idx_races_start_time ON races(advertised_start_time);
  CREATE INDEX IF NOT EXISTS idx_races_meeting_id ON races(meeting_id);
  CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(advertised_start_time);
  CREATE INDEX IF NOT EXISTS idx_events_competition ON events(competition_id);
  ```

---

## PR Summary (for reviewers)

- **PR1**: Add Visibility enum; SQL filter; tests (repo + service).
- **PR2**: Add SortBy/SortDirection; stable id tiebreak; tests ASC/DESC.
- **PR3**: `status` via SQL; add limit/offset & `nextOffset`; tests for paging & status.
- **PR4**: `GetRace` RPC + REST; repo Get; service validation & errors; tests.
- **PR5**: New sports module; `ListEvents` with filters/sort/paging/status; gateway wiring; tests.
