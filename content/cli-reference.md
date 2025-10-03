---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI"
weight: 30
---

# CLI Reference

This page provides a comprehensive reference for all dev-stack CLI commands.

> **Note**: This page contains auto-generated content from the CLI help system. If you notice any discrepancies, please [open an issue](https://github.com/isaacgarza/dev-stack/issues).

## Global Options

All commands support these global options:

- `--config, -c` - Path to configuration file
- `--verbose, -v` - Enable verbose logging
- `--quiet, -q` - Suppress output except errors
- `--help, -h` - Show help for command
- `--version` - Show version information

## Commands

### Project Management

#### `dev-stack init`

Initialize a new project from a template.

```bash
dev-stack init [template] [flags]
```

**Templates:**
- `go` - Go application with hot reload
- `node` - Node.js/TypeScript application
- `python` - Python application
- `fullstack` - Full-stack application
- `minimal` - Minimal Docker setup

**Flags:**
- `--name` - Project name
- `--port` - Application port (default: 8080)
- `--with-database` - Add database service
- `--with-cache` - Add cache service
- `--force` - Overwrite existing files

**Examples:**
```bash
# Initialize Go project
dev-stack init go --name my-api

# Initialize with database
dev-stack init go --name my-api --with-database postgres

# Initialize with custom port
dev-stack init node --name my-app --port 3000
```

#### `dev-stack up`

Start the development environment.

```bash
dev-stack up [services...] [flags]
```

**Flags:**
- `--detach, -d` - Run in background
- `--force-recreate` - Recreate containers
- `--build` - Build images before starting
- `--pull` - Pull latest images

**Examples:**
```bash
# Start all services
dev-stack up

# Start specific services
dev-stack up database cache

# Start in background
dev-stack up --detach
```

#### `dev-stack down`

Stop the development environment.

```bash
dev-stack down [services...] [flags]
```

**Flags:**
- `--volumes` - Remove volumes
- `--remove-orphans` - Remove orphaned containers
- `--timeout` - Shutdown timeout (default: 10s)

**Examples:**
```bash
# Stop all services
dev-stack down

# Stop and remove volumes
dev-stack down --volumes

# Stop specific services
dev-stack down database
```

#### `dev-stack restart`

Restart services in the development environment.

```bash
dev-stack restart [services...] [flags]
```

**Flags:**
- `--timeout` - Restart timeout (default: 10s)

### Service Management

#### `dev-stack service list`

List available services.

```bash
dev-stack service list [flags]
```

**Flags:**
- `--category` - Filter by category
- `--installed` - Show only installed services
- `--available` - Show only available services

#### `dev-stack service add`

Add a service to the project.

```bash
dev-stack service add <service> [flags]
```

**Flags:**
- `--port` - Custom port mapping
- `--version` - Service version
- `--config` - Custom configuration

**Examples:**
```bash
# Add PostgreSQL
dev-stack service add postgres

# Add Redis with custom port
dev-stack service add redis --port 6380

# Add specific version
dev-stack service add postgres --version 13
```

#### `dev-stack service remove`

Remove a service from the project.

```bash
dev-stack service remove <service> [flags]
```

**Flags:**
- `--volumes` - Remove associated volumes
- `--force` - Force removal without confirmation

#### `dev-stack service info`

Show detailed information about a service.

```bash
dev-stack service info <service>
```

### Monitoring and Health

#### `dev-stack status`

Show status of all services.

```bash
dev-stack status [flags]
```

**Flags:**
- `--watch, -w` - Watch status continuously
- `--format` - Output format (table, json, yaml)

#### `dev-stack logs`

Show logs for services.

```bash
dev-stack logs [service] [flags]
```

**Flags:**
- `--follow, -f` - Follow log output
- `--tail` - Number of lines to show
- `--since` - Show logs since timestamp
- `--timestamps` - Show timestamps

**Examples:**
```bash
# Show all logs
dev-stack logs

# Follow logs for specific service
dev-stack logs postgres --follow

# Show last 100 lines
dev-stack logs --tail 100
```

#### `dev-stack doctor`

Run system health checks.

```bash
dev-stack doctor [flags]
```

**Flags:**
- `--check` - Run specific check (docker, network, disk)
- `--fix` - Attempt to fix issues
- `--json` - Output in JSON format

#### `dev-stack health`

Check health of services.

```bash
dev-stack health [service] [flags]
```

**Flags:**
- `--wait` - Wait for services to be healthy
- `--timeout` - Health check timeout

### Configuration

#### `dev-stack config show`

Display current configuration.

```bash
dev-stack config show [flags]
```

**Flags:**
- `--format` - Output format (yaml, json, table)
- `--global` - Show global configuration

#### `dev-stack config set`

Set configuration values.

```bash
dev-stack config set <key> <value> [flags]
```

**Flags:**
- `--global` - Set global configuration

**Examples:**
```bash
# Set application port
dev-stack config set app.port 8080

# Set global default template
dev-stack config set --global default-template go
```

#### `dev-stack config get`

Get configuration values.

```bash
dev-stack config get <key> [flags]
```

**Flags:**
- `--global` - Get from global configuration

#### `dev-stack config reset`

Reset configuration to defaults.

```bash
dev-stack config reset [flags]
```

**Flags:**
- `--global` - Reset global configuration
- `--force` - Skip confirmation

### Environment Variables

#### `dev-stack env list`

List environment variables.

```bash
dev-stack env list [flags]
```

**Flags:**
- `--format` - Output format (table, json, yaml)

#### `dev-stack env set`

Set environment variable.

```bash
dev-stack env set <key> <value>
```

#### `dev-stack env get`

Get environment variable.

```bash
dev-stack env get <key>
```

#### `dev-stack env load`

Load environment variables from file.

```bash
dev-stack env load <file>
```

### Utilities

#### `dev-stack version`

Show version information.

```bash
dev-stack version [flags]
```

**Flags:**
- `--json` - Output in JSON format
- `--short` - Show only version number

#### `dev-stack completion`

Generate shell completion scripts.

```bash
dev-stack completion <shell>
```

**Supported shells:**
- `bash`
- `zsh`
- `fish`
- `powershell`

**Examples:**
```bash
# Generate bash completion
dev-stack completion bash > /etc/bash_completion.d/dev-stack

# Generate zsh completion
dev-stack completion zsh > ~/.zsh/completions/_dev-stack
```

#### `dev-stack cleanup`

Clean up unused resources.

```bash
dev-stack cleanup [flags]
```

**Flags:**
- `--containers` - Remove stopped containers
- `--images` - Remove unused images
- `--volumes` - Remove unused volumes
- `--networks` - Remove unused networks
- `--all` - Remove all unused resources

## Exit Codes

dev-stack uses the following exit codes:

- `0` - Success
- `1` - General error
- `2` - Invalid command or arguments
- `3` - Configuration error
- `4` - Docker error
- `5` - Network error
- `125` - Docker daemon error
- `126` - Container command not found
- `127` - Container command not executable

## Environment Variables

dev-stack recognizes these environment variables:

- `DEV_STACK_CONFIG` - Path to configuration file
- `DEV_STACK_DEBUG` - Enable debug logging
- `DEV_STACK_NO_COLOR` - Disable colored output
- `DEV_STACK_REGISTRY` - Default Docker registry
- `DOCKER_HOST` - Docker daemon socket
- `COMPOSE_PROJECT_NAME` - Docker Compose project name

## Configuration Files

### Global Configuration

**Location:** `~/.config/dev-stack/config.yaml`

```yaml
default_template: go
registry: docker.io
auto_update: true
parallel_start: true
```

### Project Configuration

**Location:** `dev-stack-config.yaml` (in project root)

```yaml
name: my-project
version: 1.0.0
description: My awesome project

app:
  port: 8080
  environment: development

services:
  postgres:
    port: 5432
    database: myapp
  
env:
  DATABASE_URL: postgres://postgres:password@postgres:5432/myapp
```

## Troubleshooting

### Common Issues

**Command not found:**
```bash
# Check if dev-stack is in PATH
which dev-stack

# Reinstall or add to PATH
export PATH=$PATH:/usr/local/bin
```

**Permission denied:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

**Docker daemon not running:**
```bash
# Start Docker daemon
sudo systemctl start docker
```

### Getting Help

For additional help:

1. Use `dev-stack help [command]` for command-specific help
2. Check the [troubleshooting guide]({{< ref "/usage#troubleshooting" >}})
3. [Open an issue](https://github.com/isaacgarza/dev-stack/issues) on GitHub
4. Join the [community discussions](https://github.com/isaacgarza/dev-stack/discussions)

---

> **Auto-generated**: This page is automatically updated from the CLI help system. Last updated: Built with dev-stack CLI.