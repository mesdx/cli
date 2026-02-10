#!/usr/bin/env bash
# Install MesDX for Linux (ARM64)

set -euo pipefail

VERSION="${1:-latest}"
INSTALL_DIR="${HOME}/.local"
BIN_DIR="${INSTALL_DIR}/bin"
LIB_DIR="${INSTALL_DIR}/lib/mesdx/parsers"

echo "📦 Installing MesDX (Linux ARM64)..."
echo ""

mkdir -p "${BIN_DIR}"
mkdir -p "${LIB_DIR}"

if [ "$VERSION" = "latest" ]; then
    BASE_URL="https://github.com/mesdx/cli/releases/latest/download"
else
    BASE_URL="https://github.com/mesdx/cli/releases/download/${VERSION}"
fi

echo "⬇️  Downloading binary..."
curl -L "${BASE_URL}/mesdx-linux-arm64" -o "${BIN_DIR}/mesdx"
chmod +x "${BIN_DIR}/mesdx"

echo "⬇️  Downloading parsers..."
curl -L "${BASE_URL}/mesdx-parsers-linux-arm64.tar.gz" | tar xz -C "${LIB_DIR}" --strip-components=1

echo ""
echo "✅ MesDX installed successfully!"
echo ""
echo "Binary: ${BIN_DIR}/mesdx"
echo "Parsers: ${LIB_DIR}"
echo ""
echo "Make sure ${BIN_DIR} is in your PATH:"
echo "  export PATH=\"${BIN_DIR}:\$PATH\""
echo ""
echo "Verify installation:"
echo "  mesdx --version"
