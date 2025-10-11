#!/bin/bash
set -e

echo "🗓️ Updating lastmod dates for changed documentation files..."

# Get today's date in yyyy-MM-dd format
TODAY=$(date +%Y-%m-%d)

# Get list of changed files in docs-site/content
CHANGED_FILES=$(git diff --name-only HEAD docs-site/content/*.md 2>/dev/null || true)

if [ -z "$CHANGED_FILES" ]; then
    echo "✅ No documentation files changed"
    exit 0
fi

echo "📝 Found changed files:"
echo "$CHANGED_FILES"

# Update lastmod date for each changed file
for file in $CHANGED_FILES; do
    if [ -f "$file" ]; then
        echo "🔄 Updating lastmod in $file"
        # Use sed to update the lastmod line
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS sed
            sed -i '' "s/^lastmod: .*/lastmod: \"$TODAY\"/" "$file"
        else
            # Linux sed
            sed -i "s/^lastmod: .*/lastmod: \"$TODAY\"/" "$file"
        fi
    fi
done

echo "✅ Updated lastmod dates to $TODAY"
