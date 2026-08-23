#!/usr/bin/env bash
set -e

# 1. Clean old embedded UI artifacts
rm -rf internal/ui/dist/*

# 2. Build Dashboard UI
(cd dashboard && npm run build)

# 3. Build Vite React Documentation Portal
(cd web && npm run build)

# 4. Copy clean Next.js export into internal/ui/dist
mkdir -p internal/ui/dist
cp -r dashboard/out/* internal/ui/dist/

# 5. Sync Vite build into docs/
rm -rf docs/assets docs/data docs/*.html docs/*.js docs/*.css
mkdir -p docs/assets docs/data
cp -r web/dist/* docs/
echo "tarak.vikshro.in" > docs/CNAME
touch docs/.nojekyll

# 6. Compile all 5 Go Binaries
mkdir -p bin
go build -trimpath -o bin/tarak ./cmd/tarak
go build -trimpath -o bin/tarakctl ./cmd/tarakctl
go build -trimpath -o bin/taraktl ./cmd/taraktl
go build -trimpath -o bin/tarakd ./cmd/tarakd
go build -trimpath -o bin/taraks ./cmd/taraks

echo "✅ Docs, Vite Web Portal, UI & All 5 Binaries Built Successfully!"
