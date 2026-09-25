# API Notes

This document contains a brief overview of the Chirpy HTTP interface. It is intentionally short and is meant to complement the main README.

## API groups

### Authentication and users

- `POST /api/users` - create a user
- `POST /api/login` - authenticate a user and return tokens
- `POST /api/refresh` - refresh access tokens
- `POST /api/revoke` - revoke a refresh token
- `PUT /api/users` - update a user's email or password

### Chirps

- `POST /api/chirps` - create a chirp
- `GET /api/chirps` - list chirps
- `GET /api/chirps/{chirpId}` - fetch a single chirp
- `DELETE /api/chirps/{chirpID}` - delete a chirp

### Admin and health

- `GET /api/healthz` - basic readiness check
- `GET /admin/metrics` - basic app metrics page
- `POST /admin/reset` - reset app data in development mode only
- `POST /api/polka/webhooks` - webhook endpoint for Polka integration

## Request conventions

- JSON request bodies are used for most write operations.
- Authorization is handled through JWT tokens in the request headers.
- The app expects a valid Postgres connection string and secret values defined in the environment.

## Further documentation

This project does not include a full endpoint reference in the main README intentionally. For more detail, see the implementation under the `api_*.go` files and the SQL-backed database layer in `internal/database/`.
