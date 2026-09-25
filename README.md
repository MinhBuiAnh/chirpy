# Chirpy

Chirpy is a Go-based social microblogging web app and API for creating users, logging in, publishing chirps, and managing authentication flows. The project serves a small frontend from the root directory while exposing a REST API backed by PostgreSQL.

## What this project does

This app provides the core pieces of a lightweight social platform:

- User registration and authentication
- JWT-based login, refresh, and revoke flows
- Chirp creation, listing, retrieval, and deletion
- Database-backed persistence with SQLC-generated queries
- Admin endpoints for health checks, metrics, and reset behavior in development mode
- A simple file-serving frontend served by the Go app

The project is designed as a small backend service with a simple UI layer, rather than a full-featured social network.

## Prerequisites

Before running the app, make sure you have:

- Go 1.27 or newer
- PostgreSQL running locally or in a reachable environment
- A configured `.env` file with the required environment variables

## Installation

1. Open a terminal in the project root:

   ```bash
   cd chirpy
   ```

2. Download Go dependencies:

   ```bash
   go mod download
   ```

3. Create or update a `.env` file in the project root. The app expects these values:

   ```env
   DB_URL="postgres://[username]:[password]@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"
   SECRET="your-secret-key"
   POLKA_KEY="your-polka-key"
   ```

   The included `.env` file is a local development example. For real deployments, replace the secrets with secure values.

4. Make sure your PostgreSQL database exists and the schema under `sql/schema/` has been applied before starting the server.

## Running the project

Start the server with:

```bash
go run .
```

The app listens on port `8080` by default.

You can then open the frontend or call the API endpoints:

- Health check: `GET /api/healthz`
- API base: `http://localhost:8080/api/...`
- Frontend files: `http://localhost:8080/app`

## Project structure

- `main.go` — application entry point and route setup
- `api_*.go` — handlers for users, chirps, auth, admin, and webhook features
- `internal/auth/` — JWT and password utilities
- `internal/database/` — DB access layer and generated SQL queries
- `sql/schema/` and `sql/queries/` — database schema and SQL definitions
- `assets/` and `index.html` — frontend assets and static app files

## Notes

This README is intentionally focused on setup and usage. Detailed endpoint-by-endpoint documentation is kept in a separate file so the main project docs stay concise and easy to follow.

See [API.md](API.md) for a lighter reference of the HTTP interface and request patterns.
