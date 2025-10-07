# Version Management

This document explains the comprehensive version management system in dev-stack, covering both Go runtime version management and dev-stack binary version management for multi-project environments.

## Overview

dev-stack implements a two-tier version management system:

1. **Go Runtime Version Management**: Centralized Go version management across project files and workflows
2. **dev-stack Binary Version Management**: Multi-version support for different projects with automatic version switching

This dual approach ensures both development consistency and deployment flexibility across different project requirements.

## dev-stack Binary Version Management

### Automatic Version Switching

dev-stack can manage multiple versions of itself and automatically switch between them based on project requirements. This allows different projects to use different versions of dev-stack without conflicts.

#### Version Detection

Projects can specify their dev-stack version requirements using `.dev-stack-version` files:

```bash
# Simple version requirement
echo "1.2.3" > .dev-stack-version

# Version constraint
echo ">=1.0.0" > .dev-stack-version

# YAML format with metadata
cat > .dev-stack-version.yaml << EOF
version: "^1.2.0"
metadata:
  created_by: "dev-stack"
  project: "my-project"
EOF
```

#### Supported Version Constraints

- **Exact**: `1.2.3` - Must match exactly
- **Greater/Less**: `>1.2.3`, `>=1.2.3`, `<2.0.0`, `<=1.9.9`
- **Tilde**: `~1.2.3` - Patch-level changes (>=1.2.3 <1.3.0)
- **Caret**: `^1.2.3` - Minor-level changes (>=1.2.3 <2.0.0)
- **Wildcard**: `*` - Any version

#### Version Management Commands

```bash
# List installed versions
dev-stack versions list

# Install a specific version
dev-stack versions install 1.2.3
dev-stack versions install latest

# Set active version globally
dev-stack versions use 1.2.3

# Set project-specific version requirement
dev-stack versions set ">=1.0.0" [path]

# Detect project version requirements
dev-stack versions detect [path]

# List available versions from GitHub
dev-stack versions available

# Clean up old versions
dev-stack versions cleanup --keep 3

# Uninstall a version
dev-stack versions uninstall 1.2.3
```

#### Automatic Installation and Switching

When dev-stack detects a project with version requirements:

1. **Detection**: Searches for `.dev-stack-version` files in project hierarchy
2. **Resolution**: Finds best matching installed version
3. **Auto-install**: Downloads missing versions automatically (optional)
4. **Delegation**: Transparently switches to the correct version
5. **Execution**: Runs the command with the appropriate version

```bash
# Project requires >=1.2.0, but only 1.1.0 is installed
cd my-project
dev-stack up  # Automatically suggests installing 1.2.0

# With auto-install enabled
dev-stack up  # Downloads 1.2.0 and delegates execution
```

#### Multi-Project Support

Each project can have isolated version requirements:

```bash
project-a/
├── .dev-stack-version    # Contains "1.1.0"
└── services/

project-b/
├── .dev-stack-version    # Contains ">=1.2.0"
└── docker-compose.yml

# Commands automatically use the right version per project
cd project-a && dev-stack up    # Uses version 1.1.0
cd project-b && dev-stack up    # Uses version 1.2.x
```

#### Version Storage and Management

Versions are stored in user directories:

```
~/.dev-stack/
├── versions/
│   ├── 1.1.0/
│   │   └── dev-stack
│   ├── 1.2.0/
│   │   └── dev-stack
│   └── 1.2.3/
│       └── dev-stack
└── config/
    ├── installed_versions.json
    └── project_configs.json
```

## Go Runtime Version Management

### Architecture

#### Single Source of Truth

The **`.go-version`** file at the project root serves as the single source of truth for the Go version used throughout the project.

```bash
# .go-version contains just the version number
1.21
```

#### Automated Synchronization

All configuration files that reference Go versions are automatically synchronized with the `.go-version` file using scripts and workflows.

#### Files Managed

The following files are automatically synchronized with the Go version:

1. **`.go-version`** - Source of truth
2. **`go.mod`** - Go module definition
3. **`.golangci.yml`** - Linting configuration
4. **`Dockerfile`** - Container build configuration
5. **`Taskfile.yml`** - Build system variables
6. **GitHub Actions workflows** - CI/CD pipeline configurations

#### Tools and Scripts

##### Version Management Script

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

##### Synchronization Script

The `scripts/sync-go-version.sh` script ensures all configuration files use the correct Go version:

```bash
# Check if all files are in sync
./scripts/sync-go-version.sh --check

# Fix any version mismatches
./scripts/sync-go-version.sh --fix

# Show help
./scripts/sync-go-version.sh --help
```

##### Taskfile Integration

Convenient task targets are available for version management:

```bash
# Show current Go version and CI matrix
task show-go-version

# Check version consistency
task check-version

# Sync all configuration files
task sync-version
```

#### GitHub Actions Integration

##### Composite Action

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

##### Dynamic Matrix Builds

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

#### Upgrading Go Version

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

#### Validation and CI

##### Pre-commit Checks

The version consistency check can be added to pre-commit hooks:

```bash
#!/bin/bash
# .git/hooks/pre-commit
task check-version
```

##### CI Validation

GitHub Actions workflows automatically validate version consistency:

```yaml
- name: Validate Go version consistency
  run: task check-version
```

#### Configuration Details

##### Taskfile Integration

The Taskfile dynamically reads the Go version:

```yaml
vars:
  GO_VERSION:
    sh: ./scripts/get-go-version.sh
```

This ensures build commands always use the correct version.

##### golangci-lint Configuration

The `.golangci.yml` file is automatically updated to match:

```yaml
run:
  go: "1.21"  # Automatically synchronized
```

##### Dockerfile Integration

Base images in Dockerfile are automatically updated:

```dockerfile
FROM golang:1.21-alpine AS builder
```

## Troubleshooting

### dev-stack Version Issues

#### No Compatible Version Found

```bash
# Error: No installed version satisfies requirement: >=1.2.0
# Solution: Install a compatible version
dev-stack versions install 1.2.0

# Or install latest
dev-stack versions install latest
```

#### Version File Not Detected

```bash
# Check if version file exists and is readable
dev-stack versions detect

# Create version requirement for current project
dev-stack versions set ">=1.0.0"
```

#### Binary Not Found

```bash
# List installed versions
dev-stack versions list

# Verify installation
dev-stack versions use 1.2.0
```

### Go Runtime Version Issues

#### Version Mismatch Errors

If you see version mismatch errors:

```bash
# Check what's out of sync
task check-version

# Fix automatically
task sync-version
```

#### Script Permissions

If scripts aren't executable:

```bash
chmod +x scripts/*.sh
```

#### CI Matrix Issues

If GitHub Actions matrix builds fail, verify the matrix generation:

```bash
./scripts/get-go-version.sh --github-matrix
```

## Best Practices

### dev-stack Version Management

1. **Use version files** for project-specific requirements
2. **Pin versions** for production environments
3. **Use constraints** for development flexibility
4. **Regular cleanup** to save disk space
5. **Document requirements** in project README

### Go Runtime Version Management

1. **Always use `.go-version`** as the source of truth
2. **Run `task sync-version`** after changing Go versions
3. **Include version checks** in CI pipelines
4. **Test thoroughly** after Go version upgrades
5. **Update dependencies** with `go mod tidy` after upgrades

## Benefits

### dev-stack Version Management

- **Project Isolation**: Each project can use its required version
- **Automatic Switching**: No manual version management needed
- **Backward Compatibility**: Old projects continue to work
- **Team Consistency**: Everyone uses the same version per project
- **Easy Upgrades**: Update version requirements as needed

### Go Runtime Version Management

- **Consistency**: All tools use the same Go version
- **Maintainability**: Single place to update versions
- **Automation**: Scripts handle synchronization
- **Validation**: CI ensures consistency
- **Flexibility**: Easy to upgrade or downgrade versions

## Future Enhancements

### dev-stack Version Management

- Integration with package managers (brew, apt, etc.)
- Automatic version updates and notifications
- Enhanced conflict resolution
- Integration with CI/CD pipelines
- Version analytics and usage tracking

### Go Runtime Version Management

- Pre-commit hooks for automatic validation
- Integration with more tools (IDE configurations, etc.)
- Automatic dependency updates when Go version changes
- Support for multiple Go version testing strategies
