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