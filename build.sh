#!/bin/bash

# ipmap Multi-Platform Build Script
# Builds for macOS (ARM64 + AMD64) and Linux (AMD64 + 386)

VERSION="2.0"
APP_NAME="ipmap"
BUILD_DIR="bin"

echo "🚀 Building $APP_NAME v$VERSION for multiple platforms..."
echo ""

# Create build directory
mkdir -p $BUILD_DIR

# Build for macOS ARM64 (Apple Silicon - M1/M2/M3)
echo "📦 Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o $BUILD_DIR/${APP_NAME}_darwin_arm64 .
if [ $? -eq 0 ]; then
    echo "✅ macOS ARM64 build successful: $BUILD_DIR/${APP_NAME}_darwin_arm64"
else
    echo "❌ macOS ARM64 build failed"
    exit 1
fi
echo ""

# Build for macOS AMD64 (Intel)
echo "📦 Building for macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o $BUILD_DIR/${APP_NAME}_darwin_amd64 .
if [ $? -eq 0 ]; then
    echo "✅ macOS AMD64 build successful: $BUILD_DIR/${APP_NAME}_darwin_amd64"
else
    echo "❌ macOS AMD64 build failed"
    exit 1
fi
echo ""

# Build for Linux AMD64 (x64)
echo "📦 Building for Linux AMD64 (x64)..."
GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/${APP_NAME}_linux_amd64 .
if [ $? -eq 0 ]; then
    echo "✅ Linux AMD64 build successful: $BUILD_DIR/${APP_NAME}_linux_amd64"
else
    echo "❌ Linux AMD64 build failed"
    exit 1
fi
echo ""

# Build for Linux 386 (x86)
echo "📦 Building for Linux 386 (x86)..."
GOOS=linux GOARCH=386 go build -o $BUILD_DIR/${APP_NAME}_linux_386 .
if [ $? -eq 0 ]; then
    echo "✅ Linux 386 build successful: $BUILD_DIR/${APP_NAME}_linux_386"
else
    echo "❌ Linux 386 build failed"
    exit 1
fi
echo ""

# Make binaries executable
chmod +x $BUILD_DIR/*

# Show file sizes
echo "📊 Build Summary:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
ls -lh $BUILD_DIR/ | grep -v "^total" | awk '{printf "%-35s %10s\n", $9, $5}'
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "✅ All builds completed successfully!"
echo ""
echo "📍 Binaries location: $BUILD_DIR/"
echo ""
echo "🎯 Usage:"
echo "  macOS ARM64:  ./$BUILD_DIR/${APP_NAME}_darwin_arm64 --help"
echo "  macOS Intel:  ./$BUILD_DIR/${APP_NAME}_darwin_amd64 --help"
echo "  Linux x64:    ./$BUILD_DIR/${APP_NAME}_linux_amd64 --help"
echo "  Linux x86:    ./$BUILD_DIR/${APP_NAME}_linux_386 --help"
