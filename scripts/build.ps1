# 1. Clean old embedded UI artifacts
Remove-Item -Path ".\internal\ui\dist\*" -Recurse -Force -ErrorAction SilentlyContinue

# 2. Build the new Next.js Dashboard
cd dashboard
npm run build
cd ..

# 3. Copy clean Next.js export into internal/ui/dist
Copy-Item -Path ".\dashboard\out\*" -Destination ".\internal\ui\dist" -Recurse -Force

# 4. Sync web assets into docs
if (Test-Path ".\web\public") {
    Copy-Item -Path ".\web\public\*" -Destination ".\docs" -Recurse -Force -ErrorAction SilentlyContinue
}

# 5. Compile all 5 Go Binaries
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarak.exe ./cmd/tarak
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarakctl.exe ./cmd/tarakctl
& "C:\Program Files\Go\bin\go.exe" build -o bin/taraktl.exe ./cmd/taraktl
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarakd.exe ./cmd/tarakd
& "C:\Program Files\Go\bin\go.exe" build -o bin/taraks.exe ./cmd/taraks

Write-Host "✅ Docs, UI & All 5 Binaries Built Successfully!" -ForegroundColor Cyan
