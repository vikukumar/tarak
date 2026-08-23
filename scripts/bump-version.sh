#!/usr/bin/env bash
set -e

if [ -z "$1" ]; then
  echo "Usage: ./scripts/bump-version.sh <version> (e.g. 1.0.7 or v1.0.7)"
  exit 1
fi

RAW_VER="$1"
CLEAN_VER="${RAW_VER#v}"
TAG="v$CLEAN_VER"

echo "🔄 Bumping Tarak version across all components to $TAG ($CLEAN_VER)..."

# 1. Update internal/version/version.go
if [ -f "internal/version/version.go" ]; then
  sed -i -E "s/Version = \"[^\"]+\"/Version = \"$CLEAN_VER\"/" internal/version/version.go
  echo "  ✔ Updated internal/version/version.go"
fi

# 2. Update dashboard/package.json
if [ -f "dashboard/package.json" ]; then
  npm --prefix dashboard version "$CLEAN_VER" --no-git-tag-version --allow-same-version
  echo "  ✔ Updated dashboard/package.json"
fi

# 3. Update web/package.json
if [ -f "web/package.json" ]; then
  npm --prefix web version "$CLEAN_VER" --no-git-tag-version --allow-same-version
  echo "  ✔ Updated web/package.json"
fi

echo "✅ Version synchronization complete! Run ./scripts/build.sh to build all artifacts."
