#!/bin/bash
# Build script for TaskVault

set -e

echo "🔨 TaskVault Build"
echo "=================="

# Ensure Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go 1.21+ required but not found"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✓ Go $GO_VERSION"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Run tests
echo "🧪 Running tests..."
go test ./... -v -race

# Build CLI
echo "🏗️  Building CLI..."
go build -o taskvault ./cmd/taskvault
echo "✓ Built: ./taskvault"

# Print build info
echo ""
echo "Build successful!"
echo "Run: ./taskvault --help"
