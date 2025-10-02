#!/bin/bash
# scripts/generate-release-configs.sh
# Generates all release configuration files from central config

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_FILE="$PROJECT_ROOT/.github/config/release-config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}🔧${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    print_error "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Check if yq is available
if ! command -v yq &> /dev/null; then
    print_error "yq is required but not installed."
    echo "Install with:"
    echo "  macOS: brew install yq"
    echo "  Linux: wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq && chmod +x /usr/bin/yq"
    echo "  Go: go install github.com/mikefarah/yq/v4@latest"
    exit 1
fi

print_status "Generating release configuration files from $CONFIG_FILE"

# Generate commitlint configuration
print_status "Generating .commitlintrc.json..."

# Extract commit types as a JSON array with proper formatting
COMMIT_TYPES=$(yq eval '.commit_types[].type' "$CONFIG_FILE" | jq -R . | jq -s . | jq -c .)
HEADER_MAX_LENGTH=$(yq eval '.commit_lint.header_max_length // 72' "$CONFIG_FILE")

# Generate properly formatted commitlint config
jq -n \
  --argjson commit_types "$COMMIT_TYPES" \
  --argjson header_max "$HEADER_MAX_LENGTH" \
  '{
    "extends": ["@commitlint/config-conventional"],
    "rules": {
      "type-enum": [2, "always", $commit_types],
      "subject-empty": [2, "never"],
      "subject-full-stop": [2, "never", "."],
      "header-max-length": [2, "always", $header_max]
    }
  }' > "$PROJECT_ROOT/.commitlintrc.json"

# Generate release-please configuration
print_status "Generating .release-please-config.json..."

PACKAGE_NAME=$(yq eval '.release.package_name' "$CONFIG_FILE")
RELEASE_TYPE=$(yq eval '.release.release_type' "$CONFIG_FILE")
CHANGELOG_SECTIONS=$(yq eval '.commit_types | map({"type": .type, "section": .section, "hidden": .hidden})' "$CONFIG_FILE" -o=json -I=0)
EXTRA_FILES=$(yq eval '.release.extra_files' "$CONFIG_FILE" -o=json -I=0)

# Generate properly formatted release-please config
jq -n \
  --arg release_type "$RELEASE_TYPE" \
  --arg package_name "$PACKAGE_NAME" \
  --argjson changelog_sections "$CHANGELOG_SECTIONS" \
  --argjson extra_files "$EXTRA_FILES" \
  '{
    "release-type": $release_type,
    "packages": {
      ".": {
        "package-name": $package_name,
        "changelog-sections": $changelog_sections,
        "extra-files": $extra_files
      }
    }
  }' > "$PROJECT_ROOT/.release-please-config.json"



# Generate GitHub Actions validation workflow if it doesn't exist
VALIDATE_WORKFLOW="$PROJECT_ROOT/.github/workflows/validate-commits.yml"
if [[ ! -f "$VALIDATE_WORKFLOW" ]]; then
    print_status "Generating commit validation workflow..."

    mkdir -p "$(dirname "$VALIDATE_WORKFLOW")"

    cat > "$VALIDATE_WORKFLOW" << 'EOF'
name: Validate Commits

on:
  pull_request:
    branches: [main, develop]

jobs:
  conventional-commits:
    name: Validate Conventional Commits
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Validate Conventional Commits
        uses: wagoid/commitlint-github-action@v5
        with:
          configFile: .commitlintrc.json

  config-sync:
    name: Validate Config Files are Current
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v4

      - name: Install yq
        run: |
          sudo wget https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 -O /usr/bin/yq
          sudo chmod +x /usr/bin/yq

      - name: Install jq
        run: sudo apt-get update && sudo apt-get install -y jq

      - name: Validate configuration files are up to date
        run: |
          chmod +x scripts/generate-release-configs.sh
          ./scripts/generate-release-configs.sh

          if git diff --exit-code .commitlintrc.json .release-please-config.json; then
            echo "✅ Release configuration files are up to date"
          else
            echo "❌ Release configuration files are out of date"
            echo "The following files need to be updated:"
            git diff --name-only .commitlintrc.json .release-please-config.json
            echo ""
            echo "Run 'make generate-release-configs' to update them"
            exit 1
          fi
EOF
fi

print_success "Configuration files generated successfully!"
echo ""
print_status "📚 Available commit types:"
yq eval '.commit_types[] | "  " + .type + " - " + .description' "$CONFIG_FILE"
echo ""
print_status "🔧 Next steps:"
echo "  1. git add . && git commit -m 'ci: add release automation configuration'"
echo "  2. git push"
echo ""
print_warning "Note: Package managers are currently disabled in config. Enable them when ready to create taps/buckets."
