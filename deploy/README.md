# Development and future production

This machine is the development server. Use `APP_ENV=development`, the dedicated `ttwms_dev` database, test credentials, and sample or anonymised data. Its configuration and database stay on this machine when production is created.

Build the same Go binary for every environment; select settings at runtime. The application has no Python dependency. Native development starts with:

```sh
go run ./cmd/server -env-file .env
```

For a future release, build with the committed dependency files:

```sh
go mod download
go mod verify
go build -mod=readonly -trimpath -o bin/ttwms ./cmd/server
```

Create production on a clean host with its own database, credentials, and configuration based on `production.env.example`. Supply configuration through the service environment or an explicit file outside the checkout. No `.env` file is read implicitly. Reuse the reviewed schema and application code, and import only explicitly selected operational configuration/data through supported tools; do not copy the development database.

In staging and production, normal startup only verifies migration checksums and rejects a missing, pending, or newer schema. Run migrations once during deployment using schema-owner credentials, then run the app using separate runtime credentials with only the necessary data-access privileges:

```sh
./bin/ttwms -env-file /etc/ttwms/migration.env -migrate
./bin/ttwms -env-file /etc/ttwms/app.env
```

Both environment files must select the target environment/database. Explicit process environment variables override file values, including the database URL. Migration credentials must not be present in the running application's environment. The migration runner initializes one company and warehouse on a fresh database; later changes use new numbered migrations.

The production host is a later milestone. Authentication and role enforcement are still required before enabling warehouse setup there; those routes currently remain unavailable outside development/test. Production deployment also needs HTTPS/service supervision, verified backup restoration, and monitoring. The existing Compose file is a development environment, with development database credentials; it is not a production deployment definition.
