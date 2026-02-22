#!/usr/bin/env bash
# Extracts the changelog section for a given version from CHANGELOG.md.
# Usage: ./scripts/extract-changelog.sh <version>
# Example: ./scripts/extract-changelog.sh v0.3.0
#          ./scripts/extract-changelog.sh v0.3.0-rc1  (uses [Unreleased] section)
#
# For pre-release versions (tags containing 'rc' or 'test'), the [Unreleased]
# section is used as the release body.
# For stable versions, a matching '## [X.Y.Z]' section must exist or the
# script exits with code 1, intentionally blocking the release job.
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

RAW_VERSION="$1"
# Strip leading 'v' so both 'v0.3.0' and '0.3.0' work
VERSION="${RAW_VERSION#v}"

# For pre-release tags (rc, test) use the [Unreleased] section
if [[ "$RAW_VERSION" =~ (rc|test) ]]; then
  LOOKUP="Unreleased"
else
  LOOKUP="$VERSION"
fi

# Extract lines between "## [LOOKUP]" (header skipped) and the next "## [" header.
SECTION=$(awk \
  -v lookup="$LOOKUP" \
  'BEGIN { found=0 }
   /^## \[/ && found { exit }
   /^## \[/ && $0 ~ "\\[" lookup "\\]" { found=1; next }
   found { print }' \
  "$CHANGELOG")

# Strip leading/trailing blank lines
SECTION=$(echo "$SECTION" | sed -e '/./,$!d' -e 's/[[:space:]]*$//')

if [[ -z "$SECTION" ]]; then
  if [[ "$LOOKUP" == "Unreleased" ]]; then
    echo "ERROR: The [Unreleased] section in CHANGELOG.md is empty." >&2
    echo "       Add notes to [Unreleased] before pushing a pre-release tag." >&2
  else
    echo "ERROR: No changelog entry found for version '$VERSION' in CHANGELOG.md" >&2
    echo "       Add a '## [$VERSION] - YYYY-MM-DD' section before tagging." >&2
  fi
  exit 1
fi

echo "$SECTION"
