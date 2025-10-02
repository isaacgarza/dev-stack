#!/bin/bash
set -e

# GitHub Actions Workflow Migration Script
# This script migrates from the complex multi-workflow setup to simplified workflows

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKFLOWS_DIR="$PROJECT_ROOT/.github/workflows"
SIMPLIFIED_DIR="$PROJECT_ROOT/.github/workflows-simplified"
BACKUP_DIR="$PROJECT_ROOT/.github/workflows-backup"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the right directory
check_environment() {
    log_info "Checking environment..."

    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        log_error "This doesn't appear to be a Go project root (no go.mod found)"
        exit 1
    fi

    if [ ! -d "$PROJECT_ROOT/.github" ]; then
        log_error "No .github directory found"
        exit 1
    fi

    if [ ! -d "$SIMPLIFIED_DIR" ]; then
        log_error "Simplified workflows directory not found: $SIMPLIFIED_DIR"
        exit 1
    fi

    log_success "Environment check passed"
}

# Show current workflow status
show_status() {
    log_info "Current workflow status:"
    echo

    if [ -d "$WORKFLOWS_DIR" ]; then
        echo "📁 Current workflows:"
        ls -la "$WORKFLOWS_DIR"/*.yml 2>/dev/null | while read -r line; do
            echo "  $line"
        done
    else
        echo "  No workflows directory found"
    fi

    echo
    if [ -d "$SIMPLIFIED_DIR" ]; then
        echo "📁 Simplified workflows available:"
        ls -la "$SIMPLIFIED_DIR"/*.yml 2>/dev/null | while read -r line; do
            echo "  $line"
        done
    fi

    echo
    if [ -d "$BACKUP_DIR" ]; then
        echo "📁 Backup directory exists:"
        ls -la "$BACKUP_DIR"/*.yml 2>/dev/null | while read -r line; do
            echo "  $line"
        done
    else
        echo "📁 No backup directory found"
    fi
}

# Backup existing workflows
backup_workflows() {
    log_info "Backing up existing workflows..."

    if [ ! -d "$WORKFLOWS_DIR" ]; then
        log_warning "No workflows directory to backup"
        return 0
    fi

    # Create backup directory
    mkdir -p "$BACKUP_DIR"

    # Backup existing workflows
    if ls "$WORKFLOWS_DIR"/*.yml >/dev/null 2>&1; then
        for file in "$WORKFLOWS_DIR"/*.yml; do
            if [ -f "$file" ]; then
                filename=$(basename "$file")
                cp "$file" "$BACKUP_DIR/$filename"
                log_success "Backed up: $filename"
            fi
        done
    else
        log_warning "No YAML workflow files found to backup"
    fi

    # Create backup info file
    cat > "$BACKUP_DIR/backup-info.txt" << EOF
Workflow Backup Created: $(date)
Original Location: $WORKFLOWS_DIR
Backup Location: $BACKUP_DIR
Migration Script: $0

To restore these workflows:
1. Copy files back: cp $BACKUP_DIR/*.yml $WORKFLOWS_DIR/
2. Remove simplified workflows: rm $WORKFLOWS_DIR/{ci,security,validation,release}.yml
3. Update GitHub branch protection rules to use original job names

Files backed up:
$(ls -1 "$BACKUP_DIR"/*.yml 2>/dev/null || echo "None")
EOF

    log_success "Backup completed in: $BACKUP_DIR"
}

# Install simplified workflows
install_simplified() {
    log_info "Installing simplified workflows..."

    # Ensure workflows directory exists
    mkdir -p "$WORKFLOWS_DIR"

    # Copy simplified workflows
    for file in "$SIMPLIFIED_DIR"/*.yml; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            cp "$file" "$WORKFLOWS_DIR/$filename"
            log_success "Installed: $filename"
        fi
    done

    log_success "Simplified workflows installed"
}

# Remove old workflows
cleanup_old() {
    log_info "Cleaning up old workflow files..."

    # List of old workflow files to remove
    old_files=(
        "test.yml"
        "security.yml.bak"
        "validate-commits.yml"
        "docs.yml"
        "build-release.yml"
        "release-please.yml"
        "update-packages.yml"
    )

    for file in "${old_files[@]}"; do
        if [ -f "$WORKFLOWS_DIR/$file" ]; then
            rm "$WORKFLOWS_DIR/$file"
            log_success "Removed old workflow: $file"
        fi
    done
}

# Validate installation
validate_installation() {
    log_info "Validating installation..."

    required_files=(
        "ci.yml"
        "security.yml"
        "validation.yml"
        "release.yml"
    )

    all_good=true
    for file in "${required_files[@]}"; do
        if [ -f "$WORKFLOWS_DIR/$file" ]; then
            log_success "Found: $file"
        else
            log_error "Missing: $file"
            all_good=false
        fi
    done

    # Check if setup-go-version action exists
    if [ -f "$PROJECT_ROOT/.github/actions/setup-go-version/action.yml" ]; then
        log_success "Found: setup-go-version action"
    else
        log_error "Missing: .github/actions/setup-go-version/action.yml"
        all_good=false
    fi

    if [ "$all_good" = true ]; then
        log_success "Installation validation passed"
        return 0
    else
        log_error "Installation validation failed"
        return 1
    fi
}

# Show migration summary
show_summary() {
    echo
    log_info "🎉 Migration Summary"
    echo "=================="
    echo
    echo "✅ Old workflows backed up to: $BACKUP_DIR"
    echo "✅ New simplified workflows installed in: $WORKFLOWS_DIR"
    echo
    echo "📋 Next Steps:"
    echo "1. Test the new workflows with a pull request"
    echo "2. Update GitHub branch protection rules:"
    echo "   - Remove old required checks"
    echo "   - Add new required checks: ci, security, validation"
    echo "3. Monitor the first few runs to ensure everything works"
    echo
    echo "🔧 Recommended Required Status Checks:"
    echo "   Minimal:     ci"
    echo "   Recommended: ci, security, validation"
    echo "   Maximum:     ci, security, validation, test-matrix, integration"
    echo
    echo "📖 For detailed information, see:"
    echo "   .github/workflows-simplified/README.md"
    echo
    echo "🔄 To rollback:"
    echo "   cp $BACKUP_DIR/*.yml $WORKFLOWS_DIR/"
    echo "   rm $WORKFLOWS_DIR/{ci,security,validation,release}.yml"
}

# Main execution
main() {
    echo "🚀 GitHub Actions Workflow Migration"
    echo "===================================="
    echo

    # Parse command line arguments
    FORCE=false
    DRY_RUN=false
    SKIP_BACKUP=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                FORCE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --skip-backup)
                SKIP_BACKUP=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo
                echo "Options:"
                echo "  --force        Force migration without confirmation"
                echo "  --dry-run      Show what would be done without making changes"
                echo "  --skip-backup  Skip backing up existing workflows"
                echo "  --help, -h     Show this help message"
                echo
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    # Check environment
    check_environment

    # Show current status
    show_status

    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN MODE - No changes will be made"
        echo
        echo "Would perform the following actions:"
        echo "1. Backup existing workflows to: $BACKUP_DIR"
        echo "2. Install simplified workflows from: $SIMPLIFIED_DIR"
        echo "3. Clean up old workflow files"
        echo "4. Validate installation"
        exit 0
    fi

    # Confirmation prompt
    if [ "$FORCE" = false ]; then
        echo
        read -p "🤔 Do you want to proceed with the migration? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Migration cancelled"
            exit 0
        fi
    fi

    echo
    log_info "Starting migration..."

    # Perform migration steps
    if [ "$SKIP_BACKUP" = false ]; then
        backup_workflows
    else
        log_warning "Skipping backup as requested"
    fi

    install_simplified
    cleanup_old

    if validate_installation; then
        show_summary
        exit 0
    else
        log_error "Migration completed but validation failed"
        log_error "Please check the installation manually"
        exit 1
    fi
}

# Run main function
main "$@"
