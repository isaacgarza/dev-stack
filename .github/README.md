# GitHub Actions CI/CD Pipeline

This directory contains the GitHub Actions workflows that automate testing, security scanning, documentation generation, and releases for the dev-stack project.

## Workflows Overview

### 🧪 Test Workflow (`test.yml`)

**Triggers:** Pull requests and pushes to `main` and `develop` branches

**Features:**
- **Matrix Testing**: Tests across multiple Go versions and platforms (Ubuntu, macOS, Windows)
- **Unit Tests**: Runs `go test` with race detection and coverage reporting
- **Linting**: Uses `golangci-lint` with project-specific configuration
- **Build Validation**: Builds binaries for current and all supported platforms
- **Integration Tests**: Runs integration tests when available
- **Code Quality**: Validates Go modules, formatting, and vet checks

**Artifacts:**
- Test coverage reports
- Build artifacts for all platforms

### 🔒 Security Workflow (`security.yml`)

**Triggers:** Pull requests, pushes to main branches, and weekly scheduled runs

**Security Scans:**
- **CodeQL Analysis**: GitHub's semantic code analysis for Go
- **Dependency Scanning**: Uses `govulncheck` and `nancy` for vulnerability detection
- **Secrets Detection**: Uses Gitleaks to scan for exposed secrets
- **Go Security**: Uses `gosec` for Go-specific security issues
- **Container Scanning**: Uses Trivy to scan Docker images for vulnerabilities
- **License Compliance**: Uses `go-licenses` to check dependency licenses

**Outputs:**
- SARIF reports uploaded to GitHub Security tab
- License compliance reports
- Security scan summaries

### 📚 Documentation Workflow (`docs.yml`)

**Triggers:** Changes to documentation, markdown files, or source code

**Features:**
- **Markdown Linting**: Validates documentation formatting
- **Link Checking**: Verifies all links in documentation work
- **CLI Documentation**: Auto-generates CLI help and completion scripts
- **Example Validation**: Validates code examples in documentation
- **Site Generation**: Builds documentation site using MkDocs
- **GitHub Pages**: Deploys documentation to GitHub Pages

**Outputs:**
- Generated CLI documentation
- Documentation site deployed to GitHub Pages

### 🚀 Release Workflow (`release-please.yml`)

**Triggers:** Pushes to main branch and manual workflow dispatch

**Release Process:**
1. **Release Please**: Uses conventional commits to automatically determine version bumps and generate changelogs
2. **Build and Release**: When a release is created, triggers `build-release.yml` workflow
3. **Cross-platform Builds**: Builds binaries for all supported platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
4. **Checksum Generation**: Creates SHA256 checksums for all binaries
5. **GitHub Release**: Creates release with auto-generated notes
6. **Docker Images**: Builds and pushes multi-arch Docker images to GHCR
7. **Package Managers**: Placeholder for future Homebrew/Scoop integration

**Artifacts:**
- Platform-specific binaries
- SHA256 checksums
- Docker images
- Automated changelogs

## Configuration Files

### Dependabot (`dependabot.yml`)
- **Go Modules**: Weekly updates for Go dependencies
- **GitHub Actions**: Weekly updates for workflow actions
- **Docker**: Weekly updates for base images

### Issue Templates
- **Bug Reports**: Structured template for reporting issues
- **Feature Requests**: Template for suggesting new features

### PR Template
- Comprehensive checklist covering code quality, testing, security, and documentation
- Guidelines for different types of changes

## Security Configuration

### Code Scanning
- **CodeQL**: Configured for Go with security-extended queries
- **Dependency Scanning**: Automated vulnerability detection
- **Secrets Scanning**: Prevents credential leaks
- **Container Security**: Scans Docker images for vulnerabilities

### Access Control
- Workflows use least-privilege permissions
- Secrets are properly scoped and managed
- Release process requires proper authentication

## Development Workflow

### Local Development
```bash
# Run tests locally
make test-go

# Run linting
make lint-go

# Build for current platform
make build

# Build for all platforms
make build-all
```

### Pull Request Process
1. Create feature branch
2. Make changes with conventional commits
3. Push branch (triggers test workflow)
4. Create pull request (triggers full CI pipeline)
5. Address any failing checks
6. Merge after approval

### Release Process
1. Ensure all changes are merged to `main` using conventional commits
2. Push to main triggers Release Please to:
   - Analyze commits for version bump
   - Create/update release PR with changelog
3. Merge the release PR to automatically:
   - Create GitHub release with generated notes
   - Run tests and build binaries
   - Build and push Docker images

## Monitoring and Maintenance

### Workflow Health
- All workflows include summary reporting
- Failed runs are clearly highlighted
- Artifacts are retained for debugging

### Dependency Management
- Dependabot ensures dependencies stay current
- Security vulnerabilities are automatically detected
- License compliance is monitored

### Performance
- Go module caching reduces build times
- Matrix builds run in parallel
- Artifacts are efficiently stored and retrieved

## Troubleshooting

### Common Issues

**Test Failures:**
- Check the test workflow logs
- Ensure code follows Go conventions
- Verify all dependencies are properly declared

**Security Scan Failures:**
- Review the Security tab for detailed findings
- Update vulnerable dependencies
- Fix any detected security issues

**Build Failures:**
- Ensure code compiles on all target platforms
- Check for platform-specific dependencies
- Verify build configuration in Makefile

**Release Issues:**
- Ensure tag follows semantic versioning
- Check that all required workflows pass
- Verify GitHub token permissions

### Getting Help

1. Check workflow logs in the Actions tab
2. Review the Security tab for security findings
3. Check the Issues tab for known problems
4. Create a new issue using the appropriate template

## Future Enhancements

- [ ] Package manager integration (Homebrew, Scoop) - framework ready
- [ ] Performance benchmarking in CI
- [ ] Release notifications (Slack, Discord)
- [ ] Integration with external services
- [ ] Advanced deployment strategies