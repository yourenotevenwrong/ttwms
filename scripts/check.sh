#!/usr/bin/env bash
# Run on the host when the coding session cannot download modules or use sockets.
set -euo pipefail
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"
go mod download
go mod verify
TEST_ENV_FILE="$project_dir/.env" go test -mod=readonly -race -tags integration ./...
go vet -mod=readonly ./...
mkdir -p bin
go build -mod=readonly -o bin/ttwms ./cmd/server
