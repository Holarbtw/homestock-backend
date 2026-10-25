#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 || ! "$1" =~ ^sha-[0-9a-f]{40}$ ]]; then
  echo "Usage: bash remote-deploy.sh sha-<commit> GITHUB_ACTOR" >&2
  exit 1
fi

release_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

export DOCKER_CONFIG
DOCKER_CONFIG=$(mktemp -d)
trap 'docker logout ghcr.io >/dev/null 2>&1 || true; rm -f -- "$DOCKER_CONFIG/config.json"; rmdir -- "$DOCKER_CONFIG"' EXIT
docker login ghcr.io --username "$2" --password-stdin
bash "$release_dir/deploy.sh" "$1"
