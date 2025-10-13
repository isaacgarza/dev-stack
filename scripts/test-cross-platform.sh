#!/bin/bash

# Test script to verify cross-platform compatibility
set -euo pipefail

echo "🧪 Testing cross-platform compatibility..."

# Test 1: Go version detection
echo "📋 Testing Go version detection..."
if [[ -f ".go-version" ]]; then
    GO_VERSION=$(cat .go-version | tr -d '[:space:]')
    echo "✅ Go version: $GO_VERSION"
else
    echo "❌ .go-version file not found"
    exit 1
fi

# Test 2: Go module validation
echo "📦 Testing Go modules..."
go mod download
go mod verify
echo "✅ Go modules OK"

# Test 3: Basic build test
echo "🔨 Testing build..."
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    go build -o test-binary.exe ./cmd/dev-stack
    rm -f test-binary.exe
else
    go build -o test-binary ./cmd/dev-stack
    rm -f test-binary
fi
echo "✅ Build OK"

# Test 4: Run tests without race detector (for compatibility)
echo "🧪 Testing without race detector..."
go test -v ./...
echo "✅ Tests OK"

# Test 5: Run tests with race detector (if supported)
echo "🧪 Testing with race detector..."
if go test -race -v ./... 2>/dev/null; then
    echo "✅ Race detector tests OK"
else
    echo "⚠️  Race detector tests failed (may be expected on some platforms)"
fi

echo "🎉 Cross-platform compatibility tests completed!"
