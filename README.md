# TT WMS

TikTok-first warehouse management system built with Go, server-rendered HTML, htmx, Fomantic UI CSS, PostgreSQL, and Redis.

## Current foundation

- Go HTTP server with graceful shutdown and security headers.
- Typed environment configuration validated before the server starts.
- Embedded templates and static assets.
- Liveness endpoint at `/health/live`.
- Docker Compose definition for app, PostgreSQL, and Redis.
- Initial company, user, warehouse, location, trolley, packing-station, and audit schema.

## Run locally

Docker is not installed in the current workspace. Once available:

```sh
docker compose up --build
```

The standard-library-only server can run now:

```sh
go run ./cmd/server
```

Then open `http://localhost:8080`.

## Configuration

The server reads process environment variables; it does not automatically load `.env` files.

| Variable | Default | Validation |
| --- | --- | --- |
| `APP_ENV` | `development` | `development`, `test`, `staging`, or `production` |
| `APP_ADDR` | `:8080` | Host and numeric port from 1 to 65535; an empty host listens on all interfaces |
| `DATABASE_URL` | Unset | Optional `postgres://` or `postgresql://` URL with a host and database name |

Invalid settings stop startup with an error naming the variable without exposing its value. Database configuration is validated only; the application does not connect to PostgreSQL yet, so running without a database remains supported.

For example:

```sh
APP_ENV=development APP_ADDR=127.0.0.1:8080 go run ./cmd/server
```

## Frontend dependencies

The repository contains placeholders for vendored htmx and Fomantic UI CSS. Pin and vendor reviewed releases before introducing interactive pages; production must not depend on a public CDN.

## Secrets

Copy `.env.example` to `.env` and fill values only on the target machine. Never commit TikTok, carrier, database, or session secrets.
