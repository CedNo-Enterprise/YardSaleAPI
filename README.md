# GarageSaleAPI

A Go API for garage sales. Sellers list sales and the items in them; buyers plan
an **itinerary** — a route across several sales on a given day, ordered by
proximity to a starting point.

Requires Go 1.26 and Docker.

## Setup

1. Copy `.env.example` to `.env` and set your own values (`.env` is gitignored).
2. Start the database:
   ```bash
   docker compose up -d
   docker compose ps
   ```
3. Run the API — it creates any missing tables on boot:
   ```bash
   go run .
   ```

It listens on `localhost:8080`.

## Test data

`cmd/seed` fills the database with two users, a seller, seven Ottawa-area sales
with real coordinates, and two itineraries:

```bash
go run ./cmd/seed
```

It prints the credentials and ids it created. Every row uses a fixed id, so it is
safe to re-run: it clears its own previous rows and leaves everything else alone.

Both seeded users share the password `correcthorsebattery`. `buyer@example.com`
owns the itineraries; `vendor@example.com` owns the sales, which makes it useful
for checking that ownership is enforced.

One seeded sale is deliberately left ungeocoded, so you can see how an
unlocatable stop is handled.

## Tests

```bash
go test ./...
```

Everything runs without a database — the repositories have in-memory doubles in
`infrastructure/persistence/memory`. There is no automated coverage of the GORM
repositories themselves; those are exercised by running the API against Postgres.

## API

Authenticated routes take `Authorization: Bearer <token>` from `POST /login`.
Reads are public; everything that changes an itinerary is restricted to its owner
and returns `403` otherwise.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/user` | — | Create a user |
| GET | `/user/{id}` | — | Fetch a user |
| POST | `/login` | — | Exchange credentials for a token |
| POST | `/logout` | yes | Revoke the current token |
| POST | `/seller` | yes | Become a seller |
| GET | `/seller/{id}` | — | Fetch a seller |
| GET | `/seller/user/{userId}` | — | Fetch a seller by user |
| POST | `/sale` | yes | Create a sale |
| GET | `/sale/{id}` | — | Fetch a sale |
| POST | `/itinerary` | yes | Create a route |
| GET | `/itinerary` | yes | List your routes |
| GET | `/itinerary/{id}` | — | Fetch a route |
| PATCH | `/itinerary/{id}` | owner | Rename, reschedule, move the start point |
| DELETE | `/itinerary/{id}` | owner | Delete a route |
| POST | `/itinerary/{id}/stop` | owner | Append a sale to the route |
| PATCH | `/itinerary/{id}/stop/{saleId}` | owner | Set a stop to planned/visited/skipped |
| DELETE | `/itinerary/{id}/stop/{saleId}` | owner | Remove a stop |
| PUT | `/itinerary/{id}/stops/order` | owner | Re-optimize from a start point |

### Itineraries

Create one with the sales you want to visit and where you are starting from:

```bash
curl -X POST localhost:8080/itinerary \
  -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
  -d '{
        "name": "Saturday morning run",
        "date": "2026-09-19T08:00:00Z",
        "startLatitude": 45.4215,
        "startLongitude": -75.6972,
        "saleIds": ["<sale-id>", "<sale-id>"]
      }'
```

The server decides the order — a nearest-neighbour walk from the start point —
so `saleIds` is a set, not a sequence. Stops come back with a `position` from
zero and a `status` of `planned`, `visited` or `skipped`.

Two behaviours worth knowing:

- `POST /stop` **appends** to the end rather than re-optimizing, so adding a sale
  stays a single write. Call `PUT /stops/order` when you want a fresh route.
- `PUT /stops/order` re-runs the optimizer from a new start point; omit the
  coordinates to reuse the one already stored. Reordering only moves positions,
  so per-stop status survives it.

**Sale addresses are not geocoded yet.** Nothing populates latitude and longitude
on the `POST /sale` path, so real sales are stored at `(0, 0)`. The optimizer
treats that as "location unknown" and leaves those stops in submitted order at
the end of the route. Until geocoding is added, ordering is only meaningful for
sales whose coordinates were set another way — which is why `cmd/seed` sets them
directly.

## Database

The schema is created by GORM's `AutoMigrate`, called from `main.go` on boot and
from `cmd/seed`. There are no migration files; to change the schema, edit the
structs in `infrastructure/persistence/database/records` and restart.

`AutoMigrate` cannot express every change. Anything it refuses — a column type
change needing an explicit `USING` clause, for example — goes in `Gorm.go` as a
guarded step that runs before it and is a no-op once applied. See
`widenSellerUserIdToUuid`.

Connect with psql:

```bash
# substituting the POSTGRES_USER and POSTGRES_DB values from your .env
docker exec -it yardsale-db psql -U myuser -d mydatabase
```

Or point any client at `localhost:5432` with the credentials from `.env`.

```bash
docker compose down       # stop, keep the data
docker compose down -v    # stop and delete the data volume
```

## Layout

Clean architecture; dependencies point inward. Nothing under `domain` imports
anything outside `domain`.

```
domain/            entities, repository interfaces, business rules
  itinerary/         route optimizer lives here — it is a business rule, not plumbing
application/       services (use cases), the composition root, error kinds
infrastructure/    GORM repositories, records, in-memory doubles
interfaces/        HTTP handlers, request/response DTOs, auth middleware
cmd/seed/          test-data seeder
```

Entities keep their fields unexported and are built through `CreateX` (new, with
business defaults) or `HydrateX` (rebuilt from storage). Services validate
request DTOs and return domain objects; the `responses` package maps those to
JSON. Errors carry an `apperror.Kind`, which `application/server/httpx.go` maps
to a status code.
