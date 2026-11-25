#!/usr/bin/env bash
# Run on the Ubuntu host, outside the restricted coding session.
set +x
set -euo pipefail
umask 077

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
env_file="$project_dir/.env"

if [[ "$EUID" -ne 0 ]]; then
  echo "Run with sudo: sudo bash scripts/setup-postgres.sh" >&2
  exit 1
fi
if [[ -e "$env_file" || -L "$env_file" ]]; then
  echo "An .env file already exists. Leaving it and any existing database unchanged." >&2
  exit 1
fi
if ! command -v apt-get >/dev/null; then
  echo "This setup script requires an Ubuntu/Debian host." >&2
  exit 1
fi

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y postgresql postgresql-contrib

# Only use the default local cluster; do not change other clusters or ports.
pg_version="$(pg_lsclusters --no-header | awk '$2 == "main" && $3 == "5432" { print $1; exit }')"
if [[ -z "$pg_version" ]]; then
  echo "No main PostgreSQL cluster on port 5432 was found. No database was created." >&2
  exit 1
fi
if ! pg_ctlcluster "$pg_version" main status >/dev/null; then
  pg_ctlcluster "$pg_version" main start
fi

admin_psql=(runuser -u postgres -- psql --no-psqlrc --host=/var/run/postgresql --port=5432 --username=postgres --dbname=postgres --set=ON_ERROR_STOP=1)
existing="$("${admin_psql[@]}" --tuples-only --no-align --command="SELECT EXISTS (SELECT FROM pg_roles WHERE rolname = 'ttwms_dev') OR EXISTS (SELECT FROM pg_database WHERE datname = 'ttwms_dev')")"
if [[ "$existing" != "f" ]]; then
  echo "A ttwms_dev role or database already exists. Leaving it unchanged." >&2
  exit 1
fi

database_password="$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')"

# Save credentials before database creation so they remain recoverable if a
# later command fails. Exclusive creation prevents replacing an existing file.
(
  set -o noclobber
  printf 'APP_ENV=development\nAPP_ADDR=127.0.0.1:8080\nDATABASE_URL=postgres://ttwms_dev:%s@127.0.0.1:5432/ttwms_dev?sslmode=disable\n' "$database_password" > "$env_file"
)

# The generated hex password contains no SQL quotes. Keep it out of command-line
# arguments and terminal output; this role has no cluster administration rights.
printf "CREATE ROLE ttwms_dev LOGIN PASSWORD '%s';\nCREATE DATABASE ttwms_dev OWNER ttwms_dev;\n" "$database_password" | "${admin_psql[@]}"
PGPASSWORD="$database_password" psql --no-psqlrc --host=127.0.0.1 --port=5432 --username=ttwms_dev --dbname=ttwms_dev --set=ON_ERROR_STOP=1 --command='SELECT 1' >/dev/null
unset database_password

echo "Development PostgreSQL is ready. Credentials were saved to .env with mode 600."
echo "Start the application with: go run ./cmd/server -env-file .env"
