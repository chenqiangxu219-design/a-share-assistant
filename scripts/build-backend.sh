#!/bin/bash
# Build Go backend for current platform
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/../builds"
BACKEND_DIR="$SCRIPT_DIR/../backend"

mkdir -p "$BUILD_DIR"

cd "$BACKEND_DIR"

echo "Building Go backend..."
CGO_ENABLED=0 go build -o "$BUILD_DIR/a-share-backend" .

echo "Backend built: $BUILD_DIR/a-share-backend"
ls -la "$BUILD_DIR/a-share-backend"
