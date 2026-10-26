#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: bash deploy/test-deployment.sh IMAGE_PREFIX sha-<commit>" >&2
  echo "Build auth, migrations and gateway images locally before running this check." >&2
  exit 1
fi

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
export IMAGE_PREFIX=$1 IMAGE_TAG=$2
export HTTP_PORT=18080 HTTPS_PORT=18443
export HOMESTOCK_HOME
HOMESTOCK_HOME=$(mktemp -d)
export HOMESTOCK_SKIP_PULL=1
export HOMESTOCK_PROJECT
HOMESTOCK_PROJECT="homestock-ci-$(date +%s)-$$"

bash "$script_dir/vps/init-env.sh" "$HOMESTOCK_HOME" http://localhost "$IMAGE_PREFIX"
compose=(docker compose --project-name "$HOMESTOCK_PROJECT" --env-file "$HOMESTOCK_HOME/.env" -f "$script_dir/vps/compose.yml")
cleanup() {
  status=$?
  if [[ $status -ne 0 ]]; then
    "${compose[@]}" logs --tail=100 || true
  fi
  "${compose[@]}" down --volumes --remove-orphans || true
}
trap cleanup EXIT

assert_limits() {
  local service=$1 memory=$2 swap=$3 shares=$4 container_id actual
  container_id=$("${compose[@]}" ps --quiet "$service")
  actual=$(docker inspect --format '{{.HostConfig.Memory}} {{.HostConfig.MemorySwap}} {{.HostConfig.CpuShares}}' "$container_id")
  if [[ "$actual" != "$memory $swap $shares" ]]; then
    echo "$service limits: got '$actual', expected '$memory $swap $shares'." >&2
    return 1
  fi
}

bash "$script_dir/vps/deploy.sh" "$IMAGE_TAG"
assert_limits postgres 536870912 805306368 1024
assert_limits auth-service 134217728 201326592 1024
assert_limits gateway 134217728 201326592 512
go run "$script_dir/smoke-auth.go" http://localhost:18080 "$HOMESTOCK_HOME/account.json"

old_limits_override="$HOMESTOCK_HOME/postgres-before.yml"
cat > "$old_limits_override" <<'YAML'
services:
  postgres:
    mem_limit: 768m
    memswap_limit: 1g
YAML
"${compose[@]}" -f "$old_limits_override" up -d --no-build --no-deps --pull never --wait --wait-timeout 120 postgres
assert_limits postgres 805306368 1073741824 1024
postgres_container_before=$("${compose[@]}" ps --quiet postgres)
postgres_image_before=$(docker inspect --format '{{.Image}}' "$postgres_container_before")

updated_tag="sha-$(printf '%s' "$IMAGE_TAG-update" | sha256sum | cut -c 1-40)"
docker tag "$IMAGE_PREFIX/auth-service:$IMAGE_TAG" "$IMAGE_PREFIX/auth-service:$updated_tag"
docker tag "$IMAGE_PREFIX/auth-migrations:$IMAGE_TAG" "$IMAGE_PREFIX/auth-migrations:$updated_tag"
docker tag "$IMAGE_PREFIX/gateway:$IMAGE_TAG" "$IMAGE_PREFIX/gateway:$updated_tag"
bash "$script_dir/vps/deploy.sh" "$updated_tag"
export IMAGE_TAG=$updated_tag
assert_limits postgres 536870912 805306368 1024
assert_limits auth-service 134217728 201326592 1024
assert_limits gateway 134217728 201326592 512
postgres_container_after=$("${compose[@]}" ps --quiet postgres)
postgres_image_after=$(docker inspect --format '{{.Image}}' "$postgres_container_after")
if [[ "$postgres_container_before" == "$postgres_container_after" || "$postgres_image_before" != "$postgres_image_after" ]]; then
  echo "Database limits must be applied by replacing its container while retaining its image." >&2
  exit 1
fi
go run "$script_dir/smoke-auth.go" -reuse http://localhost:18080 "$HOMESTOCK_HOME/account.json"
gzip -t "$HOMESTOCK_HOME"/backups/*.sql.gz
echo "Fresh deployment, runtime limits and redeployment with preserved database image and data passed."
