#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: bash deploy/test-deployment.sh IMAGE_PREFIX sha-<commit>" >&2
  echo "Build both images locally before running this check." >&2
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

bash "$script_dir/vps/deploy.sh" "$IMAGE_TAG"
go run "$script_dir/smoke-auth.go" http://localhost:18080 "$HOMESTOCK_HOME/account.json"

updated_tag="sha-$(printf '%s' "$IMAGE_TAG-update" | sha256sum | cut -c 1-40)"
docker tag "$IMAGE_PREFIX/auth-service:$IMAGE_TAG" "$IMAGE_PREFIX/auth-service:$updated_tag"
docker tag "$IMAGE_PREFIX/auth-migrations:$IMAGE_TAG" "$IMAGE_PREFIX/auth-migrations:$updated_tag"
bash "$script_dir/vps/deploy.sh" "$updated_tag"
export IMAGE_TAG=$updated_tag
go run "$script_dir/smoke-auth.go" -reuse http://localhost:18080 "$HOMESTOCK_HOME/account.json"
gzip -t "$HOMESTOCK_HOME"/backups/*.sql.gz
echo "Fresh deployment and redeployment with preserved data passed."
