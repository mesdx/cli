#!/usr/bin/env bash
# Initialize tree-sitter grammar submodules

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
THIRD_PARTY_DIR="${PROJECT_ROOT}/third_party"

cd "${PROJECT_ROOT}"

echo "Initializing tree-sitter grammar submodules..."

# Check if .gitmodules exists
if [[ ! -f ".gitmodules" ]]; then
    echo "ERROR: .gitmodules not found" >&2
    exit 1
fi

# Check if third_party directory exists
if [[ ! -d "${THIRD_PARTY_DIR}" ]]; then
    echo "Third party directory doesn't exist yet. Creating and adding submodules..."
    mkdir -p "${THIRD_PARTY_DIR}"
    
    # Add submodules if they don't exist
    GRAMMARS=(
        "tree-sitter-go:https://github.com/tree-sitter/tree-sitter-go.git"
        "tree-sitter-java:https://github.com/tree-sitter/tree-sitter-java.git"
        "tree-sitter-rust:https://github.com/tree-sitter/tree-sitter-rust.git"
        "tree-sitter-python:https://github.com/tree-sitter/tree-sitter-python.git"
        "tree-sitter-javascript:https://github.com/tree-sitter/tree-sitter-javascript.git"
        "tree-sitter-typescript:https://github.com/tree-sitter/tree-sitter-typescript.git"
    )
    
    for grammar in "${GRAMMARS[@]}"; do
        IFS=':' read -r name url <<< "$grammar"
        grammar_path="third_party/${name}"
        if [[ ! -d "${grammar_path}" ]]; then
            echo "  Adding ${name}..."
            git submodule add --force "${url}" "${grammar_path}" || true
        fi
    done
fi

# Initialize and update submodules
git submodule update --init --recursive

echo ""
echo "✓ Grammar submodules initialized"
echo ""
echo "Submodules:"
git submodule status
