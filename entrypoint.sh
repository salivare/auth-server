#!/usr/bin/env sh
set -euo pipefail

echo "[entrypoint] start"

# Ensure app config exists: prefer mounted /app_config/appexample.yaml, otherwise use bundled one
if [ -f "/app_config/appexample.yaml" ]; then
  CFG_PATH="/app_config/appexample.yaml"
else
  CFG_PATH="/app_config/appexample.yaml"
  # bundled file already copied to /app_config in Dockerfile; if not, fallback to /build_assets path
fi

export CFG_APP_ENV_PATH="${CFG_APP_ENV_PATH:-$CFG_PATH}"
echo "[entrypoint] CFG_APP_ENV_PATH=${CFG_APP_ENV_PATH}"

# Prepare migrator flags (use env vars or defaults)
MIGRATOR_STORAGE_PATH="${MIGRATOR_STORAGE_PATH:-/data/db.sqlite3}"
MIGRATIONS_PATH="${MIGRATIONS_PATH:-/migrations}"

echo "[entrypoint] migrator storage-path=${MIGRATOR_STORAGE_PATH}"
echo "[entrypoint] migrator migrations-path=${MIGRATIONS_PATH}"

# Run migrator (blocking). It must exit 0 on success.
if [ -x /usr/local/bin/migrator ]; then
  echo "[entrypoint] running migrator..."
  /usr/local/bin/migrator --storage-path="${MIGRATOR_STORAGE_PATH}" --migrations-path="${MIGRATIONS_PATH}" || {
    echo "[entrypoint] migrator failed, exiting"
    exit 1
  }
  echo "[entrypoint] migrator finished successfully"
else
  echo "[entrypoint] migrator binary not found, skipping"
fi

# Start main app in foreground
if [ -x /usr/local/bin/sso-auth-server ]; then
  echo "[entrypoint] starting sso-auth-server..."
  exec /usr/local/bin/sso-auth-server
else
  echo "[entrypoint] sso-auth-server binary not found, exiting"
  exit 1
fi
