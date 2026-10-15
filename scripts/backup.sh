#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd -P)"

# Backup client parameters come only from the caller's environment. Validate
# every required value before calling Docker or pausing any running service.
: "${MYSQL_HOST:?Set MYSQL_HOST for the backup client}"
: "${MYSQL_PORT:?Set MYSQL_PORT for the backup client}"
: "${MYSQL_DATABASE:?Set MYSQL_DATABASE for the backup client}"
: "${MYSQL_USER:?Set MYSQL_USER for the backup client}"
: "${MYSQL_PASSWORD:?Set MYSQL_PASSWORD for the backup client}"
: "${MINIO_ENDPOINT:?Set MINIO_ENDPOINT for the backup client}"
: "${MINIO_PORT:?Set MINIO_PORT for the backup client}"
: "${MINIO_USE_SSL:?Set MINIO_USE_SSL to true or false for the backup client}"
: "${MINIO_ACCESS_KEY:?Set MINIO_ACCESS_KEY for the backup client}"
: "${MINIO_SECRET_KEY:?Set MINIO_SECRET_KEY for the backup client}"
: "${MINIO_BUCKET:?Set MINIO_BUCKET for the backup client}"
case "$MINIO_USE_SSL" in
  true|false) ;;
  *)
    echo 'MINIO_USE_SSL must be true or false.' >&2
    exit 1
    ;;
esac

. "$script_dir/preflight.sh"
cd "$project_dir"

compose() {
  docker compose -f "$project_dir/docker-compose.yml" "$@"
}

backup_dir="${1:-./backups}"
timestamp="$(date +%Y%m%d-%H%M%S)"
mkdir -p "$backup_dir"
backup_dir="$(CDPATH= cd -- "$backup_dir" && pwd -P)"
mysql_name="mysql-$timestamp.sql.gz"
minio_name="minio-$timestamp.tar.gz"
mysql_raw="$backup_dir/.mysql-$timestamp.sql"
mysql_partial="$backup_dir/.$mysql_name.partial"
minio_partial="$backup_dir/.$minio_name.partial"
minio_stage="$backup_dir/.minio-$timestamp"
services_paused=0

umask 077

resume_services() {
  if [ "$services_paused" -eq 1 ]; then
    compose up -d api web >/dev/null
    services_paused=0
  fi
}

cleanup() {
  status=$?
  trap - EXIT HUP INT TERM
  rm -f "$mysql_raw" "$mysql_partial" "$minio_partial"
  rm -rf "$minio_stage"
  if ! resume_services; then
    echo "Backup cleanup could not restart all services." >&2
    status=1
  fi
  exit "$status"
}

trap cleanup EXIT HUP INT TERM

compose stop web api >/dev/null
services_paused=1

compose --profile ops run --rm -T --no-deps mysql-client -ceu '
  export MYSQL_PWD="$MYSQL_PASSWORD"
  exec mysqldump \
    --host="$MYSQL_HOST" \
    --port="$MYSQL_PORT" \
    --user="$MYSQL_USER" \
    --single-transaction \
    --routines \
    --triggers \
    --events \
    --hex-blob \
    --set-gtid-purged=OFF \
    --column-statistics=0 \
    "$MYSQL_DATABASE"
' \
  > "$mysql_raw"
gzip -c "$mysql_raw" > "$mysql_partial"

mkdir "$minio_stage"
compose --profile ops run --rm -T --no-deps \
  -v "$minio_stage:/backup" \
  minio-client -ceu '
    scheme=http
    if [ "$MINIO_USE_SSL" = "true" ]; then
      scheme=https
    fi
    mc alias set source "$scheme://$MINIO_ENDPOINT:$MINIO_PORT" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" >/dev/null
    mc mirror --overwrite "source/$MINIO_BUCKET" /backup/bucket
  '
tar -C "$minio_stage" -czf "$minio_partial" bucket

mv "$mysql_partial" "$backup_dir/$mysql_name"
mv "$minio_partial" "$backup_dir/$minio_name"
rm -f "$mysql_raw"
rm -rf "$minio_stage"

resume_services
trap - EXIT HUP INT TERM

echo "MySQL backup: $backup_dir/$mysql_name"
echo "MinIO backup: $backup_dir/$minio_name"
