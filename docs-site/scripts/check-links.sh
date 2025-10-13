#!/bin/bash
set -e

echo "Checking markdown links..."

# Check all markdown files and capture output
LINK_CHECK_OUTPUT=$(find content -name '*.md' -exec markdown-link-check {} \; && markdown-link-check README.md 2>&1)
echo "$LINK_CHECK_OUTPUT"

# Check if there were any dead links
if echo "$LINK_CHECK_OUTPUT" | grep -q "ERROR:.*dead links found"; then
    echo ""
    echo "❌ Link check failed: Dead links found!"
    exit 1
fi

echo "✅ All links are valid"
