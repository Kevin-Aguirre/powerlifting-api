
# RESTful Go API for OpenPowerlifting Data
A RESTful API for open-source powerlifting data from the [OpenPowerlifting](https://www.openpowerlifting.org/) project.

## Getting Started

### Prerequisites
- Go 1.23+
- Git (to clone the OPL dataset)

### 1. Clone this repo
```bash
git clone https://github.com/Kevin-Aguirre/powerlifting-api.git
cd powerlifting-api
```

### 2. Clone the OpenPowerlifting data repo
The API reads meet data from the [OPL dataset](https://gitlab.com/openpowerlifting/opl-data). Clone it inside the project directory:

```bash
git clone https://gitlab.com/openpowerlifting/opl-data.git
```

This will create an `opl-data/` folder containing the `meet-data/` directory the API needs.

> **Note:** The OPL dataset is large (~1GB+). The clone may take several minutes.

### 3. Update the data path in `main.go`
Open `main.go` and update `dataFolderPath` to point to the cloned data:

```go
var dataFolderPath = "./opl-data/meet-data"
```

### 4. Install dependencies
```bash
go mod download
```

### 5. Run the server
```bash
go run main.go
```

The server starts on **`:8080`**. Once loaded, you can visit the root endpoint at `http://localhost:8080/` for a JSON overview of all available endpoints.

> **Note:** Loading the full dataset takes a while on startup — this is expected.

---

## API Documentation

All endpoints return JSON. The root endpoint (`GET /`) returns a directory of all available endpoints and pagination info.

### Pagination

All list endpoints support pagination via query parameters:

| Parameter | Default | Max | Description |
|-----------|---------|-----|-------------|
| `limit`   | 50      | 200 | Number of results to return |
| `offset`  | 0       | —   | Number of results to skip |

Paginated responses include metadata:

```json
{
  "total": 1250,
  "limit": 50,
  "offset": 0,
  "data": [...]
}
```

### Unit Conversion

All weight values default to kilograms. Add `?unit=lbs` to convert to pounds.

| Parameter | Values | Default | Description |
|-----------|--------|---------|-------------|
| `unit`    | `kg`, `lbs` | `kg` | Unit for weight values in the response |

Applies to: `bodyweightKg`, `totalKg`, squat/bench/deadlift attempts and bests, and personal bests. DOTS scores are unaffected (unitless).

**Example:** `GET /lifters/Jessica%20Ma?unit=lbs`

### Endpoints

#### `GET /`

Returns an index of all available endpoints with descriptions, pagination, and unit conversion info.

---

#### `GET /lifters`

Returns a paginated list of all lifters with their personal bests and competition history.

**Query parameters:** `limit`, `offset`, `unit`

**Example:** `GET /lifters?limit=2&offset=0`

```json
{
  "total": 1048576,
  "limit": 2,
  "offset": 0,
  "data": [
    {
      "name": "Carlton Ford",
      "pb": {
        "Single-ply": {
          "squat": 127.5,
          "bench": 87.5,
          "deadlift": 160,
          "total": 375
        }
      },
      "competitionResults": [
        {
          "place": "4",
          "name": "Carlton Ford",
          "sex": "M",
          "equipment": "Single-ply",
          "division": "High School",
          "weightClassKg": "56",
          "squat": { "best": 127.5 },
          "bench": { "best": 87.5 },
          "deadlift": { "best": 160 },
          "totalKg": 375,
          "event": "SBD"
        }
      ]
    }
  ]
}
```

---

#### `GET /lifters/names`

Returns a paginated list of all lifter names.

**Query parameters:** `limit`, `offset`

**Example:** `GET /lifters/names?limit=3`

```json
{
  "total": 1048576,
  "limit": 3,
  "offset": 0,
  "data": ["Carlton Ford", "Jessica Ma", "John Smith"]
}
```

---

#### `GET /lifters/{lifterName}`

Returns a single lifter by name. Names with spaces or special characters should be URL-encoded.

**Query parameters:** `unit`

**Example:** `GET /lifters/Jessica%20Ma`

```json
{
  "name": "Jessica Ma",
  "pb": {
    "Raw": {
      "squat": 127.5,
      "bench": 67.5,
      "deadlift": 157.5,
      "total": 352.5,
      "dots": 386.12
    }
  },
  "competitionResults": [
    {
      "place": "1",
      "name": "Jessica Ma",
      "sex": "F",
      "age": 24,
      "equipment": "Raw",
      "division": "Open",
      "bodyweightKg": 55.4,
      "weightClassKg": "56",
      "squat": {
        "attempt1": 100,
        "attempt2": 120,
        "attempt3": 127.5,
        "best": 127.5
      },
      "bench": {
        "attempt1": 55,
        "attempt2": 62.5,
        "attempt3": 67.5,
        "best": 67.5
      },
      "deadlift": {
        "attempt1": 140,
        "attempt2": 150,
        "attempt3": 157.5,
        "best": 157.5
      },
      "totalKg": 352.5,
      "event": "SBD"
    }
  ]
}
```

**Errors:**
- `400` — Invalid lifter name
- `404` — Lifter not found

---

#### `GET /lifters/search`

Search lifters by partial name match (case-insensitive). Results are sorted alphabetically.

**Query parameters:** `q` (required), `limit`, `offset`, `unit`

**Example:** `GET /lifters/search?q=jessica&limit=2`

```json
{
  "total": 15,
  "limit": 2,
  "offset": 0,
  "data": [
    {
      "name": "Jessica Ma",
      "pb": { "Raw": { "squat": 127.5, "bench": 67.5, "deadlift": 157.5, "total": 352.5, "dots": 386.12 } },
      "competitionResults": [...]
    }
  ]
}
```

**Errors:**
- `400` — Missing `q` parameter

---

#### `GET /lifters/top`

Returns a leaderboard of lifters ranked by DOTS score (highest first). Filter by sex, equipment, and/or weight class.

**Query parameters:** `sex`, `equipment`, `weightClass`, `limit`, `offset`, `unit`

**Example:** `GET /lifters/top?sex=M&equipment=Raw&weightClass=83&limit=3`

```json
{
  "total": 5000,
  "limit": 3,
  "offset": 0,
  "data": [
    {
      "name": "John Smith",
      "equipment": "Raw",
      "pb": {
        "squat": 280,
        "bench": 180,
        "deadlift": 320,
        "total": 780,
        "dots": 520.15
      }
    }
  ]
}
```

All filters are optional — omit them to get the overall leaderboard across all categories.

---

#### `GET /records`

Returns all-time records (best squat, bench, deadlift, total) per weight class. Each lift may be held by a different lifter. Filter by sex, equipment, and/or weight class.

**Query parameters:** `sex`, `equipment`, `weightClass`, `limit`, `offset`, `unit`

**Example:** `GET /records?sex=F&equipment=Raw&weightClass=67.5`

```json
{
  "total": 1,
  "limit": 50,
  "offset": 0,
  "data": [
    {
      "weightClassKg": "67.5",
      "sex": "F",
      "equipment": "Raw",
      "squat": { "lift": 200, "lifter": "Jane Doe" },
      "bench": { "lift": 120, "lifter": "Alice Smith" },
      "deadlift": { "lift": 230, "lifter": "Jane Doe" },
      "total": { "lift": 530, "lifter": "Jane Doe" }
    }
  ]
}
```

All filters are optional — omit them to get records across all categories.

---

#### `GET /meets`

Returns a paginated list of all meets across all federations.

**Query parameters:** `limit`, `offset`

**Example:** `GET /meets?limit=2`

```json
{
  "total": 62144,
  "limit": 2,
  "offset": 0,
  "data": [
    {
      "federation": "USAPL",
      "date": "2023-06-10",
      "meetCountry": "USA",
      "meetState": "CA",
      "meetTown": "Los Angeles",
      "meetName": "California State Championships"
    }
  ]
}
```

---

#### `GET /meets/{federationName}`

Returns a paginated list of meets for a specific federation. Federation names with special characters should be URL-encoded.

**Query parameters:** `limit`, `offset`

**Example:** `GET /meets/USAPL?limit=10`

```json
{
  "total": 5432,
  "limit": 10,
  "offset": 0,
  "data": [
    {
      "federation": "USAPL",
      "date": "2023-06-10",
      "meetCountry": "USA",
      "meetState": "CA",
      "meetTown": "Los Angeles",
      "meetName": "California State Championships"
    }
  ]
}
```

**Errors:**
- `400` — Invalid federation name
- `404` — Federation not found

---

#### `GET /meets/{federationName}/{meetName}/results`

Returns all competition entries/results for a specific meet. Both federation name and meet name should be URL-encoded.

**Query parameters:** `limit`, `offset`, `unit`

**Example:** `GET /meets/USAPL/California%20State%20Championships/results?limit=3`

```json
{
  "total": 265,
  "limit": 3,
  "offset": 0,
  "data": [
    {
      "place": "1",
      "name": "Jessica Ma",
      "sex": "F",
      "age": 24,
      "equipment": "Raw",
      "division": "Open",
      "bodyweightKg": 55.4,
      "weightClassKg": "56",
      "squat": { "attempt1": 100, "attempt2": 120, "attempt3": 127.5, "best": 127.5 },
      "bench": { "attempt1": 55, "attempt2": 62.5, "attempt3": 67.5, "best": 67.5 },
      "deadlift": { "attempt1": 140, "attempt2": 150, "attempt3": 157.5, "best": 157.5 },
      "totalKg": 352.5,
      "event": "SBD"
    }
  ]
}
```

**Errors:**
- `400` — Invalid federation or meet name
- `404` — Meet not found

---

#### `GET /federations`

Returns a paginated list of all federation names.

**Query parameters:** `limit`, `offset`

**Example:** `GET /federations?limit=5`

```json
{
  "total": 250,
  "limit": 5,
  "offset": 0,
  "data": ["USAPL", "IPF", "WRPF", "USPA", "CPU"]
}
```

---

### Response Types

#### Lifter

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Lifter's full name |
| `pb` | object | Personal bests keyed by equipment type (e.g. "Raw", "Single-ply") |
| `competitionResults` | array | List of all competition results |

#### PersonalBest

| Field | Type | Description |
|-------|------|-------------|
| `squat` | number | Best squat in kg |
| `bench` | number | Best bench press in kg |
| `deadlift` | number | Best deadlift in kg |
| `total` | number | Best total (squat + bench + deadlift) in kg |
| `dots` | number | DOTS score (strength relative to bodyweight) |

#### CompetitionResult

| Field | Type | Description |
|-------|------|-------------|
| `place` | string | Placing (numeric, or "DQ", "G") |
| `name` | string | Lifter name |
| `birthDate` | string | Birth date |
| `sex` | string | "M", "F", or "Mx" |
| `birthYear` | integer | Birth year |
| `age` | number | Age at time of competition |
| `country` | string | Country |
| `state` | string | State/province |
| `equipment` | string | Equipment class (e.g. "Raw", "Single-ply", "Wraps") |
| `division` | string | Competition division (e.g. "Open", "Juniors 20-23") |
| `bodyweightKg` | number | Bodyweight in kg |
| `weightClassKg` | string | Weight class (e.g. "56", "100+") |
| `squat` | object | Squat attempts and best (see LiftAttempts) |
| `bench` | object | Bench press attempts and best (see LiftAttempts) |
| `deadlift` | object | Deadlift attempts and best (see LiftAttempts) |
| `totalKg` | number | Competition total in kg |
| `event` | string | Event type (e.g. "SBD", "B", "D") |
| `tested` | string | Drug tested status |

Empty or zero-value fields are omitted from the response.

#### LiftAttempts

| Field | Type | Description |
|-------|------|-------------|
| `attempt1` | number | First attempt in kg |
| `attempt2` | number | Second attempt in kg |
| `attempt3` | number | Third attempt in kg |
| `attempt4` | number | Fourth attempt in kg (record attempt) |
| `best` | number | Best successful attempt in kg |

#### Meet

| Field | Type | Description |
|-------|------|-------------|
| `federation` | string | Federation name |
| `date` | string | Meet date |
| `meetCountry` | string | Country |
| `meetState` | string | State/province |
| `meetTown` | string | City/town |
| `meetName` | string | Meet name |
| `ruleSet` | string | Rule set used |

#### Record

| Field | Type | Description |
|-------|------|-------------|
| `weightClassKg` | string | Weight class (e.g. "83", "100+") |
| `sex` | string | "M", "F", or "Mx" |
| `equipment` | string | Equipment class |
| `squat` | object | Record squat: `{ lift, lifter }` |
| `bench` | object | Record bench: `{ lift, lifter }` |
| `deadlift` | object | Record deadlift: `{ lift, lifter }` |
| `total` | object | Record total: `{ lift, lifter }` |

#### TopLifterEntry

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Lifter name |
| `equipment` | string | Equipment class for this PB |
| `pb` | object | PersonalBest (squat, bench, deadlift, total, dots) |

---

## TODO
* Add sorting query parameters to list endpoints
* Include OPL image/logo

## Contributing
If there's a feature missing from this API you'd like to see, feel free to make a pull request, issue, or reach out :)

## Dependencies
Built with Go version 1.24.4

## OPL Links
* [Open Powerlifting Rankings](https://www.openpowerlifting.org/)
* [Open Powerlifting Site](https://openpowerlifting.gitlab.io/opl-csv/)
* [Open Powerlifting Data](https://gitlab.com/openpowerlifting/opl-data)
