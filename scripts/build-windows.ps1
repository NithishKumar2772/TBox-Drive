$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$frontendDir = Join-Path $repoRoot "frontend"
$outputDir = Join-Path $repoRoot "dist"

Push-Location $frontendDir
try {
    npm install
    npm run build
}
finally {
    Pop-Location
}

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Push-Location $repoRoot
try {
    go build -tags desktop -o (Join-Path $outputDir "TBoxDrive.exe") ./cmd/tboxdrive-desktop
    go build -o (Join-Path $outputDir "TBoxDrive-Web.exe") ./cmd/tboxdrive
}
finally {
    Pop-Location
}

Write-Host "Build complete:"
Write-Host "  $outputDir\TBoxDrive.exe"
Write-Host "  $outputDir\TBoxDrive-Web.exe"
