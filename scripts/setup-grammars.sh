#!/usr/bin/env bash
# Initialize tree-sitter grammar submodules

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

echo "Initializing tree-sitter grammar submodules..."

# Check if .gitmodules exists
if [[ ! -f ".gitmodules" ]]; then
    echo "ERROR: .gitmodules not found" >&2
    exit 1
fi

# Initialize and update submodules
git submodule update --init --recursive third_party/

echo ""
echo "✓ Grammar submodules initialized"
echo ""
echo "Submodules:"
git submodule status third_party/
