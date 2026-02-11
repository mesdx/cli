#!/usr/bin/env bash
# Build MesDX binary and parser libraries together

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

echo "🔨 Building MesDX with embedded parsers..."
echo ""

# 1. Initialize grammar submodules
echo "📦 Initializing grammar submodules..."
bash scripts/setup-grammars.sh
echo ""

# 2. Build parser libraries
echo "🔧 Building parser libraries..."
bash scripts/build-parsers.sh
echo ""

# 3. Set parser directory for build
export MESDX_PARSER_DIR="${PROJECT_ROOT}/dist/parsers"

# 4. Build Go binary
echo "🔧 Building Go binary..."
go build -o dist/mesdx ./cmd/mesdx
echo ""

echo "✅ Build complete!"
echo ""
echo "Binary: dist/mesdx"
echo "Parsers: dist/parsers/"
echo ""
echo "To install locally:"
echo "  make install"
echo ""
echo "To run from dist:"
echo "  export MESDX_PARSER_DIR=\$(pwd)/dist/parsers"
echo "  ./dist/mesdx --version"
