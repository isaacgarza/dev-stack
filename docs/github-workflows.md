# GitHub Actions Workflows

This document describes the GitHub Actions workflows that automate testing, validation, documentation, and releases for dev-stack.

## Overview

The workflows support the complete development lifecycle from code validation to release automation:

- **CI**: Core testing and validation
- **Validation**: Code quality and documentation checks
- **Pages**: Documentation deployment
- **Security**: Vulnerability scanning
- **Release**: Multi-platform binary distribution

## Workflows

### 🔄 CI Pipeline (`ci.yml`)

**Triggers**: Push to `main`/`develop`, Pull Requests
**Purpose**: Core continuous integration
**Status**: Required for branch protection

**Jobs:**
- `ci` (required): Go testing, linting, build verification
- `test-matrix` (optional): Cross-platform testing (triggered by `test-matrix` label)
- `integration` (optional): Docker integration tests (triggered by `integration` label)

**Key Steps:**
- Go environment setup from `.go-version`
- Dependency validation (`go mod tidy`)
- Code quality (`gofmt`, `go vet`, `golangci-lint`)
- Unit tests with coverage
- Build verification

### ✅ Validation (`validation.yml`)

**Triggers**: Pull Requests, Push to `main`/`develop`  
**Purpose**: Code quality and documentation validation

**Checks:**
- Conventional commit compliance (PRs only)
- Configuration file validation
- Markdown linting
- Link checking
- **Hugo validation suite**:
  - Configuration syntax validation
  - Content structure verification
  - Build testing (dry run)
  - Internal link validation
  - Frontmatter syntax checking
- TODO/FIXME detection
- File permissions

### 📚 Documentation (`pages.yml`)

**Triggers**: Push to `main` (content changes), Manual dispatch
**Purpose**: GitHub Pages deployment

**Process:**
1. Build dev-stack CLI binary
2. Generate CLI documentation (or use placeholder)
3. Hugo site build with PaperMod theme
4. Deploy to GitHub Pages

**Requirements:**
- Hugo Extended v0.151.0+
- PaperMod theme (git submodule)
- Content structure validation
- Valid Hugo configuration

### 🛡️ Security (`security.yml`)

**Triggers**: Weekly schedule, Manual dispatch
**Purpose**: Security vulnerability scanning

**Scans:**
- Dependency vulnerabilities (GitHub advisories)
- Static code analysis (`gosec`)
- License compliance
- SARIF reporting to GitHub Security tab

### 🚀 Release (`release.yml`)

**Triggers**: Release tags (`v*`), Manual dispatch
**Purpose**: Multi-platform binary distribution

**Builds:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

**Outputs:**
- Platform-specific archives
- SHA256 checksums
- GitHub release with changelog

## Configuration

### Required Status Checks

For branch protection:
- `CI / ci` - Core CI pipeline
- `Validation / validation` - Quality validation

### PR Labels

Control workflow execution:
- `test-matrix` - Cross-platform testing
- `integration` - Integration tests
- `skip-ci` - Skip CI for docs-only changes

### Shared Components

**Setup Go Action** (`.github/actions/setup-go-version`):
- Reads Go version from `.go-version`
- Configures caching for dependencies
- Used across all Go-based workflows

**Dependabot** (`dependabot.yml`):
- Weekly Go module updates
- Monthly GitHub Actions updates
- Automatic security patches

## Local Development

**Reproduce CI locally:**
```bash
make test              # Unit tests with coverage
make lint              # Linting and static analysis
make build             # Build verification

# Hugo validation (if working with docs)
make validate-docs     # Complete Hugo validation suite
hugo config            # Validate Hugo configuration only
hugo --gc --minify --destination public-test  # Test build only
rm -rf public-test     # Clean up
```

**Debug specific issues:**
```bash
# Check formatting
gofmt -l .

# Run linter
golangci-lint run

# Test specific platform
GOOS=linux GOARCH=amd64 go build ./cmd/dev-stack
```

## Troubleshooting

### Common Issues

**Build Failures:**
- Check Go version consistency in `.go-version`
- Run `go mod tidy` to fix dependencies
- Verify code formatting with `gofmt`

**Test Failures:**
- Check for race conditions with `go test -race`
- Ensure proper test cleanup
- Verify test isolation

**Pages Deployment:**
- Run `make validate-docs` before pushing changes
- Ensure Hugo theme submodule is initialized
- Check content file frontmatter syntax
- Validate Hugo configuration with `hugo config`
- Test build locally with validation workflow
- Verify GitHub Pages is enabled in repository settings

**Security Scans:**
- Review findings in GitHub Security tab
- Update vulnerable dependencies
- Check `go.mod` for outdated packages

### Debug Mode

Enable detailed logging:
```yaml
env:
  RUNNER_DEBUG: 1
  ACTIONS_STEP_DEBUG: 1
```

---

For development setup, see [`setup.md`](setup.md).
For contributing guidelines, see [`contributing.md`](contributing.md).
