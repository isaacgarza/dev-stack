---
title: "Version Management System"
description: "Multi-version support and automatic version switching for dev-stack"
lead: "Install, manage, and switch between different dev-stack versions"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 75
toc: true
---

# Version Management System

The dev-stack version management system allows you to install, manage, and automatically switch between different versions of dev-stack based on project requirements.

## Quick Start

```bash
# Set version requirement for current project
dev-stack versions set ">=1.0.0"

# Install a specific version
dev-stack versions install 1.2.3

# List installed versions
dev-stack versions list

# Automatically use the right version for your project
dev-stack up  # Uses version based on .dev-stack-version file
```

## Core Concepts

### Version Files

Projects specify their dev-stack version requirements using `.dev-stack-version` files:

**Simple text format:**
```
1.2.3
```

**Version constraint:**
```
>=1.0.0
```

**YAML format with metadata:**
```yaml
version: "^1.2.0"
metadata:
  created_by: "dev-stack"
  project: "my-awesome-app"
```

### Automatic Version Detection

dev-stack automatically detects version requirements by searching for version files in:

1. Current directory (`.`)
2. `.dev-stack/` subdirectory
3. `.config/` subdirectory
4. `config/` subdirectory

Supported file names:
- `.dev-stack-version`
- `.dev-stack-version.yaml`
- `.dev-stack-version.yml`
- `dev-stack-version`
- `dev-stack-version.yaml`
- `dev-stack-version.yml`

### Project Root Detection

When determining which version to use, dev-stack finds the project root by looking for:

- Version files (`.dev-stack-version`)
- Git repositories (`.git` directory)
- Common project files (`go.mod`, `package.json`, `Cargo.toml`, etc.)

## Version Constraints

dev-stack supports semantic versioning constraints:

| Constraint | Description | Example |
|------------|-------------|---------|
| `1.2.3` | Exact version | Must be exactly 1.2.3 |
| `>=1.2.3` | Greater than or equal | 1.2.3, 1.2.4, 1.3.0, 2.0.0 |
| `>1.2.3` | Greater than | 1.2.4, 1.3.0, 2.0.0 |
| `<=1.2.3` | Less than or equal | 1.0.0, 1.2.2, 1.2.3 |
| `<1.2.3` | Less than | 1.0.0, 1.2.2 |
| `~1.2.3` | Tilde (patch changes) | 1.2.3, 1.2.4, 1.2.10 (but not 1.3.0) |
| `^1.2.3` | Caret (minor changes) | 1.2.3, 1.3.0, 1.9.9 (but not 2.0.0) |
| `*` | Any version | Any available version |

## Commands

### `versions list`

List all installed versions of dev-stack.

```bash
dev-stack versions list

# Example output:
VERSION  ACTIVE  INSTALLED   SOURCE
1.1.0             2024-01-15  github
1.2.0    *        2024-02-01  github
1.2.3             2024-02-15  github
```

**Flags:**
- `--json` - Output in JSON format

### `versions install`

Install a specific version of dev-stack.

```bash
# Install specific version
dev-stack versions install 1.2.3

# Install latest version
dev-stack versions install latest
```

The command will:
1. Download the binary from GitHub releases
2. Verify checksums (if available)
3. Extract and install to version-specific directory
4. Register the version in the local registry

### `versions uninstall`

Remove a specific version of dev-stack.

```bash
dev-stack versions uninstall 1.2.3
```

**Note:** Cannot uninstall the currently active version.

### `versions use`

Set the global default version of dev-stack.

```bash
dev-stack versions use 1.2.3
```

This sets the version to use when no project-specific version is found.

### `versions available`

List all available versions from GitHub releases.

```bash
dev-stack versions available

# Limit results
dev-stack versions available --limit 10

# JSON output
dev-stack versions available --json
```

### `versions detect`

Detect version requirements for a project.

```bash
# Detect in current directory
dev-stack versions detect

# Detect in specific directory
dev-stack versions detect /path/to/project
```

**Example output:**
```
Project: /home/user/my-project
Required version: >=1.0.0
Resolved to installed version: 1.2.3
```

**Flags:**
- `--json` - Output in JSON format

### `versions set`

Set version requirement for a project.

```bash
# Set for current directory
dev-stack versions set ">=1.0.0"

# Set for specific directory
dev-stack versions set "^1.2.0" /path/to/project

# Use YAML format
dev-stack versions set "1.2.3" --format yaml
```

**Flags:**
- `--format` - File format (`text` or `yaml`)

### `versions cleanup`

Clean up old versions to save disk space.

```bash
# Keep 3 most recent versions
dev-stack versions cleanup --keep 3

# Dry run to see what would be removed
dev-stack versions cleanup --dry-run
```

## Multi-Project Workflow

### Example Setup

```bash
# Project A needs dev-stack 1.1.x
cd project-a
dev-stack versions set "~1.1.0"

# Project B needs dev-stack 1.2.x or higher
cd project-b
dev-stack versions set ">=1.2.0"

# Install required versions
dev-stack versions install 1.1.5
dev-stack versions install 1.2.3
```

### Automatic Switching

Once versions are installed and requirements are set:

```bash
# Automatically uses 1.1.5
cd project-a
dev-stack up

# Automatically uses 1.2.3
cd project-b
dev-stack up
```

### Version Resolution

When you run a dev-stack command, the system:

1. **Finds project root** by searching upward for version files or project markers
2. **Reads version constraint** from `.dev-stack-version` file
3. **Resolves to best match** among installed versions
4. **Delegates execution** to the correct version binary
5. **Falls back** to active version if no constraint found

## Storage Layout

Versions are stored in your home directory:

```
~/.dev-stack/
├── versions/                 # Installed version binaries
│   ├── 1.1.0/
│   │   └── dev-stack
│   ├── 1.2.0/
│   │   └── dev-stack
│   └── 1.2.3/
│       └── dev-stack
└── ...

~/.config/dev-stack/
├── installed_versions.json   # Registry of installed versions
└── project_configs.json     # Per-project configurations
```

## Advanced Usage

### CI/CD Integration

Pin specific versions in CI environments:

```bash
# In your CI script
echo "1.2.3" > .dev-stack-version
dev-stack up  # Will use exactly 1.2.3
```

### Team Consistency

Commit `.dev-stack-version` files to ensure team consistency:

```bash
# Set project requirement
dev-stack versions set "^1.2.0"

# Commit to repo
git add .dev-stack-version
git commit -m "Pin dev-stack version to ^1.2.0"
```

### Version Verification

Verify installed versions and project requirements:

```bash
# Check what version would be used
dev-stack versions detect

# List all installed versions
dev-stack versions list

# Check available updates
dev-stack versions available | head -5
```

### Cleanup Strategy

Regular maintenance to manage disk usage:

```bash
# Keep only 3 most recent versions
dev-stack versions cleanup --keep 3

# Preview cleanup without making changes
dev-stack versions cleanup --dry-run
```

## Troubleshooting

### Common Issues

**No compatible version found:**
```bash
Error: No installed version satisfies requirement: >=1.3.0
Run 'dev-stack versions install 1.3.0' to install a compatible version.
```

**Solution:** Install a compatible version:
```bash
dev-stack versions install 1.3.0
# or install latest
dev-stack versions install latest
```

**Version file not detected:**
```bash
dev-stack versions detect
# Output: No specific version requirement found
```

**Solution:** Create a version requirement:
```bash
dev-stack versions set ">=1.0.0"
```

**Binary delegation fails:**
- Check file permissions in `~/.dev-stack/versions/`
- Verify version registry: `dev-stack versions list`
- Reinstall problematic version: `dev-stack versions install X.Y.Z`

### Debug Information

Get detailed information about version resolution:

```bash
# Enable verbose output
dev-stack --verbose versions detect

# Check version manager state
ls -la ~/.dev-stack/versions/
cat ~/.config/dev-stack/installed_versions.json
```

## Best Practices

1. **Pin versions in production**: Use exact versions for deployments
2. **Use constraints for development**: Allow flexibility with `^` or `~`
3. **Commit version files**: Ensure team consistency
4. **Regular cleanup**: Manage disk space with periodic cleanup
5. **Document requirements**: Include version info in project README
6. **Test version changes**: Verify compatibility when updating constraints

## Security

- **Checksum verification**: Downloads are verified against GitHub release checksums when available
- **HTTPS downloads**: All downloads use secure HTTPS connections
- **No auto-execution**: Downloaded binaries are not executed during installation
- **User-scoped storage**: Versions are stored in user directories, not system-wide

## Migration from Manual Management

If you're currently managing dev-stack versions manually:

1. **Install version manager**: Upgrade to a version with version management
2. **Set project requirements**: Add `.dev-stack-version` files to your projects
3. **Install needed versions**: Use `dev-stack versions install` for required versions
4. **Remove manual installations**: Clean up old manual installations
5. **Update documentation**: Update team documentation with new workflow
