#!/usr/bin/env bash
# Extracts the changelog section for a given version from CHANGELOG.md.
# Usage: ./scripts/extract-changelog.sh <version>
# Example: ./scripts/extract-changelog.sh v0.3.0
#          ./scripts/extract-changelog.sh 0.3.0
#
# Exits with code 1 if no section is found for the given version, which
# intentionally blocks the GitHub Actions release job when CHANGELOG.md
# has not been updated before tagging.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <version>" >&2
  echo "Example: $0 v0.3.0" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHANGELOG="$SCRIPT_DIR/../CHANGELOG.md"

if [[ ! -f "$CHANGELOG" ]]; then
  echo "ERROR: CHANGELOG.md not found at $CHANGELOG" >&2
  exit 1
fi

# Strip leading 'v' so both 'v0.3.0' and '0.3.0' work
VERSION="${1#v}"

# Extract lines between "## [VERSION]" (inclusive header skipped) and the next "## [" header.
# The header line itself is skipped; only the body content is printed.
SECTION=$(awk \
  -v ver="$VERSION" \
  'BEGIN { found=0 }
   /^## \[/ && found { exit }
   /^## \[/ && $0 ~ "\\[" ver "\\]" { found=1; next }
   found { print }' \
  "$CHANGELOG")

# Strip leading/trailing blank lines
SECTION=$(echo "$SECTION" | sed -e '/./,$!d' -e 's/[[:space:]]*$//')

if [[ -z "$SECTION" ]]; then
  echo "ERROR: No changelog entry found for version '$VERSION' in CHANGELOG.md" >&2
  echo "       Add a '## [$VERSION] - YYYY-MM-DD' section before tagging." >&2
  exit 1
fi

echo "$SECTION"
