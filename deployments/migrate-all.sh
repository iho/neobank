#!/bin/sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

echo "==> user migrations"
cd /svc/user && /usr/local/bin/migrate-user

echo "==> payment migrations"
cd /svc/payment && /usr/local/bin/migrate-payment

echo "==> notification migrations"
cd /svc/notification && /usr/local/bin/migrate-notification

echo "==> card migrations"
cd /svc/card && /usr/local/bin/migrate-card

echo "==> rails simulator migrations"
cd /svc/rails && /usr/local/bin/migrate-rails

echo "all service migrations applied"