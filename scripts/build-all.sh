#!/bin/bash
# Full build pipeline: Go backend + Python service + Frontend + Electron packaging
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."

echo "=== Step 1: Building Go backend ==="
bash "$SCRIPT_DIR/build-backend.sh"

echo ""
echo "=== Step 2: Building Python service ==="
cd "$ROOT_DIR/backend/python_service"
if command -v pyinstaller &> /dev/null; then
    pyinstaller --onefile --noconsole app.py
    cp dist/app "$ROOT_DIR/builds/a-share-python-service"
else
    echo "WARNING: pyinstaller not installed. Skipping Python service bundling."
    echo "Install with: pip install pyinstaller"
fi

echo ""
echo "=== Step 3: Building Frontend ==="
cd "$ROOT_DIR/frontend"
npm ci
npm run build

echo ""
echo "=== Step 4: Packaging Electron ==="
npx electron-builder --config electron-builder.config.ts --mac --publish never

echo ""
echo "=== Build Complete ==="
echo "Artifacts in: $ROOT_DIR/frontend/release/"
