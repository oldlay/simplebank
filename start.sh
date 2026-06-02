#!/bin/sh

# script exit when commends return with a non-zero status
set -e

echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start it up"
exec "$@"