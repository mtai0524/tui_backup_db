#!/bin/bash

# bakdb - Quick Run Script
# Runs the database backup manager

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_PATH="$SCRIPT_DIR/build/bakdb"

# Check if binary exists
if [ ! -f "$APP_PATH" ]; then
    echo "📦 Binary not found. Building..."
    cd "$SCRIPT_DIR"
    make build || exit 1
fi

# Run the app
"$APP_PATH" "$@"
