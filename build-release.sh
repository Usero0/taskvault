#!/usr/bin/env bash
# Cross-platform build script for TaskVault
# Generates binaries for Windows, Linux, and macOS

set -e

VERSION="0.1.0"
BUILD_DIR="dist"
BINARY_NAME="taskvault"

echo "🚀 TaskVault Multi-Platform Build"
echo "=================================="
echo "Version: $VERSION"
echo ""

# Create build directory
mkdir -p "$BUILD_DIR"

# Build for Windows (amd64)
echo "🔨 Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o "$BUILD_DIR/${BINARY_NAME}-windows-amd64.exe" ./cmd/taskvault
echo "✓ $BUILD_DIR/${BINARY_NAME}-windows-amd64.exe"

# Build for Linux (amd64)
echo "🔨 Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o "$BUILD_DIR/${BINARY_NAME}-linux-amd64" ./cmd/taskvault
echo "✓ $BUILD_DIR/${BINARY_NAME}-linux-amd64"

# Build for macOS (amd64)
echo "🔨 Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o "$BUILD_DIR/${BINARY_NAME}-darwin-amd64" ./cmd/taskvault
echo "✓ $BUILD_DIR/${BINARY_NAME}-darwin-amd64"

# Build for macOS (arm64)
echo "🔨 Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o "$BUILD_DIR/${BINARY_NAME}-darwin-arm64" ./cmd/taskvault
echo "✓ $BUILD_DIR/${BINARY_NAME}-darwin-arm64"

echo ""
echo "✅ Build complete! Binaries in $BUILD_DIR/"
echo ""
ls -lh "$BUILD_DIR"
