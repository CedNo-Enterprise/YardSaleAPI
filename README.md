# Postgres via Docker Compose

## Setup

1. Edit `.env` and set your own values (it's already gitignored, so it stays local).
2. Start the database:
   ```bash
   docker compose up -d
   ```
3. Check it's running:
   ```bash
   docker compose ps
   ```

## Connect

```bash
docker exec -it my-postgres psql -U myuser -d mydatabase
```

Or from any client (DBeaver, pgAdmin, etc.):
- Host: `localhost`
- Port: `5432`
- User / Password / DB: whatever you set in `.env`

## Stop

```bash
docker compose down       # stops and removes the container, keeps data
docker compose down -v    # also deletes the data volume (careful!)
```

## Database migrations (golang-migrate)

Schema changes live in `migrations/` as paired `.up.sql` / `.down.sql` files,
each prefixed with a version number (`000001_`, `000002_`, ...).

**Install the CLI** (one-time, on your machine):
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
Make sure `$(go env GOPATH)/bin` is on your `PATH` so the `migrate` command is available.

**Apply all pending migrations:**
```bash
migrate -database "${DATABASE_URL}" -path migrations up
```

**Roll back the last migration:**
```bash
migrate -database "${DATABASE_URL}" -path migrations down 1
```

**Create a new migration** (generates empty up/down files for you to fill in):
```bash
migrate create -ext sql -dir migrations -seq add_posts_table
```

