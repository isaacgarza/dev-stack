---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI"
weight: 30
---

# CLI Reference

This page provides comprehensive documentation for all dev-stack CLI commands and options.

## Overview

The `dev-stack` CLI is designed to streamline development environment management with intuitive commands and helpful features.

## Global Options

```bash
dev-stack [global options] command [command options] [arguments...]
```

### Global Flags

- `--help, -h` - Show help
- `--version, -v` - Print version information
- `--verbose` - Enable verbose logging
- `--config` - Specify configuration file path

## Commands

### Core Commands

#### `init`
Initialize a new development environment

```bash
dev-stack init [options] [template]
```

**Options:**
- `--name` - Project name
- `--path` - Target directory
- `--template` - Project template to use

**Examples:**
```bash
dev-stack init go --name my-api
dev-stack init --template node --name frontend-app
```

#### `up`
Start the development environment

```bash
dev-stack up [options]
```

**Options:**
- `--detach, -d` - Run in background
- `--build` - Force rebuild of containers
- `--services` - Start specific services only

**Examples:**
```bash
dev-stack up
dev-stack up --detach
dev-stack up --services postgres,redis
```

#### `down`
Stop the development environment

```bash
dev-stack down [options]
```

**Options:**
- `--volumes` - Remove volumes
- `--remove-orphans` - Remove orphaned containers

#### `status`
Show status of development environment

```bash
dev-stack status [options]
```

**Options:**
- `--watch` - Continuous monitoring
- `--format` - Output format (table, json, yaml)

### Service Management

#### `service add`
Add a service to the environment

```bash
dev-stack service add <service-name> [options]
```

**Available Services:**
- `postgres` - PostgreSQL database
- `mysql` - MySQL database
- `redis` - Redis cache
- `mongodb` - MongoDB database
- `rabbitmq` - RabbitMQ message broker
- `elasticsearch` - Elasticsearch search engine

#### `service remove`
Remove a service from the environment

```bash
dev-stack service remove <service-name>
```

#### `service list`
List available and active services

```bash
dev-stack service list [options]
```

**Options:**
- `--available` - Show all available services
- `--active` - Show only active services

### Utility Commands

#### `doctor`
Check system health and configuration

```bash
dev-stack doctor
```

Validates:
- Docker installation and permissions
- Required tools availability
- Configuration file syntax
- Network connectivity

#### `logs`
View service logs

```bash
dev-stack logs [service-name] [options]
```

**Options:**
- `--follow, -f` - Follow log output
- `--tail` - Number of lines to show
- `--since` - Show logs since timestamp

#### `exec`
Execute commands in service containers

```bash
dev-stack exec <service-name> <command>
```

**Examples:**
```bash
dev-stack exec postgres psql -U postgres
dev-stack exec app bash
```

#### `config`
Configuration management

```bash
dev-stack config <subcommand>
```

**Subcommands:**
- `show` - Display current configuration
- `validate` - Validate configuration file
- `export` - Export configuration to file
- `import` - Import configuration from file

### Advanced Commands

#### `backup`
Backup service data

```bash
dev-stack backup <service-name> [options]
```

**Options:**
- `--output` - Backup file path
- `--compress` - Compress backup

#### `restore`
Restore service data from backup

```bash
dev-stack restore <service-name> <backup-file>
```

#### `clean`
Clean up unused resources

```bash
dev-stack clean [options]
```

**Options:**
- `--volumes` - Remove unused volumes
- `--images` - Remove unused images
- `--all` - Remove all unused resources

## Configuration

### Configuration File

Dev-stack uses a configuration file (typically `dev-stack.yaml`) to define your development environment:

```yaml
project:
  name: "my-project"
  template: "go"

services:
  - name: postgres
    version: "15"
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: developer
      POSTGRES_PASSWORD: secret

  - name: redis
    version: "7"
```

### Environment Variables

- `DEV_STACK_CONFIG` - Path to configuration file
- `DEV_STACK_HOME` - Home directory for dev-stack files
- `DEV_STACK_LOG_LEVEL` - Logging level (debug, info, warn, error)

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Service error
- `4` - Network error

## Examples

### Complete Workflow

```bash
# Initialize a new Go project
dev-stack init go --name my-api

# Add required services
dev-stack service add postgres
dev-stack service add redis

# Start the environment
dev-stack up

# Check status
dev-stack status

# View logs
dev-stack logs postgres --follow

# Stop when done
dev-stack down
```

### Debugging

```bash
# Check system health
dev-stack doctor

# View detailed logs
dev-stack logs --follow

# Execute commands in containers
dev-stack exec app bash
dev-stack exec postgres psql -U postgres
```

## Getting Help

For more information:
- Use `dev-stack --help` for general help
- Use `dev-stack <command> --help` for command-specific help
- Visit the [GitHub repository](https://github.com/isaacgarza/dev-stack)
- Check the [troubleshooting guide]({{< ref "/usage#troubleshooting" >}})

## Version Information

This documentation is automatically generated and matches the installed version of dev-stack. To check your version:

```bash
dev-stack --version
```
