---
title: "Security Configuration"
description: "Security scanning and configuration options for dev-stack"
lead: "Configure security scanning, vulnerability checks, and best practices"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 70
toc: true
---

# Security Configuration

This document describes the security scanning and configuration options available in the dev-stack project.

## Overview

The dev-stack project implements multiple layers of security scanning and monitoring to ensure code quality and dependency safety:

- **Static Code Analysis** - CodeQL and Gosec for Go code security
- **Dependency Vulnerability Scanning** - Multiple tools for identifying vulnerable dependencies
- **Secret Detection** - Gitleaks for preventing credential leaks
- **Container Security** - Trivy for Docker image vulnerability scanning
- **License Compliance** - Automated license checking for dependencies

## Automated Security Scans

### GitHub Actions Security Workflow

The security workflow (`.github/workflows/security.yml`) runs automatically on:

- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches
- Weekly schedule (Sundays at 2 AM UTC)

### Scan Types

#### 1. CodeQL Analysis

- **Tool**: GitHub CodeQL
- **Purpose**: Static analysis for security vulnerabilities and code quality
- **Languages**: Go
- **Queries**: Security-extended and security-and-quality rule sets
- **Results**: Available in GitHub Security tab

#### 2. Dependency Review

- **Tool**: GitHub Dependency Review Action
- **Purpose**: Reviews dependency changes in pull requests
- **Triggers**: Pull requests only
- **Fail Threshold**: Moderate severity vulnerabilities
- **Allowed Licenses**: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC

#### 3. Dependency Vulnerability Scanning

- **Primary Tool**: Govulncheck (Go official vulnerability scanner)
- **Secondary Tool**: Nancy (OSS Index) - optional with authentication
- **Purpose**: Identify known vulnerabilities in Go dependencies
- **Database**: Go vulnerability database and Sonatype OSS Index

#### 4. Secrets Scanning

- **Tool**: Gitleaks
- **Purpose**: Detect hardcoded secrets and credentials
- **Scope**: Full git history
- **Configuration**: Custom rules for common secret patterns

#### 5. Go Security Analysis

- **Tool**: Gosec
- **Purpose**: Go-specific security issue detection
- **Output**: SARIF format uploaded to GitHub Security tab
- **Rules**: Comprehensive Go security rule set

#### 6. Container Security

- **Tool**: Trivy
- **Purpose**: Docker image vulnerability scanning
- **Scope**: Base images and dependencies
- **Triggers**: Main branch and PR from same repository

#### 7. License Compliance

- **Tool**: go-licenses (Google)
- **Purpose**: License compatibility checking
- **Output**: License report artifact
- **Scope**: All Go dependencies

## Configuration

### OSS Index Authentication (Optional)

Nancy can optionally use Sonatype OSS Index for enhanced vulnerability data. To enable:

1. Create a free account at [OSS Index](https://ossindex.sonatype.org/)
2. Add GitHub repository secrets:
   - `OSS_INDEX_USERNAME`: Your OSS Index username
   - `OSS_INDEX_TOKEN`: Your OSS Index API token

**Without authentication**: Nancy will be skipped, and Govulncheck serves as the primary scanner.

### Vulnerability Ignores

#### Nancy Ignores

Create or edit `.nancy-ignore` to exclude specific vulnerabilities:

```
# Format: one CVE or OSS Index ID per line
CVE-2021-12345
OSS-INDEX-67890
```

#### Dependabot Configuration

Dependency updates are managed by Dependabot (`.github/dependabot.yml`):

- **Schedule**: Weekly on Mondays at 9 AM PST
- **Ecosystems**: Go modules, GitHub Actions, Docker, npm
- **Limits**: Controlled PR limits per ecosystem
- **Auto-assignment**: Assigned to repository maintainer

### License Allowlist

Modify the dependency review action in `.github/workflows/security.yml` to adjust allowed licenses:

```yaml
- name: Dependency Review
  uses: actions/dependency-review-action@v4
  with:
    fail-on-severity: moderate
    allow-licenses: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0
```

## Local Security Testing

### Install Security Tools

```bash
# Install vulnerability scanner
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install Nancy (optional)
go install github.com/sonatype-nexus-community/nancy@latest

# Install Gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install Gitleaks
brew install gitleaks  # macOS
# or download from https://github.com/gitleaks/gitleaks/releases
```

### Run Security Scans

```bash
# Vulnerability scanning
govulncheck ./...

# Nancy with authentication (if configured)
go list -json -deps ./... | nancy sleuth --username $OSS_USERNAME --token $OSS_TOKEN

# Nancy without authentication (rate limited)
go list -json -deps ./... | nancy sleuth

# Static security analysis
gosec ./...

# Secret scanning
gitleaks detect --source . --verbose

# License checking
go-licenses check ./...
go-licenses report ./... > license-report.txt
```

### Build Makefile Integration

```bash
# Run all security checks
task vet

# Individual scans
task lint              # Static code analysis and security checks
go mod download        # Dependency management
go vet ./...          # Go static analysis
```

## Security Best Practices

### Dependency Management

1. **Regular Updates**: Review Dependabot PRs weekly
2. **Vulnerability Response**: Address moderate+ severity issues promptly
3. **License Compliance**: Ensure all dependencies use compatible licenses
4. **Minimal Dependencies**: Regularly audit and remove unused dependencies

### Code Security

1. **Input Validation**: Validate all external inputs
2. **Error Handling**: Avoid exposing sensitive information in errors
3. **Authentication**: Use proper authentication mechanisms
4. **Secrets Management**: Never commit secrets to version control

### Container Security

1. **Base Images**: Use minimal, regularly updated base images
2. **Multi-stage Builds**: Minimize final image size and attack surface
3. **Non-root Users**: Run containers as non-privileged users
4. **Regular Scanning**: Monitor container images for new vulnerabilities

## Incident Response

### Vulnerability Discovered

1. **Assessment**: Evaluate severity and impact
2. **Prioritization**: Critical/High severity issues should be addressed immediately
3. **Patching**: Update affected dependencies or implement mitigations
4. **Testing**: Verify fixes don't break functionality
5. **Documentation**: Update security documentation if needed

### False Positives

1. **Verification**: Confirm the vulnerability doesn't apply to your use case
2. **Documentation**: Document why the vulnerability is not applicable
3. **Ignore Lists**: Add to appropriate ignore files with justification
4. **Review**: Periodically review ignored vulnerabilities

## Monitoring and Alerts

### GitHub Security Alerts

- **Dependabot Alerts**: Automatic alerts for vulnerable dependencies
- **Code Scanning Alerts**: CodeQL and Gosec findings
- **Secret Scanning Alerts**: Detected secrets in code

### Workflow Notifications

Security workflow failures will:

- Block PR merges (for dependency review failures)
- Create GitHub Security tab entries
- Generate workflow run summaries
- Send notifications to assigned reviewers

## Compliance and Reporting

### Security Reports

Weekly security scans generate:

- Vulnerability assessment summaries
- License compliance reports
- Dependency update recommendations
- Security metric trends

### Audit Trail

All security activities are tracked via:

- GitHub Actions workflow logs
- Security tab findings history
- Dependabot update history
- Git commit history for security changes

## Troubleshooting

### Common Issues

#### Nancy Authentication Errors

```
Error: [401 Unauthorized] error accessing OSS Index
```

**Solution**: Add OSS_INDEX_USERNAME and OSS_INDEX_TOKEN secrets, or ignore Nancy failures.

#### Govulncheck Network Issues

```
Error: fetching vulnerability database
```

**Solution**: Check network connectivity and retry. Database updates may be temporarily unavailable.

#### License Check Failures

```
Error: disallowed license found
```

**Solution**: Review dependency licenses and update allowlist or replace dependencies.

#### False Positive Vulnerabilities

**Solution**: Verify applicability, document reasoning, and add to ignore lists if confirmed false positive.

### Getting Help

1. **Documentation**: Check individual tool documentation for specific issues
2. **GitHub Issues**: Report persistent problems or feature requests
3. **Security Team**: Contact security team for critical vulnerabilities
4. **Community**: Use tool-specific communities for complex configuration issues

## References

- [Go Vulnerability Database](https://vuln.go.dev/)
- [Sonatype OSS Index](https://ossindex.sonatype.org/)
- [GitHub Security Features](https://docs.github.com/en/code-security)
- [CodeQL Documentation](https://codeql.github.com/docs/)
- [Gosec Rules](https://github.com/securecodewarrior/gosec#available-rules)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [Gitleaks Configuration](https://github.com/gitleaks/gitleaks)
