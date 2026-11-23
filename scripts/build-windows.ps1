$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$frontendDir = Join-Path $repoRoot "frontend"
$outputDir = Join-Path $repoRoot "dist"

Push-Location $frontendDir
try {
    npm install
    if ($LASTEXITCODE -ne 0) { throw "npm install failed with exit code $LASTEXITCODE" }
    npm run build
    if ($LASTEXITCODE -ne 0) { throw "frontend build failed with exit code $LASTEXITCODE" }
}
finally {
    Pop-Location
}

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Push-Location $repoRoot
try {
    go build -tags desktop -o (Join-Path $outputDir "TBoxDrive.exe") ./cmd/tboxdrive-desktop
    if ($LASTEXITCODE -ne 0) { throw "Windows desktop build failed with exit code $LASTEXITCODE" }
    go build -o (Join-Path $outputDir "TBoxDrive-Web.exe") ./cmd/tboxdrive
    if ($LASTEXITCODE -ne 0) { throw "Windows web build failed with exit code $LASTEXITCODE" }
}
finally {
    Pop-Location
}

Write-Host "Build complete:"
Write-Host "  $outputDir\TBoxDrive.exe"
Write-Host "  $outputDir\TBoxDrive-Web.exe"
