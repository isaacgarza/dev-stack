#!/bin/bash
set -e

# Project validation script
# This script consolidates validation logic to reduce workflow complexity

CONFIG_FILE=".github/config/workflow-config.yml"
DOCS_DIR="docs-site"

echo "🔍 Starting project validation..."

# Load configuration if available
if [ -f "$CONFIG_FILE" ] && command -v yq >/dev/null 2>&1; then
    DOCS_DIR=$(yq eval '.paths.docs_dir' "$CONFIG_FILE" 2>/dev/null || echo "docs-site")
fi

# Function to validate configuration files
validate_configs() {
    echo "🔍 Validating configuration files..."
    
    if [ -f "scripts/generate-release-configs.sh" ]; then
        chmod +x scripts/generate-release-configs.sh
        ./scripts/generate-release-configs.sh

        if git diff --exit-code .commitlintrc.json .release-please-config.json; then
            echo "✅ Release configuration files are up to date"
        else
            echo "❌ Release configuration files are out of date"
            echo "Run 'task generate-release-configs' to update them"
            git diff --name-only .commitlintrc.json .release-please-config.json
            return 1
        fi
    else
        echo "⚠️ Release config generation script not found, skipping validation"
    fi
}

# Function to validate Hugo configuration
validate_hugo() {
    echo "🔍 Validating Hugo configuration..."
    
    if [ ! -f "$DOCS_DIR/config/_default/hugo.toml" ]; then
        echo "❌ Hugo configuration file not found"
        echo "Expected: $DOCS_DIR/config/_default/hugo.toml"
        return 1
    fi

    cd "$DOCS_DIR"
    if ! hugo config > /dev/null 2>&1; then
        echo "❌ Hugo configuration is invalid"
        hugo config
        return 1
    fi
    cd ..

    echo "✅ Hugo configuration is valid"
}

# Function to validate content structure
validate_content() {
    echo "🔍 Validating Hugo content structure..."
    
    # Check required content files
    REQUIRED_FILES=(
        "$DOCS_DIR/content/_index.md"
        "$DOCS_DIR/content/setup.md"
        "$DOCS_DIR/content/usage.md"
        "$DOCS_DIR/content/services.md"
        "$DOCS_DIR/content/configuration.md"
        "$DOCS_DIR/content/contributing.md"
        "$DOCS_DIR/content/troubleshooting.md"
    )

    for file in "${REQUIRED_FILES[@]}"; do
        if [ ! -f "$file" ]; then
            echo "❌ Required content file missing: $file"
            return 1
        fi
    done

    # Check CLI reference structure
    if [ ! -d "$DOCS_DIR/content/cli-reference" ]; then
        echo "⚠️ CLI reference directory not found, will be created during build"
    fi

    # Validate frontmatter in main content files
    echo "🔍 Checking frontmatter syntax..."
    FRONTMATTER_FILES=(
        "$DOCS_DIR/content/_index.md"
        "$DOCS_DIR/content/setup.md"
        "$DOCS_DIR/content/usage.md"
        "$DOCS_DIR/content/services.md"
        "$DOCS_DIR/content/configuration.md"
        "$DOCS_DIR/content/contributing.md"
        "$DOCS_DIR/content/troubleshooting.md"
        "$DOCS_DIR/content/reference.md"
    )

    for file in "${FRONTMATTER_FILES[@]}"; do
        if [ -f "$file" ]; then
            if ! head -n 1 "$file" | grep -q "^---$"; then
                echo "❌ Missing frontmatter in: $file"
                return 1
            fi
        fi
    done

    # Basic frontmatter validation
    cd "$DOCS_DIR"
    if ! hugo list all --source . >/dev/null 2>&1; then
        echo "❌ Invalid frontmatter detected in content files"
        return 1
    fi
    cd ..

    echo "✅ Content structure is valid"
}

# Function to test Hugo build
test_hugo_build() {
    echo "🔍 Testing Hugo build..."
    
    # Create temporary CLI reference if it doesn't exist
    if [ ! -f "$DOCS_DIR/content/cli-reference/index.md" ]; then
        mkdir -p "$DOCS_DIR/content/cli-reference"
        cat > "$DOCS_DIR/content/cli-reference/index.md" << 'EOF'
---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI"
weight: 30
---

# CLI Reference

This page will be automatically generated during deployment.
EOF
    fi

    # Test Hugo build without deploying
    cd "$DOCS_DIR"
    if hugo --gc --minify --destination public-test; then
        echo "✅ Hugo build test successful"
        
        # Check build output
        echo "📊 Build statistics:"
        echo "- HTML files: $(find public-test -name "*.html" | wc -l)"
        echo "- CSS files: $(find public-test -name "*.css" | wc -l)"
        echo "- JS files: $(find public-test -name "*.js" | wc -l)"
        
        # Clean up test build
        rm -rf public-test
    else
        echo "❌ Hugo build test failed"
        return 1
    fi
    cd ..
}

# Function to check code quality
check_code_quality() {
    echo "🔍 Checking code quality..."
    
    # Check for TODO/FIXME comments
    if grep -r "TODO\|FIXME" --include="*.go" --include="*.md" . 2>/dev/null; then
        echo "⚠️ Found TODO/FIXME comments - consider addressing before release"
    else
        echo "✅ No TODO/FIXME comments found"
    fi

    # Check file permissions
    echo "🔍 Checking file permissions..."
    find . -type f -perm /111 ! -path "./.git/*" ! -path "./build/*" ! -path "./scripts/*" ! -name "*.sh" ! -name "dev-stack*" | while read -r file; do
        echo "⚠️ Unexpected executable file: $file"
    done
}

# Main validation flow
main() {
    validate_configs || exit 1
    
    if command -v hugo >/dev/null 2>&1; then
        validate_hugo || exit 1
        validate_content || exit 1
        test_hugo_build || exit 1
    else
        echo "⚠️ Hugo not found, skipping Hugo validation"
    fi
    
    check_code_quality
    
    echo "✅ All validations completed successfully!"
}

# Run main function
main "$@"
