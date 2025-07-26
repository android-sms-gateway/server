#!/bin/sh
# docker-entrypoint.sh

# Immediately exit if any command has a non-zero exit status.
set -e

# Execute DB migrations only when the main application is about to run
if [ "${1:-}" = "/app/app" ]; then
  /app/app db:migrate up
fi

# Execute the main application
exec "$@"