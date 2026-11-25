# TT WMS

TikTok-first warehouse management system built with Go, server-rendered HTML, htmx, Fomantic UI CSS, PostgreSQL, and Redis.

## Current foundation

- Go HTTP server with graceful shutdown and security headers.
- Typed environment configuration validated before the server starts.
- Embedded templates and static assets.
- Liveness endpoint at `/health/live`.
- Database readiness endpoint at `/health/ready`.
- Create/edit screens for locations, trolleys, and packing stations, with validation and transactional audit records.
- Embedded, versioned PostgreSQL migrations with checksums and serialized startup.
- Docker Compose definition for app, PostgreSQL, and Redis.
- Initial company, user, warehouse, location, trolley, packing-station, and audit schema.

## Run locally

Docker is not installed in the current workspace. Once available:

```sh
docker compose up --build
```

Run the application with settings from the ignored `.env` file:

```sh
go run ./cmd/server -env-file .env
```

The application reads the explicitly selected file as data without evaluating shell expressions; process environment variables take priority. No Python runtime is needed. You can also run directly with settings exported in the environment:

```sh
go run ./cmd/server
```

Then open `http://localhost:8080`.

With `DATABASE_URL` configured in development/test, startup connects to PostgreSQL, applies pending migrations, and initializes one company and warehouse. Open `/setup/locations`, `/setup/trolleys`, or `/setup/packing-stations` to create and edit assets. Audit records are committed atomically with changes. Setup writes are available only in `development` and `test`; keep the development server private until authentication and authorization are implemented.

In development/test without `DATABASE_URL`, the home page and liveness endpoint still work; setup routes are unavailable and readiness returns 503. Staging/production require `DATABASE_URL`. An unavailable configured database prevents startup.

### Native development PostgreSQL

On an Ubuntu/Debian host, run outside the restricted coding session:

```sh
sudo bash scripts/setup-postgres.sh
```

This installs PostgreSQL, starts the default local cluster on port 5432, creates a dedicated `ttwms_dev` role and database, and writes a generated password to the ignored `.env` file with owner-only permissions. It refuses to replace an existing `.env`, role, or database. The file is saved before database creation so credentials remain available if setup fails partway through. The script does not apply application migrations or change PostgreSQL's network access settings.

## Configuration

The server reads process environment variables. Select a file explicitly with `-env-file .env` to use file defaults; it never loads a file implicitly. Files support literal `KEY=value` lines, optional matching single/double quotes, blank lines, and full-line comments. Duplicate keys and malformed lines are rejected without including their values in errors.

| Variable | Default | Validation |
| --- | --- | --- |
| `APP_ENV` | `development` | `development`, `test`, `staging`, or `production` |
| `APP_ADDR` | `:8080` | Host and numeric port from 1 to 65535; an empty host listens on all interfaces |
| `DATABASE_URL` | Unset | `postgres://` or `postgresql://` URL with a host and database name; required in staging/production |

Invalid settings stop startup with an error naming the variable without exposing its value. When configured, PostgreSQL is required and checked before the server listens.

For example:

```sh
APP_ENV=development APP_ADDR=127.0.0.1:8080 go run ./cmd/server
```

## Frontend dependencies

The setup pages use native HTML forms and the application stylesheet. The htmx file is a placeholder and is not loaded; htmx and Fomantic UI CSS must be pinned and vendored before adding the planned frontend enhancements. Production must not depend on a public CDN.

## Verification and migrations

Run on the host with network access and PostgreSQL available:

```sh
bash scripts/check.sh
```

This downloads and verifies locked Go modules, runs the race-enabled test suite including PostgreSQL integration tests, runs `go vet`, and builds `bin/ttwms`. Builds and tests use `-mod=readonly` so dependency changes must be made deliberately and committed with `go.sum`. Integration tests read `.env` through `TEST_ENV_FILE` (with process settings taking priority), reject staging/production environments, use a disposable schema, and check concurrent migrations, read-only schema verification, create/edit persistence, duplicate codes/barcodes, warehouse isolation, atomic audit rollback, and migration checksum protection. They ensure the `pgcrypto` extension exists in `public` and remove their disposable schema afterward.

Tests that do not require PostgreSQL or the database driver can run independently:

```sh
go test -race ./internal/... ./migrations
```

Add a new numbered SQL file for schema changes; never edit an applied migration. Databases created using the earlier Compose initialization scripts have no migration history and require an explicit baseline before using this migration runner. It does not automatically adopt or drop existing tables.

Staging/production startup only checks schema compatibility. Apply pending migrations explicitly using `./bin/ttwms -env-file /path/to/migration.env -migrate` before starting the app. See [deployment notes](deploy/README.md) for the separation between this development server and a future clean production environment.

## Secrets

Copy `.env.example` to `.env` and fill values only on the target machine. Never commit TikTok, carrier, database, or session secrets.
