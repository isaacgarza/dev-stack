#!/bin/bash

# Get Go version from .go-version file
# This script serves as the single source of truth for Go version across the project

set -euo pipefail

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
GO_VERSION_FILE="$PROJECT_ROOT/.go-version"

# Check if .go-version file exists
if [[ ! -f "$GO_VERSION_FILE" ]]; then
    echo "Error: .go-version file not found at $GO_VERSION_FILE" >&2
    exit 1
fi

# Read and validate the version
GO_VERSION=$(cat "$GO_VERSION_FILE" | tr -d '[:space:]')

if [[ -z "$GO_VERSION" ]]; then
    echo "Error: .go-version file is empty" >&2
    exit 1
fi

# Validate version format (basic check for x.y or x.y.z)
if [[ ! "$GO_VERSION" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
    echo "Error: Invalid Go version format in .go-version: $GO_VERSION" >&2
    echo "Expected format: x.y or x.y.z (e.g., 1.21 or 1.21.0)" >&2
    exit 1
fi

# Output based on the requested format
case "${1:-}" in
    --major-minor)
        # Output just major.minor (e.g., 1.21)
        echo "$GO_VERSION" | sed -E 's/^([0-9]+\.[0-9]+).*/\1/'
        ;;
    --full)
        # Output full version with patch if available
        echo "$GO_VERSION"
        ;;
    --env)
        # Output as environment variable format
        echo "GO_VERSION=$GO_VERSION"
        ;;
    --github-matrix)
        # Output as JSON array for GitHub Actions matrix
        # Only include current version since go.mod specifies minimum version
        MAJOR_MINOR=$(echo "$GO_VERSION" | sed -E 's/^([0-9]+\.[0-9]+).*/\1/')
        echo "['$MAJOR_MINOR']"
        ;;
    --help|-h)
        cat << EOF
Usage: $0 [option]

Get Go version from .go-version file in the project root.

Options:
    (no option)      Output the version as-is from .go-version
    --major-minor    Output only major.minor version (e.g., 1.21)
    --full           Output full version (same as no option)
    --env            Output as environment variable (GO_VERSION=x.y.z)
    --github-matrix  Output as JSON array for GitHub Actions matrix
    --help, -h       Show this help message

Examples:
    $0                    # 1.21
    $0 --major-minor      # 1.21
    $0 --env              # GO_VERSION=1.21
    $0 --github-matrix    # ['1.19', '1.20', '1.21']

The .go-version file should contain just the version number (e.g., "1.21" or "1.21.0").
EOF
        ;;
    "")
        # Default: output the version as-is
        echo "$GO_VERSION"
        ;;
    *)
        echo "Error: Unknown option '$1'" >&2
        echo "Use --help for usage information" >&2
        exit 1
        ;;
esac
