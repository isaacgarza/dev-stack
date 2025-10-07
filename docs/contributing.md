---
title: "Contributing"
description: "How to contribute to dev-stack development and documentation"
lead: "Join the dev-stack community and help make it better for everyone"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 60
toc: true

---

# Contributing Guide

## 📝 Quick Contributor Checklist

- [ ] Set up Python environment with pyenv (see below)
- [ ] Install dependencies: `pip install -r requirements.txt`
- [ ] Edit YAML manifests (`scripts/commands.yaml`, `services/services.yaml`) for changes
- [ ] Run `dev-stack docs` to update docs
- [ ] Commit both the manifest and generated docs
- [ ] Follow the contributing guide and PR template

This guide explains how to contribute to the Local Development Framework, including adding new services, improving existing functionality, and maintaining the codebase.

## 📋 Overview

We welcome contributions from the development community! Whether you're fixing bugs, adding new services, improving documentation, or enhancing existing features, your contributions help make the framework better for everyone.

## 🤝 Community & Support

### GitHub Repository

- **Main Repository**: [isaacgarza/dev-stack](https://github.com/isaacgarza/dev-stack)
- **Issues**: [Report bugs and request features](https://github.com/isaacgarza/dev-stack/issues)
- **Discussions**: [Ask questions and share ideas](https://github.com/isaacgarza/dev-stack/discussions)
- **Releases**: [Latest versions and changelog](https://github.com/isaacgarza/dev-stack/releases)

### Getting Help

**Before opening an issue:**
1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review [existing issues](https://github.com/isaacgarza/dev-stack/issues)
3. Search [discussions](https://github.com/isaacgarza/dev-stack/discussions)

**For support requests:**
- Use [GitHub Discussions](https://github.com/isaacgarza/dev-stack/discussions) for questions
- Check the [CLI Reference](reference.md) for command help
- Run `dev-stack doctor` for system diagnostics

**For bug reports:**
- Use [GitHub Issues](https://github.com/isaacgarza/dev-stack/issues)
- Include system information (`dev-stack --version`, OS, Docker version)
- Provide steps to reproduce the issue
- Include relevant logs and error messages

### Contributing Options

**Ways to contribute:**
- 🐛 **Bug fixes**: Fix issues and improve reliability
- ✨ **New features**: Add new services, commands, or functionality  
- 📚 **Documentation**: Improve guides, examples, and reference materials
- 🧪 **Testing**: Add tests, report bugs, validate on different platforms
- 💡 **Ideas**: Share suggestions in discussions

**Quick contributions:**
- Fix typos in documentation
- Add examples to existing guides
- Report unclear documentation
- Test new releases and report issues

---

## 📚 Automated Documentation & YAML Manifests

**dev-stack** uses YAML manifest files as the single source of truth for commands and services:

---

### 📚 Documentation Generation (Go-based)

The project is built entirely in Go and requires no additional language dependencies for development.

To auto-generate documentation from YAML manifests, dev-stack uses a native Go implementation with built-in YAML processing.

**Run the Doc Generation Command:**
```bash
dev-stack docs
```

This will update `docs/reference.md` and `docs/services.md` based on the latest YAML manifests.

**Contributor Workflow Checklist:**
1. Ensure you have Go 1.21+ installed and the project built (`task build`).
2. Edit `scripts/commands.yaml` and/or `services/services.yaml` to add or update commands/services.
3. Run `dev-stack docs` to regenerate documentation from YAML manifests.
4. Commit both the manifest and the updated docs.
5. Never manually edit auto-generated docs (`docs/reference.md`, `docs/services.md`).
6. Optionally, set up CI or pre-commit hooks to automate doc generation.

**Additional Documentation Options:**
```bash
# Generate only command reference
dev-stack docs --commands-only

# Generate only services guide  
dev-stack docs --services-only

# Preview changes without writing files
dev-stack docs --dry-run

# Show detailed progress
dev-stack docs --verbose
```

- `scripts/commands.yaml`: Lists all available CLI commands and flags for the dev-stack CLI.
- `services/services.yaml`: Lists all supported services and their configuration options.

Documentation for commands (`docs/reference.md`) and services (`docs/services.md`) is auto-generated from these manifests using the native Go `dev-stack docs` command.

See `dev-stack docs --help` for usage instructions.

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (see `.go-version`)
- Make (for running project automation tasks)
- Basic understanding of YAML, Bash scripting, and Docker
- Familiarity with the framework's architecture and usage

### Development Setup

Follow these steps to set up your environment for contributing:

1. Clone the framework repository:
   ```bash
   git clone <framework-repo> dev-stack-framework-dev
   cd dev-stack-framework-dev
   ```

2. Set up Python environment and install dependencies:
   ```bash
   task deps
   ```
   This will download and manage Go dependencies.

3. Create a test project to work with:
   ```bash
   mkdir test-project
   cd test-project
   ```

4. Link to your development framework:
   ```bash
   ln -s ../dev-stack-framework-dev dev-stack-framework
   ```

5. Test your changes:
   ```bash
   dev-stack init
   dev-stack up
   ```

```bash
# Clone the framework repository
git clone <framework-repo> dev-stack-framework-dev
cd dev-stack-framework-dev

# Create a test project to work with
mkdir test-project
cd test-project

# Link to your development framework
ln -s ../dev-stack-framework-dev dev-stack-framework

# Test your changes
dev-stack init
dev-stack up
```

## 🏗️ Architecture Overview

### Directory Structure

```
dev-stack/
├── scripts/                       # Main framework scripts and manifests
│   ├── setup.sh                  # Setup and configuration script
│   ├── manage.sh                 # Service management script
│   ├── lib/                      # Shared library functions
│   ├── commands.yaml             # YAML manifest for CLI commands (single source of truth)
│   └── generate_docs.py          # Python script to auto-generate docs from manifests
├── services/                     # Service definitions and manifests
│   ├── redis/                    # Redis service
│   ├── postgres/                 # PostgreSQL service
│   ├── mysql/                    # MySQL service
│   ├── jaeger/                   # Jaeger service
│   ├── prometheus/               # Prometheus service
│   ├── localstack/               # LocalStack service
│   ├── kafka/                    # Kafka service
│   └── services.yaml             # YAML manifest for all services/options (single source of truth)
├── config/                       # Framework configuration
│   └── framework.yaml            # Framework metadata and defaults
├── templates/                    # Configuration templates
│   └── spring-boot/              # Spring Boot integration templates
├── docs/                         # Documentation (auto-generated and manual)
│   ├── reference.md              # Auto-generated command reference
│   ├── services.md               # Auto-generated services guide
└── dev-stack-config.sample.yaml # Sample configuration
```

### Service Structure

Each service follows a standard structure:

```
services/service-name/
├── service.yaml                  # Service metadata and configuration
├── docker-compose.yml           # Docker Compose service definition
├── scripts/                     # Service-specific scripts (optional)
│   ├── init.sh                  # Initialization script
│   └── health-check.sh          # Health check script
└── config/                      # Service configuration files
    └── service.conf             # Service-specific configuration
```

## 🛠️ Adding New Services

### Step 1: Create Service Directory

```bash
mkdir services/my-service
cd services/my-service
```

### Step 2: Create Service Metadata

Create `service.yaml`:

```yaml
# Service metadata and configuration schema
name: my-service
description: "Brief description of the service"
category: database|cache|observability|messaging|cloud-services
version: "1.0"
maintainer: "Your Name <your.email@example.com>"

# Docker image configuration
image:
  name: "my-service"
  tag: "latest"
  registry: "docker.io"  # Optional

# Default configuration
defaults:
  port: 8080
  memory_limit: "256m"
  cpu_limit: "0.5"
  restart_policy: "unless-stopped"

  # Environment variables
  environment:
    MY_SERVICE_HOST: "localhost"
    MY_SERVICE_PORT: "${MY_SERVICE_PORT:-8080}"
    MY_SERVICE_PASSWORD: "${MY_SERVICE_PASSWORD:-}"

# Configuration overrides schema
overrides:
  port:
    type: integer
    default: 8080
    description: "Service port"
  memory_limit:
    type: string
    default: "256m"
    description: "Container memory limit"
  password:
    type: string
    default: ""
    description: "Service password"
  custom_config:
    type: string
    description: "Custom configuration content"

# Health check configuration
health_check:
  enabled: true
  command: ["curl", "-f", "http://localhost:8080/health"]
  interval: "30s"
  timeout: "10s"
  retries: 3
  start_period: "40s"

# Dependencies (other services required)
dependencies:
  - redis  # Optional dependency

# Spring Boot integration
spring_boot:
  enabled: true
  config_template: |
    my:
      service:
        endpoint: http://localhost:${MY_SERVICE_PORT:-8080}
        password: ${MY_SERVICE_PASSWORD:-}
  dependencies:
    - "org.springframework.boot:spring-boot-starter-web"
    - "com.my-service:my-service-client:1.0.0"

# Resource requirements
resources:
  memory:
    min: "128m"
    recommended: "256m"
    max: "1g"
  cpu:
    min: "0.1"
    recommended: "0.5"
  disk:
    min: "100m"
    recommended: "1g"

# Tags for service discovery and grouping
tags:
  - "database"
  - "nosql"
  - "development"
```

### Step 3: Create Docker Compose Definition

Create `docker-compose.yml`:

```yaml
services:
  my-service:
    image: ${MY_SERVICE_IMAGE:-my-service:latest}
    container_name: ${PROJECT_NAME}_my-service
    ports:
      - "${MY_SERVICE_PORT:-8080}:8080"
    environment:
      - MY_SERVICE_HOST=${MY_SERVICE_HOST:-localhost}
      - MY_SERVICE_PORT=${MY_SERVICE_PORT:-8080}
      - MY_SERVICE_PASSWORD=${MY_SERVICE_PASSWORD:-}
    volumes:
      - my-service-data:/data
      - ${MY_SERVICE_CONFIG_FILE:-./config/my-service.conf}:/etc/my-service/my-service.conf:ro
    networks:
      - ${NETWORK_NAME:-dev-stack-framework}
    restart: ${MY_SERVICE_RESTART_POLICY:-unless-stopped}
    mem_limit: ${MY_SERVICE_MEMORY_LIMIT:-256m}
    cpus: ${MY_SERVICE_CPU_LIMIT:-0.5}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    labels:
      - "dev-stack-framework.service=my-service"
      - "dev-stack-framework.project=${PROJECT_NAME}"

volumes:
  my-service-data:
    driver: local
    labels:
      - "dev-stack-framework.service=my-service"
      - "dev-stack-framework.project=${PROJECT_NAME}"

networks:
  dev-stack-framework:
    external: true
```

### Step 4: Create Initialization Script (Optional)

Create `scripts/init.sh`:

```bash
#!/bin/bash
set -e

# Service initialization script
echo "Initializing my-service..."

# Wait for service to be ready
until curl -f http://localhost:${MY_SERVICE_PORT:-8080}/health > /dev/null 2>&1; do
    echo "Waiting for my-service to be ready..."
    sleep 2
done

# Perform initialization tasks
echo "Creating default configuration..."
curl -X POST http://localhost:${MY_SERVICE_PORT:-8080}/api/init \
    -H "Content-Type: application/json" \
    -d '{"setup": true}'

echo "my-service initialization complete"
```

### Step 5: Create Configuration Template (Optional)

Create `config/my-service.conf`:

```
# My Service Configuration
host = localhost
port = 8080
log_level = info

# Authentication
password = ${MY_SERVICE_PASSWORD:-}

# Performance tuning
max_connections = 100
timeout = 30s
```

### Step 6: Update Framework Configuration

Add your service to `config/framework.yaml`:

```yaml
services:
  my-service:
    enabled: false  # Disabled by default
    category: "database"
    description: "My custom service"
    ports:
      - 8080
    dependencies: []
    resource_tier: "light"  # light, medium, heavy
```

### Step 7: Test Your Service

```bash
# Test service in isolation
cd test-project
echo "services:
  enabled:
    - my-service" > dev-stack-config.yaml

dev-stack up
dev-stack status

# Test with other services
echo "services:
  enabled:
    - redis
    - postgres
    - my-service" > dev-stack-config.yaml

dev-stack up
```

## 🔧 Improving Existing Services

### Modifying Service Configuration

1. **Update service.yaml**: Add new configuration options
2. **Update docker-compose.yml**: Add new environment variables or volumes
3. **Test changes**: Ensure backward compatibility
4. **Update documentation**: Document new configuration options

Example - Adding new Redis configuration:

```yaml
# In services/redis/service.yaml
overrides:
  # ... existing overrides ...
  max_connections:
    type: integer
    default: 1000
    description: "Maximum number of client connections"
```

```yaml
# In services/redis/docker-compose.yml
services:
  redis:
    # ... existing configuration ...
    command: >
      redis-server
      --maxclients ${REDIS_MAX_CONNECTIONS:-1000}
      --requirepass ${REDIS_PASSWORD:-}
```

### Adding Service Features

1. **Initialization scripts**: Add setup automation
2. **Health checks**: Improve service monitoring
3. **Configuration templates**: Better integration support
4. **Backup/restore**: Add data management features

## 📝 Documentation Contributions

### Adding Documentation

1. **Service documentation**: Document new services in `docs/services.md`
2. **Configuration examples**: Add examples to `docs/configuration.md`
3. **Troubleshooting**: Add common issues to `docs/troubleshooting.md`
4. **Quick reference**: Update command references in `docs/reference.md`

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Provide troubleshooting tips
- Keep formatting consistent
- Update table of contents

Example service documentation:

```markdown
### My Service

**Description**: Brief description of what the service does and its use cases.

**Default Configuration**:
- **Port**: 8080
- **Memory Limit**: 256MB
- **Version**: Latest

**Configuration Options**:
```yaml
overrides:
  my-service:
    port: 8080
    password: "secure-password"
    memory_limit: "512m"
```

**Spring Boot Integration**:
```yaml
my:
  service:
    endpoint: http://localhost:8080
    password: secure-password
```

**Useful Commands**:
```bash
# Connect to service
dev-stack connect my-service

# View service logs
dev-stack logs my-service
```
```

## 🐛 Bug Fixes and Improvements

### Common Areas for Improvement

1. **Script enhancements**: Improve error handling, logging, validation
2. **Performance optimizations**: Reduce startup time, memory usage
3. **Configuration validation**: Better error messages, type checking
4. **Cross-platform compatibility**: Windows, macOS, Linux support

### Bug Fix Process

1. **Reproduce the issue**: Create minimal test case
2. **Identify root cause**: Use debugging techniques
3. **Implement fix**: Make minimal, focused changes
4. **Test thoroughly**: Verify fix works, no regressions
5. **Document change**: Update relevant documentation

## 🧪 Testing

### Manual Testing

```bash
# Test new service (configure in dev-stack-config.yaml first)
dev-stack up
dev-stack status
dev-stack logs my-service

# Test cleanup
dev-stack cleanup
```

### Integration Testing

```bash
# Test multi-repository scenarios
cd repo1 && dev-stack up
cd ../repo2 && dev-stack up

# Test configuration variations
echo "Different config scenarios" > test-configs.yaml
dev-stack --config=test-configs.yaml up

# Test error scenarios
echo "Invalid YAML" > broken-config.yaml
dev-stack --config=broken-config.yaml up
```

### Validation Checklist

- [ ] Service starts successfully
- [ ] Health checks pass
- [ ] Service stops cleanly
- [ ] Configuration validation works
- [ ] Generated files are correct
- [ ] Documentation is accurate
- [ ] No regressions in existing services
- [ ] Cross-platform compatibility
- [ ] Resource usage is reasonable
- [ ] Error messages are helpful

## 📦 Submitting Contributions

### Pull Request Process

1. **Fork the repository** (if external contributor)
2. **Create feature branch**: `git checkout -b feature/my-new-service`
3. **Make changes**: Follow coding standards and conventions
4. **Test thoroughly**: Ensure everything works
5. **Commit changes**: Use descriptive commit messages
6. **Submit pull request**: Include detailed description

### Commit Message Format

```
type(scope): brief description

Detailed description of what was changed and why.

- List specific changes
- Include any breaking changes
- Reference related issues

Fixes #123
```

Examples:
- `feat(services): add MongoDB service with authentication`
- `fix(postgres): resolve connection timeout issue`
- `docs(configuration): add LocalStack DynamoDB examples`
- `refactor(scripts): improve error handling in setup.sh`

### Pull Request Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Changes Made
- List specific changes
- Include any new configuration options
- Mention any breaking changes

## Testing
- [ ] Manual testing completed
- [ ] Integration testing completed
- [ ] Documentation updated
- [ ] Examples provided

## Checklist
- [ ] Code follows project conventions
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
- [ ] Tested with existing services
```

## 🎯 Coding Standards

### Shell Scripting Standards

```bash
#!/bin/bash
# Use strict error handling
set -euo pipefail

# Use consistent formatting
function my_function() {
    local param1="$1"
    local param2="${2:-default}"

    # Use descriptive variable names
    local service_name="$param1"
    local config_file="$param2"

    # Quote variables
    echo "Processing service: ${service_name}"

    # Check for required parameters
    if [[ -z "${service_name}" ]]; then
        echo "Error: Service name is required" >&2
        return 1
    fi
}

# Use consistent error handling
if ! command -v docker >/dev/null 2>&1; then
    echo "Error: Docker is not installed" >&2
    exit 1
fi
```

### YAML Standards

```yaml
# Use consistent indentation (2 spaces)
# Quote strings with special characters
# Use descriptive keys
# Group related configuration

services:
  my-service:
    port: 8080
    memory_limit: "256m"
    environment:
      SERVICE_HOST: "localhost"
      SERVICE_PASSWORD: "${PASSWORD:-}"

    # Use comments for complex configurations
    health_check:
      # Check every 30 seconds with 10 second timeout
      interval: "30s"
      timeout: "10s"
```

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Provide troubleshooting guidance
- Keep formatting consistent
- Use proper markdown syntax

---

## 📝 How to Update Commands and Services Documentation

You can use Taskfile to automate documentation generation:

1. Edit the YAML manifests (`scripts/commands.yaml`, `services/services.yaml`).
2. Run `task docs` to regenerate documentation.
3. Commit both the manifest and the updated docs.

See the Contributor Workflow Checklist above for the recommended steps.

**Note:**
Do not manually edit `docs/reference.md` or `docs/services.md`—these files are always generated from the manifests.

**Tip:**
You can automate this process with a pre-commit hook or CI workflow using `./dev-stack docs`.

## 🤖 GitHub Workflows & CI/CD

For information about GitHub Actions workflows, CI/CD pipeline, and automation:
- **[GitHub Workflows Documentation](../.github/github-workflows.md)** - Complete overview of CI/CD setup
- **[Workflows README](../.github/workflows/README.md)** - Detailed workflow documentation
- **[Contributing to CI](../.github/workflows/README.md#development-workflow)** - How to work with the automation

Key workflows:
- **CI Pipeline**:# 🏆 Recognition

Contributors are recognized in several ways:

- **README acknowledgments**: Major contributors listed in main README
- **Changelog entries**: Contributions noted in release notes
- **Code comments**: Maintainer attribution in contributed code
- **Community recognition**: Shoutouts in team communications

## 📞 Getting Help

### Development Questions

- **Architecture questions**: Review existing service implementations
- **Integration questions**: Check Spring Boot template examples
- **Testing questions**: Look at existing service test patterns

### Code Review

- **Request feedback**: Ask for review before submitting large changes
- **Pair programming**: Available for complex contributions
- **Mentoring**: Help available for new contributors

### Resources

- **Existing services**: Use as implementation examples
- **Framework scripts**: Study `setup.sh` and `manage.sh` for patterns
- **Documentation**: Comprehensive guides in `docs/` directory

## 🔄 Maintenance

### Regular Maintenance Tasks

- **Update service images**: Keep Docker images current
- **Review dependencies**: Update framework dependencies
- **Performance monitoring**: Check resource usage patterns
- **Security updates**: Apply security patches promptly

### Release Process

## 📚 See Also

- [README](../README.md)
- [Configuration Guide](configuration.md)
- [Services Guide](services.md)
- [Setup Guide](setup.md)
- [Usage Guide](usage.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Integration Guide](integration.md)

1. **Version planning**: Plan new features and improvements
2. **Development cycle**: Implement and test changes
3. **Documentation updates**: Ensure docs are current
4. **Testing phase**: Comprehensive testing across environments
5. **Release preparation**: Prepare changelog and release notes
6. **Deployment**: Update framework in projects

Thank you for contributing to the Local Development Framework! Your contributions help make development easier and more productive for the entire team.
