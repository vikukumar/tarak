#!/usr/bin/env bash
set -e

# 1. Clean old embedded UI artifacts
rm -rf internal/ui/dist/*

# 2. Build Dashboard UI
(cd dashboard && npm run build)

# 3. Copy clean Next.js export into internal/ui/dist
mkdir -p internal/ui/dist
cp -r dashboard/out/* internal/ui/dist/

# 4. Sync web assets into docs
if [ -d "web/public" ]; then
  mkdir -p docs
  cp -r web/public/* docs/ 2>/dev/null || true
fi

# 5. Compile all 5 Go Binaries
mkdir -p bin
go build -trimpath -o bin/tarak ./cmd/tarak
go build -trimpath -o bin/tarakctl ./cmd/tarakctl
go build -trimpath -o bin/taraktl ./cmd/taraktl
go build -trimpath -o bin/tarakd ./cmd/tarakd
go build -trimpath -o bin/taraks ./cmd/taraks

echo "✅ Docs, UI & All 5 Binaries Built Successfully!"
