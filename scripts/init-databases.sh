#!/bin/sh
(
set -eu

for database_name in \
    "$AUTH_DB_NAME" \
    "$INVENTORY_DB_NAME" \
    "$CATALOG_DB_NAME" \
    "$SHOPPING_DB_NAME" \
    "$NOTIFICATION_DB_NAME"; do
    psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
        -v ON_ERROR_STOP=1 -v database_name="$database_name" <<'SQL'
SELECT format('CREATE DATABASE %I', :'database_name')
WHERE NOT EXISTS (
    SELECT 1 FROM pg_database WHERE datname = :'database_name'
) \gexec
SQL
done
)
