#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: bash deploy/test-deployment.sh IMAGE_PREFIX sha-<commit>" >&2
  echo "Build auth, catalog, both migrations and gateway images locally before running this check." >&2
  exit 1
fi

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
export IMAGE_PREFIX=$1 IMAGE_TAG=$2
export HTTP_PORT=18080 HTTPS_PORT=18443
export CATALOG_DB_NAME=homestock_catalog
export HOMESTOCK_HOME
HOMESTOCK_HOME=$(mktemp -d)
export HOMESTOCK_SKIP_PULL=1
export HOMESTOCK_PROJECT
HOMESTOCK_PROJECT="homestock-ci-$(date +%s)-$$"

export SWAGGER_USERNAME=ci-docs
SWAGGER_PASSWORD=$(openssl rand -hex 24)
export SWAGGER_PASSWORD
SWAGGER_PASSWORD_HASH=$(printf '%s\n' "$SWAGGER_PASSWORD" | docker run --rm -i "$IMAGE_PREFIX/gateway:$IMAGE_TAG" caddy hash-password --algorithm bcrypt --bcrypt-cost 10)
export SWAGGER_PASSWORD_HASH

bash "$script_dir/vps/init-env.sh" "$HOMESTOCK_HOME" http://localhost "$IMAGE_PREFIX"
compose=(docker compose --project-name "$HOMESTOCK_PROJECT" --env-file "$HOMESTOCK_HOME/.env" -f "$script_dir/vps/compose.yml")
catalog_smoke_container="$HOMESTOCK_PROJECT-catalog-smoke"
catalog_smoke_binary="$HOMESTOCK_HOME/smoke-catalog"
cleanup() {
  status=$?
  if [[ $status -ne 0 ]]; then
    "${compose[@]}" logs --tail=100 || true
  fi
  docker rm -f "$catalog_smoke_container" >/dev/null 2>&1 || true
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

assert_private_catalog() {
  local container_id published_ports
  container_id=$("${compose[@]}" ps --quiet catalog-service)
  published_ports=$(docker inspect --format '{{json (index .NetworkSettings.Ports "9090/tcp")}}' "$container_id")
  if [[ "$published_ports" != null ]]; then
    echo "Catalog gRPC must remain private; found published ports: $published_ports" >&2
    return 1
  fi
}

smoke_catalog() {
  local container_exit_code
  docker create --name "$catalog_smoke_container" \
    --network "${HOMESTOCK_PROJECT}_default" \
    --entrypoint /tmp/smoke-catalog \
    "$IMAGE_PREFIX/catalog-service:$IMAGE_TAG" catalog-service:9090 >/dev/null
  docker cp "$catalog_smoke_binary" "$catalog_smoke_container:/tmp/smoke-catalog"
  docker start --attach "$catalog_smoke_container"
  container_exit_code=$(docker inspect --format '{{.State.ExitCode}}' "$catalog_smoke_container")
  docker rm "$catalog_smoke_container" >/dev/null
  if [[ "$container_exit_code" != 0 ]]; then
    echo "Catalog smoke client exited with status $container_exit_code." >&2
    return 1
  fi
}

(cd "$script_dir/.." && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$catalog_smoke_binary" ./deploy/smoke-catalog.go)
chmod 755 "$catalog_smoke_binary"

legacy_postgres_override="$HOMESTOCK_HOME/auth-only-postgres.yml"
cat > "$legacy_postgres_override" <<'YAML'
services:
  postgres:
    environment:
      CATALOG_DB_NAME: ""
YAML
"${compose[@]}" -f "$legacy_postgres_override" up -d --no-build --pull never --wait --wait-timeout 120 postgres

bash "$script_dir/vps/deploy.sh" "$IMAGE_TAG"
assert_limits postgres 536870912 805306368 1024
assert_limits auth-service 134217728 201326592 1024
assert_limits catalog-service 134217728 201326592 1024
assert_limits gateway 134217728 201326592 512
assert_private_catalog
go run "$script_dir/smoke-auth.go" http://localhost:18080 "$HOMESTOCK_HOME/account.json"

# shellcheck disable=SC2016
"${compose[@]}" exec -T postgres sh -c \
  'exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1"' sh "$CATALOG_DB_NAME" <<'SQL'
INSERT INTO products (id, gtin, name, brand, category, package_quantity, package_unit, image_url)
VALUES ('11111111-1111-4111-8111-111111111111', '04601234567893', 'Catalog smoke product',
        'Smoke brand', 'Smoke category', 12.345, 'ml', 'https://example.com/catalog-smoke.jpg');
INSERT INTO products (id, gtin, name)
VALUES ('22222222-2222-4222-8222-222222222222', '03017620422003', 'Catalog smoke optional product');
SQL
smoke_catalog

"${compose[@]}" stop postgres
if "${compose[@]}" exec -T catalog-service /app/catalog-service -healthcheck; then
  echo "Catalog health check must fail while PostgreSQL is stopped." >&2
  exit 1
fi
"${compose[@]}" up -d --no-build --pull never --wait --wait-timeout 120 postgres
"${compose[@]}" exec -T catalog-service /app/catalog-service -healthcheck

catalog_container=$("${compose[@]}" ps --quiet catalog-service)
"${compose[@]}" stop catalog-service
if [[ $(docker inspect --format '{{.State.ExitCode}}' "$catalog_container") != 0 ]]; then
  echo "Catalog did not stop cleanly." >&2
  exit 1
fi
if ! docker logs "$catalog_container" 2>&1 | grep -F 'gRPC server stopped gracefully' >/dev/null; then
  echo "Catalog did not confirm graceful shutdown." >&2
  exit 1
fi
"${compose[@]}" up -d --no-build --no-deps --pull never --wait --wait-timeout 120 catalog-service

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
docker tag "$IMAGE_PREFIX/catalog-service:$IMAGE_TAG" "$IMAGE_PREFIX/catalog-service:$updated_tag"
docker tag "$IMAGE_PREFIX/catalog-migrations:$IMAGE_TAG" "$IMAGE_PREFIX/catalog-migrations:$updated_tag"
docker tag "$IMAGE_PREFIX/gateway:$IMAGE_TAG" "$IMAGE_PREFIX/gateway:$updated_tag"
bash "$script_dir/vps/deploy.sh" "$updated_tag"
export IMAGE_TAG=$updated_tag
assert_limits postgres 536870912 805306368 1024
assert_limits auth-service 134217728 201326592 1024
assert_limits catalog-service 134217728 201326592 1024
assert_limits gateway 134217728 201326592 512
assert_private_catalog
postgres_container_after=$("${compose[@]}" ps --quiet postgres)
postgres_image_after=$(docker inspect --format '{{.Image}}' "$postgres_container_after")
if [[ "$postgres_container_before" == "$postgres_container_after" || "$postgres_image_before" != "$postgres_image_after" ]]; then
  echo "Database limits must be applied by replacing its container while retaining its image." >&2
  exit 1
fi
go run "$script_dir/smoke-auth.go" -reuse http://localhost:18080 "$HOMESTOCK_HOME/account.json"
smoke_catalog
gzip -t "$HOMESTOCK_HOME"/backups/*.sql.gz
echo "Fresh deployment, private catalog RPC, health recovery, graceful shutdown and redeployment with preserved data passed."
