# Dockerfile for building Universal Cleaner Linux packages.
# Uses a Debian-based image with all GUI build dependencies pre-installed.
#
# Usage:
#   docker build -t universal-cleaner-builder -f docker/Dockerfile .
#   docker run --rm -v "$(pwd)/dist:/dist" universal-cleaner-builder
#
# The built .tar.xz and .deb packages will appear in ./dist/

FROM golang:1.23-bookworm AS builder

# Install GUI build dependencies (Fyne requires OpenGL/GLFW headers).
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-dev \
    libegl1-mesa-dev \
    libgles2-mesa-dev \
    libx11-dev \
    xorg-dev \
    gcc \
    pkg-config \
    upx-ucl \
    && rm -rf /var/lib/apt/lists/*

# Install the fyne CLI for packaging.
RUN go install fyne.io/fyne/v2/cmd/fyne@latest

WORKDIR /src

# Copy go.mod/go.sum first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source.
COPY . .

# Build the raw binary.
RUN go build -ldflags="-s -w" -o universal-cleaner .

# Package as .tar.xz and .deb using fyne.
RUN fyne package -os linux -icon Icon.png -name "universal-cleaner" \
    -appVersion "docker" -appID "com.universalcleaner.app"

# Move artifacts to /dist for the host to pick up.
RUN mkdir -p /dist && \
    mv *.tar.xz /dist/ 2>/dev/null || true && \
    mv *.deb /dist/ 2>/dev/null || true && \
    cp universal-cleaner /dist/ 2>/dev/null || true

# ─── Output stage ──────────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1 \
    libegl1 \
    libgles2 \
    libx11-6 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /dist /dist
COPY --from=builder /src/universal-cleaner /usr/local/bin/universal-cleaner

ENTRYPOINT ["universal-cleaner"]
