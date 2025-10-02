#!/bin/bash

# Local Security Scan Script for dev-stack
# This script runs the same security checks as the CI pipeline locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the project root
if [ ! -f "go.mod" ] || [ ! -f ".go-version" ]; then
    log_error "This script must be run from the project root directory"
    exit 1
fi

# Create results directory
mkdir -p security-results

log_info "Starting local security scan..."

# Check if required tools are installed
check_tool() {
    local tool="$1"
    local gopath_bin="$(go env GOPATH)/bin"

    if command -v "$tool" &> /dev/null || [ -f "$gopath_bin/$tool" ]; then
        log_success "$tool is available"
        return 0
    else
        log_warning "$tool is not installed. Installing..."
        return 1
    fi
}

# Install tools if needed
install_tools() {
    log_info "Checking and installing security tools..."

    # Set up GOPATH/bin
    local gopath_bin="$(go env GOPATH)/bin"
    export PATH="$gopath_bin:$PATH"

    # Check Go
    if ! check_tool go; then
        log_error "Go is not installed. Please install Go first."
        exit 1
    fi

    # Install Gosec
    if ! check_tool gosec; then
        log_info "Installing Gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi

    # Install govulncheck
    if ! check_tool govulncheck; then
        log_info "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi

    # Install staticcheck
    if ! check_tool staticcheck; then
        log_info "Installing staticcheck..."
        go install honnef.co/go/tools/cmd/staticcheck@latest
    fi
}

# Run Gosec security scan
run_gosec() {
    log_info "Running Gosec security scan..."

    local gopath_bin="$(go env GOPATH)/bin"
    local gosec_cmd="$gopath_bin/gosec"

    if ! [ -f "$gosec_cmd" ]; then
        gosec_cmd="gosec"  # fallback to PATH
    fi

    if [ -f ".gosec.conf" ]; then
        log_info "Using configuration file .gosec.conf"
        "$gosec_cmd" -conf .gosec.conf -fmt text -out security-results/gosec-report.txt ./... || {
            log_warning "Gosec scan completed with issues"
        }
    else
        log_info "Using default Gosec configuration"
        "$gosec_cmd" -fmt text -out security-results/gosec-report.txt ./... || {
            log_warning "Gosec scan completed with issues"
        }
    fi

    if [ -f "security-results/gosec-report.txt" ]; then
        log_success "Gosec scan completed - results saved to security-results/gosec-report.txt"
    else
        log_error "Gosec scan failed to generate results"
    fi
}

# Run vulnerability scan
run_vulnerability_scan() {
    log_info "Running Go vulnerability scan..."

    local gopath_bin="$(go env GOPATH)/bin"
    local govulncheck_cmd="$gopath_bin/govulncheck"

    if ! [ -f "$govulncheck_cmd" ]; then
        govulncheck_cmd="govulncheck"  # fallback to PATH
    fi

    "$govulncheck_cmd" -json ./... > security-results/vulncheck-results.json 2>/dev/null || {
        log_warning "Vulnerability scan completed with issues"
    }

    # Also generate human-readable report
    "$govulncheck_cmd" ./... > security-results/vulncheck-report.txt 2>&1 || {
        log_warning "Vulnerability scan found issues - check security-results/vulncheck-report.txt"
    }

    log_success "Vulnerability scan completed"
}

# Run basic security checks
run_basic_checks() {
    log_info "Running basic security checks..."

    local gopath_bin="$(go env GOPATH)/bin"
    local staticcheck_cmd="$gopath_bin/staticcheck"

    if ! [ -f "$staticcheck_cmd" ]; then
        staticcheck_cmd="staticcheck"  # fallback to PATH
    fi

    # Go vet
    log_info "Running go vet..."
    go vet ./... > security-results/vet-report.txt 2>&1 || {
        log_warning "go vet found issues - check security-results/vet-report.txt"
    }

    # Staticcheck
    log_info "Running staticcheck..."
    "$staticcheck_cmd" ./... > security-results/staticcheck-report.txt 2>&1 || {
        log_warning "staticcheck found issues - check security-results/staticcheck-report.txt"
    }

    # Check for hardcoded secrets patterns
    log_info "Checking for potential hardcoded secrets..."
    if grep -r -E "(password|pwd|secret|key|token)\s*[:=]\s*['\"][^'\"]{8,}" --include="*.go" . > security-results/secrets-check.txt 2>&1; then
        log_warning "Potential hardcoded secrets found - check security-results/secrets-check.txt"
    else
        echo "No obvious hardcoded secrets found" > security-results/secrets-check.txt
        log_success "No hardcoded secrets detected"
    fi

    # Check for unsafe functions
    log_info "Checking for unsafe function usage..."
    if grep -r "unsafe\." --include="*.go" . > security-results/unsafe-check.txt 2>&1; then
        log_warning "Unsafe package usage found - check security-results/unsafe-check.txt"
    else
        echo "No unsafe package usage found" > security-results/unsafe-check.txt
        log_success "No unsafe package usage found"
    fi

    log_success "Basic security checks completed"
}

# Generate summary report
generate_summary() {
    log_info "Generating security summary..."

    cat > security-results/SUMMARY.md << EOF
# Security Scan Summary

Generated on: $(date)
Project: dev-stack

## Scan Results

### Gosec Security Scanner
$(if [ -f "security-results/gosec-report.txt" ]; then
    if grep -q "Issues : 0" security-results/gosec-report.txt 2>/dev/null; then
        echo "✅ **Status**: PASSED - No security issues found"
    else
        echo "⚠️  **Status**: ISSUES DETECTED"
        echo ""
        echo "**Issues found:**"
        grep -E "Severity|Rule|File" security-results/gosec-report.txt 2>/dev/null | head -10 || echo "Check gosec-report.txt for details"
    fi
else
    echo "❌ **Status**: SCAN FAILED"
fi)

### Vulnerability Scanner (govulncheck)
$(if [ -f "security-results/vulncheck-report.txt" ]; then
    if grep -q "No vulnerabilities found" security-results/vulncheck-report.txt 2>/dev/null; then
        echo "✅ **Status**: PASSED - No vulnerabilities found"
    else
        echo "⚠️  **Status**: VULNERABILITIES DETECTED"
        echo ""
        echo "**Check vulncheck-report.txt for details**"
    fi
else
    echo "❌ **Status**: SCAN FAILED"
fi)

### Basic Security Checks

#### Go Vet
$(if [ -s "security-results/vet-report.txt" ]; then
    echo "⚠️  **Issues found** - Check vet-report.txt"
else
    echo "✅ **Passed**"
fi)

#### Static Analysis (staticcheck)
$(if [ -s "security-results/staticcheck-report.txt" ]; then
    echo "⚠️  **Issues found** - Check staticcheck-report.txt"
else
    echo "✅ **Passed**"
fi)

#### Hardcoded Secrets Check
$(if grep -q "No obvious hardcoded secrets found" security-results/secrets-check.txt 2>/dev/null; then
    echo "✅ **Passed** - No hardcoded secrets detected"
else
    echo "⚠️  **Potential issues found** - Check secrets-check.txt"
fi)

#### Unsafe Package Usage
$(if grep -q "No unsafe package usage found" security-results/unsafe-check.txt 2>/dev/null; then
    echo "✅ **Passed** - No unsafe package usage"
else
    echo "⚠️  **Unsafe usage detected** - Check unsafe-check.txt"
fi)

## Files Generated

- \`gosec-report.txt\` - Gosec security analysis report

- \`vulncheck-results.json\` - JSON vulnerability scan results
- \`vulncheck-report.txt\` - Human-readable vulnerability report
- \`vet-report.txt\` - Go vet output
- \`staticcheck-report.txt\` - Staticcheck analysis results
- \`secrets-check.txt\` - Hardcoded secrets check results
- \`unsafe-check.txt\` - Unsafe package usage check results

## Next Steps

1. Review any issues found in the individual report files
2. Fix security issues before committing code
3. Re-run this script to verify fixes

## CI Integration

The same checks run automatically in GitHub Actions. Ensure all issues are resolved locally before pushing.

## Report Format

All reports are generated in human-readable text format for easy review and integration into development workflows.
EOF

    log_success "Security summary generated: security-results/SUMMARY.md"
}

# Main execution
main() {
    local scan_type="${1:-all}"

    case "$scan_type" in
        "gosec")
            install_tools
            run_gosec
            ;;
        "vuln")
            install_tools
            run_vulnerability_scan
            ;;
        "basic")
            install_tools
            run_basic_checks
            ;;
        "all"|"")
            install_tools
            run_gosec
            run_vulnerability_scan
            run_basic_checks
            generate_summary
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [scan_type]"
            echo ""
            echo "Scan types:"
            echo "  all     - Run all security scans (default)"
            echo "  gosec   - Run only Gosec security scanner"
            echo "  vuln    - Run only vulnerability scan"
            echo "  basic   - Run only basic security checks"
            echo "  help    - Show this help message"
            echo ""
            echo "Results are saved to the security-results/ directory"
            exit 0
            ;;
        *)
            log_error "Unknown scan type: $scan_type"
            echo "Run '$0 help' for usage information"
            exit 1
            ;;
    esac

    log_success "Security scan completed! Check security-results/ for detailed reports."

    if [ "$scan_type" = "all" ] || [ "$scan_type" = "" ]; then
        echo ""
        log_info "Quick summary available at: security-results/SUMMARY.md"
        echo ""
        log_info "To view the summary:"
        echo "  cat security-results/SUMMARY.md"
        echo ""
        log_info "To integrate with your IDE, import: security-results/gosec-results.sarif"
    fi
}

# Run main function with all arguments
main "$@"
