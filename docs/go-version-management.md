# Go Version Management

This document explains how Go version management is centralized across the dev-stack project to ensure consistency and maintainability.

## Overview

The dev-stack project uses a centralized approach to manage Go versions across all configuration files and workflows. This prevents version mismatches and makes it easy to upgrade Go versions across the entire project.

## Architecture

### Single Source of Truth

The **`.go-version`** file at the project root serves as the single source of truth for the Go version used throughout the project.

```bash
# .go-version contains just the version number
1.21
```

### Automated Synchronization

All configuration files that reference Go versions are automatically synchronized with the `.go-version` file using scripts and workflows.

## Files Managed

The following files are automatically synchronized with the Go version:

1. **`.go-version`** - Source of truth
2. **`go.mod`** - Go module definition
3. **`.golangci.yml`** - Linting configuration
4. **`Dockerfile`** - Container build configuration
5. **`Taskfile.yml`** - Build system variables
6. **GitHub Actions workflows** - CI/CD pipeline configurations

## Tools and Scripts

### Version Management Script

The `scripts/get-go-version.sh` script provides various ways to access the Go version:

```bash
# Get the raw version
./scripts/get-go-version.sh                 # Output: 1.21

# Get major.minor only
./scripts/get-go-version.sh --major-minor   # Output: 1.21

# Get as environment variable
./scripts/get-go-version.sh --env           # Output: GO_VERSION=1.21

# Get as GitHub Actions matrix
./scripts/get-go-version.sh --github-matrix # Output: ['1.19', '1.20', '1.21']
```

### Synchronization Script

The `scripts/sync-go-version.sh` script ensures all configuration files use the correct Go version:

```bash
# Check if all files are in sync
./scripts/sync-go-version.sh --check

# Fix any version mismatches
./scripts/sync-go-version.sh --fix

# Show help
./scripts/sync-go-version.sh --help
```

### Taskfile Integration

Convenient task targets are available for version management:

```bash
# Show current Go version and CI matrix
task show-go-version

# Check version consistency
task check-version

# Sync all configuration files
task sync-version
```

## GitHub Actions Integration

### Composite Action

The `.github/actions/setup-go-version` composite action automatically:

1. Reads the Go version from `.go-version`
2. Sets up Go with the correct version
3. Configures module caching
4. Outputs the version for use in other steps

Usage in workflows:
```yaml
- name: Setup Go with centralized version
  uses: ./.github/actions/setup-go-version
  id: setup-go

- name: Use Go version
  run: echo "Using Go ${{ steps.setup-go.outputs.go-version }}"
```

### Dynamic Matrix Builds

The test workflow automatically generates a matrix of Go versions for testing:

```yaml
jobs:
  get-versions:
    outputs:
      go-matrix: ${{ steps.versions.outputs.go-matrix }}
    steps:
      - run: |
          GO_MATRIX=$(./scripts/get-go-version.sh --github-matrix)
          echo "go-matrix=$GO_MATRIX" >> $GITHUB_OUTPUT

  test:
    needs: get-versions
    strategy:
      matrix:
        go-version: ${{ fromJson(needs.get-versions.outputs.go-matrix) }}
```

## Upgrading Go Version

To upgrade the Go version across the entire project:

1. **Update the source file:**
   ```bash
   echo "1.22" > .go-version
   ```

2. **Synchronize all configuration files:**
   ```bash
   task sync-version
   ```

3. **Verify consistency:**
   ```bash
   task check-version
   ```

4. **Update dependencies if needed:**
   ```bash
   go mod tidy
   ```

5. **Test the changes:**
   ```bash
   task test
   ```

6. **Commit the changes:**
   ```bash
   git add .
   git commit -m "feat: upgrade Go version to 1.22"
   ```

## Validation and CI

### Pre-commit Checks

The version consistency check can be added to pre-commit hooks:

```bash
#!/bin/bash
# .git/hooks/pre-commit
task check-version
```

### CI Validation

GitHub Actions workflows automatically validate version consistency:

```yaml
- name: Validate Go version consistency
  run: task check-version
```

## Configuration Details

### Taskfile Integration

The Taskfile dynamically reads the Go version:

```yaml
vars:
  GO_VERSION:
    sh: ./scripts/get-go-version.sh
```

This ensures build commands always use the correct version.

### golangci-lint Configuration

The `.golangci.yml` file is automatically updated to match:

```yaml
run:
  go: "1.21"  # Automatically synchronized
```

### Dockerfile Integration

Base images in Dockerfile are automatically updated:

```dockerfile
FROM golang:1.21-alpine AS builder
```

## Troubleshooting

### Version Mismatch Errors

If you see version mismatch errors:

```bash
# Check what's out of sync
task check-version

# Fix automatically
task sync-version
```

### Script Permissions

If scripts aren't executable:

```bash
chmod +x scripts/*.sh
```

### CI Matrix Issues

If GitHub Actions matrix builds fail, verify the matrix generation:

```bash
./scripts/get-go-version.sh --github-matrix
```

## Best Practices

1. **Always use `.go-version`** as the source of truth
2. **Run `task sync-version`** after changing Go versions
3. **Include version checks** in CI pipelines
4. **Test thoroughly** after Go version upgrades
5. **Update dependencies** with `go mod tidy` after upgrades

## Benefits

- **Consistency**: All tools use the same Go version
- **Maintainability**: Single place to update versions
- **Automation**: Scripts handle synchronization
- **Validation**: CI ensures consistency
- **Flexibility**: Easy to upgrade or downgrade versions

## Future Enhancements

- Pre-commit hooks for automatic validation
- Integration with more tools (IDE configurations, etc.)
- Automatic dependency updates when Go version changes
- Support for multiple Go version testing strategies