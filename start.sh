#!/bin/sh

set -e

# echo "run database migration"
# /app/migrate -path /app/migration -database "$DB_SOURCE" --version up

echo "start the app"
exec "$@"