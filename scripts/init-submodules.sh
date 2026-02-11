#!/usr/bin/env bash
# One-time script to add all grammar submodules to the repository
# Run this once, then commit the changes

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

echo "Adding tree-sitter grammar submodules..."
echo ""

# Add submodules
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
    
    if [[ -d "${grammar_path}" ]]; then
        echo "✓ ${name} already exists"
    else
        echo "Adding ${name}..."
        git submodule add "${url}" "${grammar_path}"
    fi
done

echo ""
echo "✓ All grammar submodules added"
echo ""
echo "Next steps:"
echo "1. Commit the changes:"
echo "   git add .gitmodules third_party/"
echo "   git commit -m 'Add tree-sitter grammar submodules'"
echo ""
echo "2. Push to repository:"
echo "   git push"
