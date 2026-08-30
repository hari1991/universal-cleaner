#!/usr/bin/env bash
# stop.sh — Stop any running Universal Cleaner process.

set -euo pipefail

APP_NAME="universal-cleaner"
FOUND=false

echo "Stopping Universal Cleaner..."

# Try to find and kill by process name
if pgrep -x "$APP_NAME" &>/dev/null; then
    pkill -x "$APP_NAME" && FOUND=true
    echo "Stopped $APP_NAME process."
fi

# Also check for the macOS .app bundle process name
if [[ "$(uname)" == "Darwin" ]]; then
    if pgrep -x "UniversalCleaner" &>/dev/null; then
        pkill -x "UniversalCleaner" && FOUND=true
        echo "Stopped UniversalCleaner.app process."
    fi
fi

if ! $FOUND; then
    echo "No running Universal Cleaner process found."
fi

echo "Done."
