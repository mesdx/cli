#!/usr/bin/env bash
# Build tree-sitter parser libraries for MesDX

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
THIRD_PARTY_DIR="${PROJECT_ROOT}/third_party"
DIST_DIR="${PROJECT_ROOT}/dist/parsers"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
esac

# Determine library extension
case "$OS" in
    Darwin)
        LIB_EXT="dylib"
        CFLAGS="-O2 -fPIC -dynamiclib"
        ;;
    Linux)
        LIB_EXT="so"
        CFLAGS="-O2 -fPIC -shared"
        ;;
    *)
        echo "Unsupported OS: $OS" >&2
        exit 1
        ;;
esac

echo "Building parsers for ${OS}-${ARCH}..."
echo "Output directory: ${DIST_DIR}"

# Create dist directory
mkdir -p "${DIST_DIR}"

# Language grammars to build
# Format: "language:grammar-dir:source-path"
GRAMMARS=(
    "go:tree-sitter-go:src/parser.c"
    "java:tree-sitter-java:src/parser.c"
    "rust:tree-sitter-rust:src/parser.c"
    "python:tree-sitter-python:src/parser.c"
    "javascript:tree-sitter-javascript:src/parser.c"
    "typescript:tree-sitter-typescript/typescript:src/parser.c"
)

build_grammar() {
    local lang_name="$1"
    local grammar_dir="$2"
    local source_path="$3"
    
    local grammar_path="${THIRD_PARTY_DIR}/${grammar_dir}"
    local lib_name="libtree-sitter-${lang_name}.${LIB_EXT}"
    local output_path="${DIST_DIR}/${lib_name}"
    
    echo "Building ${lang_name}..."
    
    if [[ ! -d "$grammar_path" ]]; then
        echo "  ERROR: Grammar directory not found: $grammar_path" >&2
        echo "  Please run: git submodule update --init --recursive" >&2
        return 1
    fi
    
    local parser_c="${grammar_path}/${source_path}"
    if [[ ! -f "$parser_c" ]]; then
        echo "  ERROR: Parser source not found: $parser_c" >&2
        return 1
    fi
    
    # Check for scanner.c (some grammars have an external scanner)
    local scanner_c="${grammar_path}/src/scanner.c"
    local scanner_cc="${grammar_path}/src/scanner.cc"
    
    local sources="$parser_c"
    local cc_flags="$CFLAGS"
    
    if [[ -f "$scanner_c" ]]; then
        sources="$sources $scanner_c"
        echo "  Including scanner.c"
    elif [[ -f "$scanner_cc" ]]; then
        sources="$sources $scanner_cc"
        cc_flags="$cc_flags -xc++"
        echo "  Including scanner.cc"
    fi
    
    # Compile
    local include_dir="${grammar_path}/src"
    
    if [[ -f "$scanner_cc" ]]; then
        # Use C++ compiler for grammars with C++ scanner
        c++ $cc_flags \
            -I"${include_dir}" \
            -o "${output_path}" \
            $sources
    else
        # Use C compiler
        cc $cc_flags \
            -I"${include_dir}" \
            -o "${output_path}" \
            $sources
    fi
    
    echo "  ✓ ${lib_name}"
}

# Build all grammars
failed=0
for grammar_spec in "${GRAMMARS[@]}"; do
    IFS=':' read -r lang_name grammar_dir source_path <<< "$grammar_spec"
    if ! build_grammar "$lang_name" "$grammar_dir" "$source_path"; then
        failed=$((failed + 1))
    fi
done

if [[ $failed -gt 0 ]]; then
    echo ""
    echo "ERROR: $failed grammar(s) failed to build" >&2
    exit 1
fi

echo ""
echo "✓ All parsers built successfully"
echo "Libraries available in: ${DIST_DIR}"
ls -lh "${DIST_DIR}"

# If MESDX_PARSER_DIR is set, also copy there for local development
if [[ -n "${MESDX_PARSER_DIR:-}" ]]; then
    echo ""
    echo "Copying to MESDX_PARSER_DIR: ${MESDX_PARSER_DIR}"
    mkdir -p "${MESDX_PARSER_DIR}"
    cp "${DIST_DIR}"/* "${MESDX_PARSER_DIR}/"
    echo "✓ Parsers copied to ${MESDX_PARSER_DIR}"
fi
