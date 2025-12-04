# ipmap Multi-Platform Build Script
# Builds for macOS (ARM64 + AMD64) and Linux (AMD64)

$VERSION = "2.0"
$APP_NAME = "ipmap"
$BUILD_DIR = "bin"

Write-Host "Building $APP_NAME v$VERSION for multiple platforms..." -ForegroundColor Cyan
Write-Host ""

# Create build directory
if (-not (Test-Path $BUILD_DIR)) {
    New-Item -ItemType Directory -Path $BUILD_DIR | Out-Null
}

# Build for macOS ARM64
Write-Host "Building for macOS ARM64 (Apple Silicon)..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o "$BUILD_DIR/${APP_NAME}_darwin_arm64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: macOS ARM64 build completed" -ForegroundColor Green
} else {
    Write-Host "ERROR: macOS ARM64 build failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Build for macOS AMD64
Write-Host "Building for macOS AMD64 (Intel)..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o "$BUILD_DIR/${APP_NAME}_darwin_amd64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: macOS AMD64 build completed" -ForegroundColor Green
} else {
    Write-Host "ERROR: macOS AMD64 build failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Build for Linux AMD64
Write-Host "Building for Linux AMD64..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "$BUILD_DIR/${APP_NAME}_linux_amd64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: Linux AMD64 build completed" -ForegroundColor Green
} else {
    Write-Host "ERROR: Linux AMD64 build failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Show file sizes
Write-Host "Build Summary:" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Gray
Get-ChildItem -Path $BUILD_DIR | ForEach-Object {
    $size = "{0:N2} MB" -f ($_.Length / 1MB)
    Write-Host ("{0,-35} {1,10}" -f $_.Name, $size)
}
Write-Host "================================================" -ForegroundColor Gray
Write-Host ""

Write-Host "All builds completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Binaries location: $BUILD_DIR/" -ForegroundColor Cyan

# Reset environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
