#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "Usage: bash init-env.sh HOME_DIR API_DOMAIN IMAGE_PREFIX" >&2
  exit 1
fi

deployment_home=$1
api_domain=$2
image_prefix=$3

if [[ ! "$api_domain" =~ ^[a-zA-Z0-9][a-zA-Z0-9.-]+$ && "$api_domain" != http://localhost ]]; then
  echo "Use a domain without a scheme, or http://localhost for a local check." >&2
  exit 1
fi
if [[ ! "$image_prefix" =~ ^[a-z0-9][a-z0-9./_-]+$ ]]; then
  echo "IMAGE_PREFIX must be a lowercase registry/repository path." >&2
  exit 1
fi

swagger_username=${SWAGGER_USERNAME:-docs}
swagger_password_hash=${SWAGGER_PASSWORD_HASH:-}
if [[ ! "$swagger_username" =~ ^[a-zA-Z0-9_.-]+$ ]]; then
  echo "SWAGGER_USERNAME must contain only letters, digits, dots, underscores or hyphens." >&2
  exit 1
fi
if [[ -n "$swagger_password_hash" && ! "$swagger_password_hash" =~ ^\$2[aby]\$[0-9]{2}\$[./A-Za-z0-9]{53}$ ]]; then
  echo "SWAGGER_PASSWORD_HASH must be a complete bcrypt hash." >&2
  exit 1
fi

umask 077
mkdir -p "$deployment_home"
env_file="$deployment_home/.env"
if [[ -e "$env_file" ]]; then
  echo "$env_file already exists; it has not been changed." >&2
  exit 1
fi

db_password=$(openssl rand -hex 32)
jwt_secret=$(openssl rand -hex 64)
(
  set -o noclobber
  cat > "$env_file" <<EOF
API_DOMAIN=$api_domain
IMAGE_PREFIX=$image_prefix
DB_USER=homestock
AUTH_DB_NAME=homestock_auth
CATALOG_DB_NAME=homestock_catalog
CATALOG_SERVICE_PORT=9090
DB_PASSWORD=$db_password
JWT_SECRET=$jwt_secret
SWAGGER_USERNAME=$swagger_username
SWAGGER_PASSWORD_HASH='$swagger_password_hash'
HTTP_PORT=80
HTTPS_PORT=443
ACCESS_TOKEN_TTL_MINUTES=15
REFRESH_TOKEN_TTL_DAYS=30
BCRYPT_COST=10
EOF
)
echo "Created $env_file with generated secrets. Keep this file between deployments."
if [[ -z "$swagger_password_hash" ]]; then
  echo "Before deployment, set SWAGGER_PASSWORD_HASH in $env_file using caddy hash-password."
fi
