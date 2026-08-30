#!/usr/bin/env bash
# startup.sh — Build and launch Universal Cleaner.
#
# Usage:
#   ./startup.sh          # build + run the GUI
#   ./startup.sh --cli    # run the CLI version (pass --path, --dry-run, etc.)
#   ./startup.sh --build  # build only, don't run
#   ./startup.sh --docker # build & run inside Docker (Linux only)

set -euo pipefail
cd "$(dirname "$0")"

APP_NAME="universal-cleaner"
BINARY="./${APP_NAME}"

echo "========================================"
echo "  Universal Cleaner — Startup"
echo "========================================"

# Parse arguments
MODE="gui"
BUILD_ONLY=false
USE_DOCKER=false
CLI_ARGS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --cli)
            MODE="cli"
            shift
            # Collect remaining args for CLI
            CLI_ARGS=("$@")
            break
            ;;
        --build)
            BUILD_ONLY=true
            shift
            ;;
        --docker)
            USE_DOCKER=true
            shift
            ;;
        *)
            CLI_ARGS+=("$1")
            shift
            ;;
    esac
done

# ─── Docker mode ────────────────────────────────────────────────────────────
if $USE_DOCKER; then
    echo "[Docker] Building and running in container..."
    if ! command -v docker &>/dev/null; then
        echo "Error: docker is not installed."
        exit 1
    fi
    docker build -t universal-cleaner -f docker/Dockerfile .
    if $BUILD_ONLY; then
        echo "[Docker] Build complete."
        exit 0
    fi
    # X11 forwarding for GUI
    if [[ -n "${DISPLAY:-}" ]]; then
        xhost +local:docker 2>/dev/null || true
        docker run --rm -e DISPLAY="$DISPLAY" \
            -v /tmp/.X11-unix:/tmp/.X11-unix \
            universal-cleaner
    else
        echo "[Docker] No DISPLAY set — running CLI mode."
        docker run --rm universal-cleaner --path /home
    fi
    exit 0
fi

# ─── Local build ────────────────────────────────────────────────────────────
echo "[1/2] Checking dependencies..."
if ! command -v go &>/dev/null; then
    echo "Error: Go is not installed. Install from https://go.dev/dl/"
    exit 1
fi

echo "[2/2] Building ${APP_NAME}..."
go build -ldflags="-s -w" -o "$BINARY" .

if $BUILD_ONLY; then
    echo ""
    echo "Build complete: $BINARY"
    ls -lh "$BINARY"
    exit 0
fi

echo ""
if [[ "$MODE" == "cli" ]]; then
    echo "Running CLI mode with args: ${CLI_ARGS[*]:-}"
    "$BINARY" "${CLI_ARGS[@]}"
else
    echo "Launching GUI..."
    "$BINARY"
fi
