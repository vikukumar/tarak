# 1. Clean old embedded UI artifacts
Remove-Item -Path ".\internal\ui\dist\*" -Recurse -Force -ErrorAction SilentlyContinue

# 2. Build the new Next.js Dashboard
cd dashboard
npm run build
cd ..

# 3. Copy clean Next.js export into internal/ui/dist
Copy-Item -Path ".\dashboard\out\*" -Destination ".\internal\ui\dist" -Recurse -Force

# 4. Compile all 5 Go Binaries
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarak.exe ./cmd/tarak
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarakctl.exe ./cmd/tarakctl
& "C:\Program Files\Go\bin\go.exe" build -o bin/taraktl.exe ./cmd/taraktl
& "C:\Program Files\Go\bin\go.exe" build -o bin/tarakd.exe ./cmd/tarakd
& "C:\Program Files\Go\bin\go.exe" build -o bin/taraks.exe ./cmd/taraks

Write-Host "✅ UI & All 5 Binaries Built Successfully!" -ForegroundColor Cyan
