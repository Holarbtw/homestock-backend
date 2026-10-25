#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 || ! "$1" =~ ^sha-[0-9a-f]{40}$ ]]; then
  echo "Usage: bash deploy.sh sha-<full-40-character-commit>" >&2
  exit 1
fi

export IMAGE_TAG=$1
release_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
deployment_home=${HOMESTOCK_HOME:-/opt/homestock}
env_file="$deployment_home/.env"

if [[ ! -f "$env_file" ]]; then
  echo "Missing $env_file. Run init-env.sh first." >&2
  exit 1
fi

exec 9>"$deployment_home/deploy.lock"
flock -n 9 || { echo "Another deployment is running." >&2; exit 1; }

compose=(docker compose --project-name "${HOMESTOCK_PROJECT:-homestock-vps}" --env-file "$env_file" -f "$release_dir/compose.yml")
diagnostics() {
  echo "Deployment failed. Inspect the logs before retrying; no database rollback was performed." >&2
  "${compose[@]}" ps -a || true
  "${compose[@]}" logs --tail=60 auth-service gateway || true
}
trap diagnostics ERR

"${compose[@]}" config --quiet
if [[ ${HOMESTOCK_SKIP_PULL:-0} != 1 ]]; then
  "${compose[@]}" --profile tools pull
fi

"${compose[@]}" up -d --no-build --no-recreate --wait --wait-timeout 120 postgres

umask 077
mkdir -p "$deployment_home/backups"
backup_file="$deployment_home/backups/auth-$(date -u +%Y%m%dT%H%M%S)-$IMAGE_TAG.sql.gz"

# Expand these variables inside the PostgreSQL container, not on the host.
# shellcheck disable=SC2016
"${compose[@]}" exec -T postgres sh -c \
  'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB"' | gzip > "$backup_file.tmp"
mv -- "$backup_file.tmp" "$backup_file"

"${compose[@]}" run --rm --no-deps auth-migrations up
"${compose[@]}" up -d --no-build --no-deps --wait --wait-timeout 120 auth-service
"${compose[@]}" up -d --no-build --no-deps --force-recreate --wait --wait-timeout 120 gateway

printf '%s\n' "$IMAGE_TAG" > "$release_dir/IMAGE_TAG"
ln -sfn -- "$release_dir" "$deployment_home/current"
trap - ERR
echo "Deployed $IMAGE_TAG. Database backup: $backup_file"
