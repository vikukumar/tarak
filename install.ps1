# Tarak Universal Installer for Windows PowerShell
$ErrorActionPreference = "Stop"
$Repo = "vikukumar/tarak"
$InstallDir = "$env:USERPROFILE\.tarak\bin"

Write-Host "⚡ Installing Tarak for Windows..." -ForegroundColor Cyan

# 1. Detect Architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}

# 2. Get Latest Release Tag
try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $LatestTag = $Release.tag_name
} catch {
    $LatestTag = "v1.0.0"
}
$Version = $LatestTag.TrimStart("v")

$ZipFile = "tarak_${Version}_windows_${Arch}.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$LatestTag/$ZipFile"
$TempZip = "$env:TEMP\$ZipFile"

Write-Host "📦 Downloading $DownloadUrl..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Write-Host "🚀 Extracting binaries to $InstallDir..." -ForegroundColor Yellow
Expand-Archive -Path $TempZip -DestinationPath $InstallDir -Force
Remove-Item $TempZip -Force

# 3. Add to User Path if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", [EnvironmentVariableTarget]::User)
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "✨ Added $InstallDir to your PATH!" -ForegroundColor Green
}

Write-Host "✅ Successfully installed Tarak $LatestTag!" -ForegroundColor Green
Write-Host "   Run 'tarak.exe version' or 'tarakctl.exe version' to get started." -ForegroundColor Cyan
