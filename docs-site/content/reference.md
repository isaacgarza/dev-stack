---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI"
lead: "Comprehensive reference for all dev-stack CLI commands and their usage"
date: "2025-10-01"
lastmod: "2025-10-13"
draft: false
weight: 50
toc: true
---

# dev-stack CLI Reference

Development stack management tool

Version: 0.1.0

## Commands

### conflicts

```
Check if the specified services have any conflicts that would prevent
them from running together. Identifies port conflicts, resource conflicts,
and incompatible service combinations.

Usage:
  dev-stack conflicts <service1> <service2> [service...] [flags]

Examples:
  dev-stack conflicts postgres mysql
    Check if postgres and mysql conflict

  dev-stack conflicts postgres redis kafka-broker
    Check conflicts between multiple services



Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### deps

```
Display the complete dependency tree for a service, showing all required
dependencies and the resolved start order. Helps understand service
relationships and startup sequences.

Usage:
  dev-stack deps <service> [flags]

Examples:
  dev-stack deps kafka-ui
    Show dependencies for kafka-ui service

  dev-stack deps postgres
    Show dependencies for postgres service



Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### doctor

```
Run comprehensive health checks on your development stack. Identifies
common issues, provides troubleshooting suggestions, and validates
service configurations.

Usage:
  dev-stack doctor [service...] [flags]

Examples:
  dev-stack doctor
    Run health checks on all services

  dev-stack doctor postgres
    Diagnose a specific service

  dev-stack doctor --fix
    Attempt to fix detected issues



Flags:
      --fix             Attempt to automatically fix issues
  -f, --format string   Output format (table|json) (default "table")
  -v, --verbose         Show detailed diagnostic information

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
      --version           Show version information
```

### down

```
Stop one or more services in the development stack. By default, containers
are removed but volumes are preserved. Use --volumes to also remove data.

Usage:
  dev-stack down [service...] [flags]

Aliases:
  down, stop

Examples:
  dev-stack down
    Stop all running services

  dev-stack down postgres redis
    Stop specific services

  dev-stack down --volumes
    Stop services and remove volumes

  dev-stack down --timeout 5
    Stop services with custom timeout



Flags:
      --remove-images string   Remove images (all|local)
      --remove-orphans         Remove containers for services not in compose file
  -t, --timeout int            Shutdown timeout in seconds (default 10)
  -v, --volumes                Remove named volumes and anonymous volumes

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### init

```
Initialize a new dev-stack project in the current directory with an
interactive setup process. Guides you through selecting services,
configuring validation and advanced settings, and creates all
necessary configuration files.

Usage:
  dev-stack init [flags]

Examples:
  dev-stack init
    Interactive project initialization (recommended)

  dev-stack init --name myproject --minimal
    Non-interactive minimal setup

  dev-stack init --force
    Overwrite existing configuration



Flags:
  -f, --force   Overwrite existing files

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### restart

```
Restart one or more services. This is equivalent to running down followed
by up, but more efficient for quick restarts.

Usage:
  dev-stack restart [service...] [flags]

Examples:
  dev-stack restart
    Restart all services

  dev-stack restart postgres
    Restart a specific service

  dev-stack restart --timeout 5
    Restart with custom timeout



Flags:
      --no-deps       Don't restart linked services
  -t, --timeout int   Restart timeout in seconds (default 10)

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### services

```
List all available services organized by category (database, cache,
messaging, observability, cloud). Shows service descriptions and
dependencies for easy discovery and selection.

Usage:
  dev-stack services [flags]

Examples:
  dev-stack services
    List all services grouped by category

  dev-stack services --category database
    List services in database category

  dev-stack services --category cache
    List cache services



Flags:
  -c, --category string   Show services in specific category

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### status

```
Display comprehensive status information for services including running
state, health checks, resource usage, and port mappings. Supports multiple
output formats and real-time monitoring.

Usage:
  dev-stack status [service...] [flags]

Aliases:
  status, ps, ls

Examples:
  dev-stack status
    Show status of all services

  dev-stack status postgres redis
    Show status of specific services

  dev-stack status --format json
    Output status in JSON format

  dev-stack status --watch
    Watch for status changes in real-time

  dev-stack status --filter running
    Show only running services



Flags:
      --filter string   Filter services by status
  -f, --format string   Output format (table|json|yaml) (default "table")
      --no-trunc        Don't truncate output
  -q, --quiet           Only show service names and basic status
  -w, --watch           Watch for status changes

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### up

```
Start one or more services in the development stack. Services are started
with their configured dependencies and health checks. Use profiles to start
predefined service combinations.

Usage:
  dev-stack up [service...] [flags]

Aliases:
  up, start, run

Examples:
  dev-stack up
    Start all configured services

  dev-stack up postgres redis
    Start specific services

  dev-stack up --profile web
    Start services using the 'web' profile

  dev-stack up --detach --build
    Build images and start services in background



Flags:
  -b, --build             Build images before starting services
      --check-conflicts   Check for service conflicts before starting
  -d, --detach            Run services in background (detached mode)
      --force-recreate    Recreate containers even if config hasn't changed
      --no-deps           Don't start linked services
  -p, --profile string    Use a specific service profile
      --resolve-deps      Show dependency resolution tree before starting
  -t, --timeout string    Timeout for service startup (e.g., 30s, 2m) (default "30s")

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```

### validate

```
Validate dev-stack configurations, service definitions, and YAML
manifests. Checks for syntax errors, missing dependencies, and
configuration inconsistencies.

Usage:
  dev-stack validate [file...] [flags]

Examples:
  dev-stack validate
    Validate all configuration files

  dev-stack validate dev-stack-config.yaml
    Validate specific configuration file

  dev-stack validate --strict
    Use strict validation rules



Flags:
      --fix             Attempt to fix validation errors
  -f, --format string   Output format (table|json) (default "table")
  -s, --strict          Use strict validation rules

Global Flags:
  -c, --config string     Config file (default: $HOME/.dev-stack.yaml)
  -h, --help              Show help information
      --json              Output in JSON format (CI-friendly)
      --no-color          Disable colored output (CI-friendly)
      --non-interactive   Run in non-interactive mode (CI-friendly)
  -q, --quiet             Suppress non-essential output (CI-friendly)
  -v, --verbose           Enable verbose output
      --version           Show version information
```
