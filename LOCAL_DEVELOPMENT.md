# Local development

This setup runs only the infrastructure in Docker. The Go API and Vite
development server run directly on the host.

## Requirements

- Docker Desktop with Docker Compose
- Go 1.27 or newer
- Node.js 24 or newer
- pnpm 11

## Start PostgreSQL and Redis

From the repository root, create the local environment file and start the
containers:

```bash
cp .env.example .env
docker compose up -d
docker compose ps
```

The containers expose PostgreSQL on `localhost:5433` and Redis on
`localhost:6379`. Their data is stored in the `postgres-data` and
`redis-data` Docker volumes.

## Start the Go API

In a new terminal using Git Bash, load the local environment and start the API:

```bash
set -a
source .env
set +a
go run . web
```

The API is available at `http://localhost:3456`.

## Start the Vite frontend

In another terminal:

```bash
cp frontend/.env.local.example frontend/.env.local
cd frontend
pnpm install
pnpm dev
```

Open `http://localhost:4173`. Vite proxies `/api` requests to the local Go
API at `http://localhost:3456`.

## Stop the infrastructure

Stop the containers while keeping their data:

```bash
docker compose down
```

To remove the containers and all local database/Redis data, use this only when
you intentionally want a fresh environment:

```bash
docker compose down -v
```

The credentials in `.env.example` are for local development only. Keep `.env`
uncommitted and use different credentials and secrets outside your local
development environment.
