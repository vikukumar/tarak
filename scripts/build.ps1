# 1. Clean old embedded UI artifacts
Remove-Item -Path ".\internal\ui\dist\*" -Recurse -Force -ErrorAction SilentlyContinue

# 2. Build the Next.js Dashboard UI
Set-Location dashboard
npm run build
Set-Location ..

# 3. Build Vite React Documentation Portal
Set-Location web
npm run build
Set-Location ..

# 4. Copy clean Next.js export into internal/ui/dist
Copy-Item -Path ".\dashboard\out\*" -Destination ".\internal\ui\dist" -Recurse -Force

# 5. Sync canonical assets
if (Test-Path ".\docs\assets") {
    Copy-Item -Path ".\docs\assets\*" -Destination ".\web\public\assets" -Recurse -Force -ErrorAction SilentlyContinue
    Copy-Item -Path ".\docs\assets\*" -Destination ".\dashboard\public\assets" -Recurse -Force -ErrorAction SilentlyContinue
}

# 6. Compile all 5 Go Binaries
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarak.exe ./cmd/tarak
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarakctl.exe ./cmd/tarakctl
& "C:\Program Files\Go\bin\go.exe" build -o bin/taraktl.exe ./cmd/taraktl
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarakd.exe ./cmd/tarakd
& "C:\Program Files\Go\bin\go.exe" build -o bin/taraks.exe ./cmd/taraks

Write-Host "✅ Docs, Vite Web Portal, UI & All 5 Binaries Built Successfully!" -ForegroundColor Cyan