#!/bin/sh

set -e
INSTANCES="${PM2_INSTANCES:-8}"

echo "Starting App server with pm2-runtime (instances: $INSTANCES)..."
exec pm2-runtime start /workplace/app.js -i "$INSTANCES" --name app
