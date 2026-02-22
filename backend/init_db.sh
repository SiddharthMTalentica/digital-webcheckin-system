#!/bin/sh
set -e

# Wait for Postgres
echo "Waiting for postgres at $DB_HOST:$DB_PORT..."
until nc -z "$DB_HOST" "$DB_PORT"; do
  echo "Retrying connection to $DB_HOST:$DB_PORT..."
  sleep 1
done
echo "Postgres started"

# Run Migrations
echo "Running Migrations..."
./migrate

echo "Running application..."
exec ./main
