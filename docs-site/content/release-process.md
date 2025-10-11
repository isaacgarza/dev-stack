---
title: "Release Process"
description: "Automated release process using Release-Please and conventional commits"
lead: "Understand how dev-stack releases are automated and managed"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 95
toc: true
---

# Release Process

This document describes the automated release process for dev-stack using Release-Please and conventional commits.

## Overview

dev-stack uses a fully automated release process that:
- Analyzes conventional commits to determine version bumps
- Generates changelogs automatically
- Creates GitHub releases with binaries and Docker images
- Updates package managers (when enabled)
- Provides full audit trail and rollback capabilities

## Quick Start

### 1. One-Time Setup

```bash
# Install dependencies and generate configuration files
task release-setup
```

This will:
- Check for required tools (yq, jq)
- Generate all release configuration files

### 2. Making Changes

Use conventional commits for all changes:

```bash
# New features
git commit -m "feat: add service discovery integration"
git commit -m "feat(auth): implement OAuth2 flow"

# Bug fixes
git commit -m "fix: resolve timeout issue in Docker startup"
git commit -m "fix(cli): handle missing config file gracefully"

# Breaking changes
git commit -m "feat!: redesign configuration API"
# or
git commit -m "feat: redesign configuration API

BREAKING CHANGE: The configuration file format has changed.
See migration guide for details."
```

### 3. Releasing

Releases happen automatically when you push to `main`:

1. **Push commits** → Release-Please analyzes changes
2. **Release PR created** → Review changelog and version bump
3. **Merge PR** → Release is automatically created

You can also trigger manually via GitHub Actions → Release Please → Run workflow.

## Conventional Commits

### Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description | Changelog | Version Bump |
|------|-------------|-----------|--------------|
| `feat` | New feature | ✅ Features | Minor |
| `fix` | Bug fix | ✅ Bug Fixes | Patch |
| `perf` | Performance improvement | ✅ Performance | Patch |
| `deps` | Dependency updates | ✅ Dependencies | Patch |
| `revert` | Revert previous commit | ✅ Reverts | Patch |
| `docs` | Documentation only | ❌ Hidden | None |
| `style` | Code style changes | ❌ Hidden | None |
| `refactor` | Code refactoring | ❌ Hidden | None |
| `test` | Test changes | ❌ Hidden | None |
| `build` | Build system changes | ❌ Hidden | None |
| `ci` | CI/CD changes | ❌ Hidden | None |
| `chore` | Other maintenance | ❌ Hidden | None |

### Examples

#### Simple commits
```bash
git commit -m "feat: add backup command"
git commit -m "fix: resolve port conflict detection"
git commit -m "docs: update installation instructions"
```

#### With scope
```bash
git commit -m "feat(cli): add interactive mode"
git commit -m "fix(docker): improve container health checks"
git commit -m "perf(config): optimize YAML parsing"
```

#### Breaking changes
```bash
# Option 1: Use ! after type
git commit -m "feat!: change default configuration location"

# Option 2: Use footer
git commit -m "feat: change default configuration location

BREAKING CHANGE: Configuration files are now located in ~/.config/dev-stack/
instead of ~/.dev-stack/. Run 'dev-stack migrate-config' to update."
```

#### Multi-line commits
```bash
git commit -m "feat: add service health monitoring

- Add health check endpoints for all services
- Implement automatic restart on failure
- Add monitoring dashboard integration

Closes #123"
```

## Version Bumping

Release-Please automatically determines version bumps based on commit types:

- **Major (1.0.0 → 2.0.0)**: Breaking changes (`feat!:`, `fix!:`, or `BREAKING CHANGE:` footer)
- **Minor (1.0.0 → 1.1.0)**: New features (`feat:`)
- **Patch (1.0.0 → 1.0.1)**: Bug fixes, performance improvements, etc. (`fix:`, `perf:`)

## Release Workflow

### Automatic Process

1. **Commit Analysis**: Release-Please scans commits since last release
2. **PR Creation**: Creates release PR with:
   - Version bump in relevant files
   - Updated CHANGELOG.md
   - Release notes
3. **Review**: Team reviews the release PR
4. **Merge**: Merging triggers the release process:
   - Builds binaries for all platforms
   - Creates Docker images
   - Uploads to GitHub Releases
   - Updates package managers (if enabled)

### Manual Trigger

You can manually trigger releases via GitHub Actions:

1. Go to **Actions** → **Release Please**
2. Click **Run workflow**
3. Optionally check **Force create a release** to skip commit analysis

## Configuration

All release settings are centralized in `.github/config/release-config.yaml`:

```yaml
# Commit types and changelog sections
commit_types:
  - type: "feat"
    section: "Features"
    hidden: false

# Package managers
package_managers:
  homebrew:
    enabled: false  # Enable when ready
    tap: "isaacgarza/homebrew-dev-stack"
  scoop:
    enabled: false  # Enable when ready
    bucket: "isaacgarza/scoop-dev-stack"
```

When you modify this file, regenerate the config files:

```bash
task generate-release-configs
```

## Package Managers

### Homebrew (macOS/Linux)

To enable Homebrew releases:

1. Create a tap repository: `isaacgarza/homebrew-dev-stack`
2. Enable in config: `package_managers.homebrew.enabled: true`
3. Add `HOMEBREW_TOKEN` secret to GitHub
4. Regenerate configs: `task generate-release-configs`

### Scoop (Windows)

To enable Scoop releases:

1. Create a bucket repository: `isaacgarza/scoop-dev-stack`
2. Enable in config: `package_managers.scoop.enabled: true`
3. Add `SCOOP_TOKEN` secret to GitHub
4. Regenerate configs: `task generate-release-configs`

## Troubleshooting

### Release PR not created

**Cause**: No conventional commits since last release
**Solution**: Ensure commits follow conventional format

### Version bump incorrect

**Cause**: Incorrect commit type or missing breaking change notation
**Solution**: Use correct commit types or add `BREAKING CHANGE:` footer

### Build failures

**Cause**: Tests failing or build errors
**Solution**: Fix issues and push again; Release-Please will update the PR

### Config files out of sync

**Cause**: Modified `.github/config/release-config.yaml` without regenerating
**Solution**: Run `task generate-release-configs`

## Best Practices

### Commit Messages

- **Be descriptive**: "fix: resolve timeout issue" not "fix: bug"
- **Use present tense**: "add" not "added" or "adds"
- **Keep subject under 50 characters**
- **Use body for context**: Explain what and why, not how

### Release Strategy

- **Small, frequent releases**: Better than large, infrequent ones
- **Test before merging**: Ensure CI passes
- **Review release PRs**: Check changelog and version bump
- **Document breaking changes**: Provide migration guidance

### Development Workflow

```bash
# Start feature branch
git checkout -b feat/awesome-feature

# Make commits with conventional format
git commit -m "feat: add awesome feature"
git commit -m "docs: update feature documentation"
git commit -m "test: add tests for awesome feature"

# Push and create PR to main
git push origin feat/awesome-feature

# After PR review and merge to main:
# - Release-Please creates release PR automatically
# - Review and merge release PR to trigger release
```

## Rollback

If you need to rollback a release:

1. **Revert the release commit** on main branch
2. **Delete the tag**: `git tag -d v1.2.3 && git push origin :refs/tags/v1.2.3`
3. **Delete the GitHub release** (optional)
4. **Create hotfix release** with the fix

## Getting Help

- Check the [Release-Please documentation](https://github.com/googleapis/release-please)
- Review [Conventional Commits specification](https://www.conventionalcommits.org/)
- View recent releases for examples
- Ask in team chat for guidance
