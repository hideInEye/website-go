#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd -P)"
config_file="$project_dir/config/config.toml"
data_dir="$project_dir/data"

if [ ! -f "$config_file" ]; then
  echo "Missing backend configuration: $config_file" >&2
  exit 1
fi
if ! command -v docker >/dev/null 2>&1; then
  echo 'Docker is required.' >&2
  exit 1
fi
docker compose version >/dev/null

# Probe without capturing output: BSD stat may print before rejecting -c.
if stat -c '%u' "$config_file" >/dev/null 2>&1; then
  stat_flavor=gnu
else
  stat_flavor=bsd
fi

file_stat() {
  if [ "$stat_flavor" = gnu ]; then
    stat -c "$1" "$3"
  else
    stat -f "$2" "$3"
  fi
}

mode="$(file_stat '%a' '%Lp' "$config_file")"
case "$mode" in
  400|600) ;;
  *)
    echo "Backend TOML contains secrets. Run: chmod 600 \"$config_file\"" >&2
    exit 1
    ;;
esac

owner_uid="$(file_stat '%u' '%u' "$config_file")"
owner_gid="$(file_stat '%g' '%g' "$config_file")"
APP_UID="${APP_UID:-$owner_uid}"
APP_GID="${APP_GID:-$owner_gid}"
export APP_UID APP_GID
for identity in "$APP_UID" "$APP_GID"; do
  case "$identity" in
    ''|*[!0-9]*)
      echo 'APP_UID and APP_GID must be numeric IDs.' >&2
      exit 1
      ;;
  esac
done
if [ "$owner_uid" != "$APP_UID" ]; then
  echo "TOML owner UID $owner_uid does not match the explicit APP_UID=$APP_UID." >&2
  echo 'Use the configuration owner UID, or deliberately adjust ownership before deployment.' >&2
  exit 1
fi

# Create runtime storage as the invoking user, never as Docker's root user.
# Existing directories are checked, not silently chowned or chmodded.
mkdir -p "$data_dir"
data_owner="$(file_stat '%u' '%u' "$data_dir")"
data_mode="$(file_stat '%a' '%Lp' "$data_dir")"
if [ "$data_owner" != "$APP_UID" ] || [ ! -w "$data_dir" ] || [ ! -x "$data_dir" ]; then
  echo "Runtime directory must be writable and owned by APP_UID=$APP_UID: $data_dir" >&2
  echo 'Ask its owner or an administrator to adjust it; this script will not change ownership.' >&2
  exit 1
fi
case "$data_mode" in
  ''|*[!0-7]*)
    echo "Cannot inspect runtime directory permissions: $data_dir" >&2
    exit 1
    ;;
esac
if [ "$((0$data_mode & 0300))" -ne 192 ]; then
  echo "Runtime directory owner needs write and execute permissions: $data_dir" >&2
  exit 1
fi

docker compose -f "$project_dir/docker-compose.yml" config --quiet
echo 'Compose paths, runtime directory and configuration permissions are ready. No database, media probe or TOML runtime validation was performed.'
