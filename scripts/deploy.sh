#!/usr/bin/env sh
set -eu
script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd -P)"
. "$script_dir/preflight.sh"
compose() {
  docker compose -f "$project_dir/docker-compose.yml" "$@"
}
echo 'Pulling prebuilt images (no server build)...'
compose pull api
compose pull web
echo 'Checking/initializing blog tables without importing or replacing existing data...'
compose --profile ops run --rm --no-deps migrate
echo 'Starting Go API and static Web containers...'
compose up -d --no-build api web
compose ps
