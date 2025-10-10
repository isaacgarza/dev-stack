---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI"
lead: "Comprehensive reference for all dev-stack CLI commands and their usage"
date: 2025-10-10T04:10:05.115Z
lastmod: 2025-10-10T04:10:05.117Z
draft: false
weight: 50
toc: true
---

# dev-stack CLI Reference

Development stack management tool

Version: 0.1.0

## Commands

### config

```
Show the current configuration

Usage:
  dev-stack config [flags]

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### down

```
Stop the specified services or all services if none specified

Usage:
  dev-stack down [services...] [flags]

Flags:
      --remove        Remove containers (default true)
      --timeout int   Timeout in seconds (default 10)
      --volumes       Remove volumes

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### exec

```
Execute a command in a running service container

Usage:
  dev-stack exec <service> <command> [flags]

Flags:
  -i, --interactive   Interactive mode
  -t, --tty           Allocate a pseudo-TTY

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### init

```
Initialize a new dev-stack project with optional template

Usage:
  dev-stack init [template] [flags]

Flags:
      --name string       Project name
      --template string   Template to use

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### logs

```
Show logs for specified services or all services if none specified

Usage:
  dev-stack logs [services...] [flags]

Flags:
  -f, --follow        Follow log output
      --tail string   Number of lines to show from end of logs (default "100")

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### restart

```
Restart the specified services or all services if none specified

Usage:
  dev-stack restart [services...] [flags]

Flags:
      --timeout int   Timeout in seconds (default 10)

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### services

```
List all available services and their descriptions

Usage:
  dev-stack services [flags]

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### status

```
Show the status of specified services or all services if none specified

Usage:
  dev-stack status [services...] [flags]

Flags:
      --format string   Output format (table, json, yaml) (default "table")

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

### up

```
Start the specified services or all services if none specified

Usage:
  dev-stack up [services...] [flags]

Flags:
      --build            Build images before starting
      --force-recreate   Force recreate containers

Global Flags:
  -c, --config string   Config file (default: $HOME/.dev-stack.yaml)
  -h, --help            Show help information
  -v, --verbose         Enable verbose output
      --version         Show version information
```

