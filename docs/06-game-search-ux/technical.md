# Game Search UX — Technical Tasks

## Backend

### T6.1 — BGG client: search ranking

`internal/bgg/client.go` — update `Search(query string)`:

After fetching and parsing the BGG XML response, apply server-side ranking before returning:

```
tier 1 — strings.EqualFold(name, query)
tier 2 — strings.HasPrefix(lower(name), lower(query)) && !strings.Contains(name, ":")
tier 3 — strings.Contains(lower(name), lower(query)) && !strings.Contains(name, ":")
tier 4 — everything else
```

Within each tier, sort by year ascending (nil years sorted last).

---

### T6.2 — BGG client: GetThing

`internal/bgg/client.go` — add method:

```go
func (c *Client) GetThing(bggID int) (*BGGGameDetail, error)
```

- Calls `GET /thing?id=<bggID>&stats=1`.
- If response is HTTP 202, wait 2 seconds and retry once.
- Returns `404`-equivalent error if item not found in response.
- Parses and returns:

```go
type BGGGameDetail struct {
    BGGId         int
    Name          string
    Description   string
    CoverImageURL string  // image[@type="image"]
    ThumbnailURL  string
    MinPlayers    int
    MaxPlayers    int
    BGGRating     float64
}
```

---

### T6.3 — New handler: GET /api/bgg/thing

`internal/handlers/bgg.go` — add `Thing` handler:

- Query param: `id` (integer, required).
- Calls `bgg.GetThing(id)`.
- Returns `404` if BGG ID not found, `502` if BGG unreachable.
- Returns JSON matching `BGGGameDetail`.

Register in router:
```go
r.Get("/api/bgg/thing", bggHandler.Thing)
```

---

### T6.4 — Update POST /api/games handler

`internal/handlers/games.go` — update `Create` handler:

Request body changes to:
```json
{ "bgg_id": "integer | null", "name": "string | null" }
```

Logic:
- If `bgg_id` is set: call `bgg.GetThing(bgg_id)`, use returned data to populate all game fields. Ignore any `name` in the request.
- If only `name` is set: store name, leave all other fields null.
- If neither is set: return `400`.
- Existing `409` on duplicate `bgg_id` is unchanged.
- On BGG unreachable: return `502`.

---

## Frontend

### T6.5 — Rewrite /games/new page

`src/routes/(app)/(global)/games/new/+page.svelte`

Three states:

**State 1 — Search**
- Text input with debounced calls to `GET /api/bgg/search?q=<query>`.
- Results list: thumbnail, name, year. No client-side sorting (server handles it).
- "Add manually" link below the search box → State 3.

**State 2 — Preview**
- Triggered by selecting a result.
- Calls `GET /api/bgg/thing?id=<bgg_id>`.
- Shows read-only card: cover image, name, year, player count range, BGG rating.
- No editable fields.
- "Add to catalog" button → `POST /api/games` with `{ bgg_id }`.
  - On `409`: show "Already in catalog" message with link to add to collection.
  - On success: redirect to `/games/<new_id>`.
- "← Back" link → State 1.

**State 3 — Manual add**
- Single name text field.
- "Add" button → `POST /api/games` with `{ name }`.
- On success: redirect to `/games/<new_id>`.
