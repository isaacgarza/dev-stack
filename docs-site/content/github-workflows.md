---
title: "GitHub Actions Workflows"
description: "Documentation for CI/CD workflows and automation"
lead: "Understand the GitHub Actions workflows that power dev-stack development"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 85
toc: true
---

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
- Centralized environment setup (Go, Task, tools)
- Dependency validation (`go mod tidy`)
- Code quality (`gofmt`, `go vet`, `golangci-lint`)
- Unit tests with coverage
- Build verification using Task runner

### ✅ Validation (`validation.yml`)

**Triggers**: Pull Requests, Push to `main`/`develop`
**Purpose**: Code quality and documentation validation

**Checks:**
- Conventional commit compliance (PRs only)
- Centralized project validation via `scripts/validate-project.sh`
- Markdown linting
- Link checking
- **Hugo validation** (using configurable Hugo version):
  - Configuration syntax validation
  - Content structure verification
  - Build testing (dry run)
  - Internal link validation
  - Frontmatter syntax checking

### 📚 Documentation (`pages.yml`)

**Triggers**: Push to `main` (content changes), Manual dispatch
**Purpose**: GitHub Pages deployment

**Process:**
1. Load centralized configuration for Hugo version and paths
2. Build dev-stack CLI binary using Task runner
3. Generate CLI documentation (or use placeholder)
4. Hugo site build with configurable version
5. Deploy to GitHub Pages

**Requirements:**
- Hugo Extended (version from centralized config)
- PaperMod theme (git submodule)
- Content structure validation
- Valid Hugo configuration

### 🛡️ Security (`security.yml`)

**Triggers**: Push/PR to `main`/`develop`, Weekly schedule, Manual dispatch
**Purpose**: Security vulnerability scanning

**Scans:**
- **Gosec**: Go security scanner with configurable settings
- **Govulncheck**: Go vulnerability database scanner
- **TruffleHog**: Secrets scanning (diff-based for PRs, full for scheduled)
- **Basic checks**: Hardcoded secrets and unsafe function detection
- **CodeQL**: Advanced security analysis (scheduled or labeled PRs)

### 🚀 Release (`release.yml`)

**Triggers**: Push to `main`, Manual dispatch
**Purpose**: Multi-platform binary distribution and Docker images

**Process:**
1. Create release using release-please
2. Build multi-platform binaries using Task runner
3. Generate checksums and verify artifacts
4. Build and push Docker images to configurable registry
5. Upload release assets to GitHub

**Outputs:**
- Platform-specific binaries (Linux, macOS, Windows for amd64/arm64)
- SHA256 checksums
- Docker images (multi-platform)
- GitHub release with changelog

## Configuration

### Centralized Configuration

All workflows use `.github/config/workflow-config.yml` for consistent settings:

```yaml
versions:
  hugo: "0.151.0"
  node: "18"

paths:
  build_dir: "build"
  docs_dir: "docs-site"
  cli_binary: "dev-stack"

docker:
  registry: "ghcr.io"
  platforms: "linux/amd64,linux/arm64"
```

### Reusable Actions

- **`setup-environment`**: Common setup for Go, Task, and tools
- **`load-config`**: Loads centralized configuration values
- **`setup-go-version`**: Go setup with caching from `.go-version`

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

**Reusable Actions**:
- **`setup-environment`**: Centralized setup for Go, Task runner, and development tools
- **`load-config`**: Loads configuration from `.github/config/workflow-config.yml`
- **`setup-go-version`**: Reads Go version from `.go-version` with caching

**Validation Script** (`scripts/validate-project.sh`):
- Consolidates validation logic from workflows
- Configuration file validation
- Hugo site and content validation
- Code quality checks
- Reusable across different contexts

**Dependabot** (`dependabot.yml`):
- Weekly Go module updates
- Monthly GitHub Actions updates
- Automatic security patches

## Local Development

**Reproduce CI locally:**
```bash
task test              # Unit tests with coverage
task lint              # Linting and static analysis
task build             # Build verification

# Project validation (matches CI)
./scripts/validate-project.sh  # Complete validation suite

# Hugo validation (if working with docs)
task validate-docs     # Complete Hugo validation suite
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
- Run `task validate-docs` before pushing changes
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
