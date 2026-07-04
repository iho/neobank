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

echo "==> cardproc simulator migrations"
cd /svc/cardproc && /usr/local/bin/migrate-cardproc

echo "==> kyc simulator migrations"
cd /svc/kyc && /usr/local/bin/migrate-kyc

echo "==> fx simulator migrations"
cd /svc/fx && /usr/local/bin/migrate-fx

echo "all service migrations applied"