# Contributing Guide

This guide explains how to contribute to the Local Development Framework, including adding new services, improving existing functionality, and maintaining the codebase.

## 📋 Overview

We welcome contributions from the development community! Whether you're fixing bugs, adding new services, improving documentation, or enhancing existing features, your contributions help make the framework better for everyone.

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Basic understanding of YAML, Bash scripting, and Docker
- Familiarity with the framework's architecture and usage

### Development Setup

```bash
# Clone the framework repository
git clone <framework-repo> local-dev-framework-dev
cd local-dev-framework-dev

# Create a test project to work with
mkdir test-project
cd test-project

# Link to your development framework
ln -s ../local-dev-framework-dev local-dev-framework

# Test your changes
./local-dev-framework/scripts/setup.sh --init
./local-dev-framework/scripts/setup.sh
```

## 🏗️ Architecture Overview

### Directory Structure

```
local-dev-framework/
├── scripts/                       # Main framework scripts
│   ├── setup.sh                  # Setup and configuration script
│   ├── manage.sh                 # Service management script
│   └── lib/                      # Shared library functions
├── services/                     # Service definitions
│   ├── redis/                    # Redis service
│   ├── postgres/                 # PostgreSQL service
│   ├── mysql/                    # MySQL service
│   ├── jaeger/                   # Jaeger service
│   ├── prometheus/               # Prometheus service
│   ├── localstack/               # LocalStack service
│   └── kafka/                    # Kafka service
├── config/                       # Framework configuration
│   └── framework.yaml            # Framework metadata and defaults
├── templates/                    # Configuration templates
│   └── spring-boot/              # Spring Boot integration templates
├── docs/                         # Documentation
└── local-dev-config.sample.yaml # Sample configuration
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
      - ${NETWORK_NAME:-local-dev-framework}
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
      - "local-dev-framework.service=my-service"
      - "local-dev-framework.project=${PROJECT_NAME}"

volumes:
  my-service-data:
    driver: local
    labels:
      - "local-dev-framework.service=my-service"
      - "local-dev-framework.project=${PROJECT_NAME}"

networks:
  local-dev-framework:
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
    - my-service" > local-dev-config.yaml

./local-dev-framework/scripts/setup.sh
./local-dev-framework/scripts/manage.sh status
./local-dev-framework/scripts/manage.sh info my-service

# Test with other services
echo "services:
  enabled:
    - redis
    - my-service" > local-dev-config.yaml

./local-dev-framework/scripts/setup.sh
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
./scripts/manage.sh connect my-service

# View service logs
./scripts/manage.sh logs my-service
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
# Test new service
./scripts/setup.sh --services=my-service --dry-run
./scripts/setup.sh --services=my-service
./scripts/manage.sh status
./scripts/manage.sh info my-service
./scripts/manage.sh logs my-service

# Test with existing services
./scripts/setup.sh --services=redis,postgres,my-service
./scripts/manage.sh restart my-service

# Test cleanup
./scripts/manage.sh cleanup
```

### Integration Testing

```bash
# Test multi-repository scenarios
cd repo1 && ./scripts/setup.sh
cd ../repo2 && ./scripts/setup.sh --connect-existing

# Test configuration variations
echo "Different config scenarios" > test-configs.yaml
./scripts/setup.sh --config=test-configs.yaml

# Test error scenarios
echo "Invalid YAML" > broken-config.yaml
./scripts/setup.sh --config=broken-config.yaml
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

## 🏆 Recognition

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

1. **Version planning**: Plan new features and improvements
2. **Development cycle**: Implement and test changes
3. **Documentation updates**: Ensure docs are current
4. **Testing phase**: Comprehensive testing across environments
5. **Release preparation**: Prepare changelog and release notes
6. **Deployment**: Update framework in projects

Thank you for contributing to the Local Development Framework! Your contributions help make development easier and more productive for the entire team.
