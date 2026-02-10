#!/usr/bin/env bash
# Download and setup tree-sitter runtime library header

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
THIRD_PARTY_DIR="${PROJECT_ROOT}/third_party"

# Tree-sitter version (pinned for reproducibility)
TREE_SITTER_VERSION="0.22.6"

echo "Setting up tree-sitter runtime headers..."
echo "Version: ${TREE_SITTER_VERSION}"

# Create third_party directory
mkdir -p "${THIRD_PARTY_DIR}"

# Download tree-sitter header
TS_HEADER_URL="https://raw.githubusercontent.com/tree-sitter/tree-sitter/v${TREE_SITTER_VERSION}/lib/include/tree_sitter/api.h"
HEADER_DEST="${PROJECT_ROOT}/internal/treesitter/tree_sitter_api.h"

echo "Downloading tree-sitter API header..."
curl -L -o "${HEADER_DEST}" "${TS_HEADER_URL}"

if [[ ! -f "${HEADER_DEST}" ]]; then
    echo "ERROR: Failed to download tree-sitter API header" >&2
    exit 1
fi

echo "✓ Tree-sitter API header installed to ${HEADER_DEST}"
