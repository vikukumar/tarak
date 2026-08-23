param (
    [Parameter(Mandatory = $true)]
    [string]$NewVersion
)

$CleanVer = $NewVersion.TrimStart('v')
$Tag = "v$CleanVer"
$BuildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$DateOnly = (Get-Date).ToString("yyyy-MM-dd")

Write-Host "🔄 Bumping Tarak version across all components to $Tag ($CleanVer)..." -ForegroundColor Cyan

# 1. Update Go internal/version/version.go
$VersionGoPath = ".\internal\version\version.go"
if (Test-Path $VersionGoPath) {
    (Get-Content $VersionGoPath) -replace 'Version\s*=\s*"[^"]+"', "Version = `"$CleanVer`"" | Set-Content $VersionGoPath
    Write-Host "  ✔ Updated $VersionGoPath" -ForegroundColor Green
}

# 2. Update dashboard/package.json
if (Test-Path ".\dashboard\package.json") {
    npm --prefix dashboard version $CleanVer --no-git-tag-version --allow-same-version
    Write-Host "  ✔ Updated dashboard/package.json to $CleanVer" -ForegroundColor Green
}

# 3. Update web/package.json
if (Test-Path ".\web\package.json") {
    npm --prefix web version $CleanVer --no-git-tag-version --allow-same-version
    Write-Host "  ✔ Updated web/package.json to $CleanVer" -ForegroundColor Green
}

# 4. Update docs/data/releases.json & web/public/data/releases.json
$ReleasesJsonPath = ".\docs\data\releases.json"
if (Test-Path $ReleasesJsonPath) {
    $releases = Get-Content $ReleasesJsonPath | ConvertFrom-Json
    # Mark old latest as false
    foreach ($r in $releases) {
        $r.isLatest = $false
        if ($r.status -eq "Latest (Production Ready)") {
            $r.status = "Stable"
        }
    }
    # Prepend new version if not exists
    $existing = $releases | Where-Object { $_.version -eq $CleanVer }
    if (-not $existing) {
        $newRelease = [PSCustomObject]@{
            version = $CleanVer
            tag = $Tag
            status = "Latest (Production Ready)"
            isLatest = $true
            date = $DateOnly
            name = "Tarak $Tag — Production Ready Release"
            highlights = @(
                "Automated multi-platform binary compilation and release delivery",
                "Full Kubernetes REST API and service mesh parity",
                "Hardware telemetry and live container streaming"
            )
            binaries = @(
                "tarak-windows-amd64.exe",
                "tarak-windows-arm64.exe",
                "tarak-linux-amd64",
                "tarak-linux-arm64",
                "tarak-darwin-amd64",
                "tarak-darwin-arm64"
            )
            downloadUrl = "https://github.com/vikukumar/tarak/releases/tag/$Tag"
        }
        $releases = @($newRelease) + $releases
    } else {
        $existing.isLatest = $true
        $existing.status = "Latest (Production Ready)"
        $existing.date = $DateOnly
    }
    $releases | ConvertTo-Json -Depth 10 | Set-Content $ReleasesJsonPath
    if (Test-Path ".\web\public\data") {
        $releases | ConvertTo-Json -Depth 10 | Set-Content ".\web\public\data\releases.json"
    }
    Write-Host "  ✔ Updated $ReleasesJsonPath and web/public/data/releases.json" -ForegroundColor Green
}

Write-Host "✅ Version synchronization complete! Run .\scripts\build.ps1 to compile all artifacts." -ForegroundColor Cyan
