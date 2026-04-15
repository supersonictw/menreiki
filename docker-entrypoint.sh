#!/bin/sh

set -e
INSTANCES="${PM2_INSTANCES:-8}"

INTERPRETER="/usr/local/bin/node"
ENTRYPOINT="/workplace/app.ts"

echo "Starting App server with pm2-runtime (instances: $INSTANCES)..."
exec pm2-runtime start "$ENTRYPOINT" -i "$INSTANCES" --interpreter "$INTERPRETER" --name app
