---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI with all available commands and options"
lead: "Comprehensive reference for all dev-stack CLI commands and their usage"
date: 2025-10-07T12:20:14-05:00
lastmod: 2025-10-07T12:20:14-05:00
draft: false
weight: 50
toc: true
---

<!-- AUTO-GENERATED-START -->
# dev-stack CLI Reference

Development stack management tool

**Version:** 0.1.0
**Generated:** 2024-01-01 00:00:00

## Quick Reference


### 💾 Data Management

Commands for backup, restore, and data operations


- **[backup](#backup)** - Backup service data and configurations

- **[restore](#restore)** - Restore service data from backups

- **[exec](#exec)** - Execute commands in running service containers

- **[connect](#connect)** - Quick connect to service databases and interfaces



### 🛠️ Development Tools

Commands for development workflow and documentation


- **[docs](#docs)** - Generate and manage documentation

- **[generate](#generate)** - Generate configuration files and templates

- **[validate](#validate)** - Validate configurations and manifests



### 🚀 Lifecycle Management

Commands for starting, stopping, and managing service lifecycles


- **[up](#up)** - Start development stack services

- **[down](#down)** - Stop development stack services

- **[restart](#restart)** - Restart development stack services

- **[scale](#scale)** - Scale services up or down



### 🧹 Maintenance & Cleanup

Commands for cleanup, initialization, and maintenance


- **[cleanup](#cleanup)** - Clean up unused resources and data

- **[init](#init)** - Initialize a new dev-stack project

- **[version](#version)** - Show version information



### 📊 Monitoring & Observability

Commands for monitoring services and viewing logs


- **[status](#status)** - Show status of development stack services

- **[logs](#logs)** - View logs from services

- **[monitor](#monitor)** - Real-time monitoring dashboard for services

- **[doctor](#doctor)** - Diagnose and troubleshoot stack health




## Global Flags


- **--config**, **-c** (string) - Config file (default: $HOME/.dev-stack.yaml)

- **--help**, **-h** (bool) - Show help information

- **--verbose**, **-v** (bool) - Enable verbose output

- **--version** (bool) - Show version information


## Commands



### 💾 Data Management

Commands for backup, restore, and data operations


#### backup

**Usage:** `backup [service...]`

Backup service data and configurations


Create backups of service data, configurations, and state. Supports
multiple backup formats and compression. Backups include metadata
for easy restoration.






**Flags:**

- **--compress**, **-c** (bool) - Compress backup files

- **--exclude-logs** (bool) - Exclude log files from backup

- **--format**, **-f** (string) - Backup format (native|sql|json) (default: native)

- **--include-config** (bool) - Include service configurations (default: true)

- **--output**, **-o** (string) - Output directory for backups (default: ./backups)




**Examples:**

```bash
dev-stack backup
```
Backup all services


```bash
dev-stack backup postgres redis
```
Backup specific services


```bash
dev-stack backup --output ./backups --compress
```
Backup with compression to custom directory







**See also:** restore, cleanup


---


#### restore

**Usage:** `restore <service> <backup-path>`

Restore service data from backups


Restore service data and configurations from previously created backups.
Supports multiple restore strategies and validation of backup integrity.






**Flags:**

- **--clean** (bool) - Clean existing data before restore

- **--create-db** (bool) - Create database if it doesn't exist (default: true)

- **--no-owner** (bool) - Skip ownership commands

- **--single-transaction** (bool) - Perform restore in single transaction

- **--validate** (bool) - Validate backup before restore (default: true)




**Examples:**

```bash
dev-stack restore postgres ./backups/postgres-20240101.sql
```
Restore PostgreSQL from SQL backup


```bash
dev-stack restore redis ./backups/redis-20240101.rdb
```
Restore Redis from RDB backup


```bash
dev-stack restore --clean postgres backup.sql
```
Clean database before restore







**See also:** backup, cleanup


---


#### exec

**Usage:** `exec <service> <command> [args...]`

Execute commands in running service containers


Execute commands inside running service containers. Useful for database
operations, debugging, and maintenance tasks. Supports interactive and
non-interactive modes.






**Flags:**

- **--detach**, **-d** (bool) - Run command in background

- **--env**, **-e** (stringArray) - Set environment variables

- **--interactive**, **-i** (bool) - Keep STDIN open (interactive mode) (default: true)

- **--tty**, **-t** (bool) - Allocate a pseudo-TTY (default: true)

- **--user**, **-u** (string) - Username to execute command as

- **--workdir**, **-w** (string) - Working directory for command




**Examples:**

```bash
dev-stack exec postgres psql -U postgres
```
Connect to PostgreSQL with psql


```bash
dev-stack exec redis redis-cli
```
Connect to Redis CLI


```bash
dev-stack exec postgres bash
```
Open bash shell in postgres container





**Tips:**

- Use for database maintenance and debugging

- Combine with --user to run as specific user




**See also:** connect, logs


---


#### connect

**Usage:** `connect <service>`

Quick connect to service databases and interfaces


Quickly connect to service databases and management interfaces using
appropriate client tools. Automatically configures connection parameters
based on service configuration.






**Flags:**

- **--database**, **-d** (string) - Database name to connect to

- **--host**, **-h** (string) - Host to connect to (default: localhost)

- **--port**, **-p** (int) - Port to connect to

- **--read-only** (bool) - Connect in read-only mode

- **--user**, **-u** (string) - Username for connection




**Examples:**

```bash
dev-stack connect postgres
```
Connect to PostgreSQL database


```bash
dev-stack connect redis
```
Connect to Redis CLI


```bash
dev-stack connect mysql
```
Connect to MySQL database





**Tips:**

- Automatically uses correct client tools for each service

- Use --read-only for safe data exploration




**See also:** exec, status


---




### 🛠️ Development Tools

Commands for development workflow and documentation


#### docs

**Usage:** `docs [subcommand]`

Generate and manage documentation


Generate documentation from YAML manifests and manage documentation
files. Supports multiple output formats and automatic synchronization
with documentation websites.






**Flags:**

- **--commands-only** (bool) - Generate only command reference

- **--dry-run** (bool) - Preview changes without writing files

- **--format**, **-f** (string) - Output format (markdown|html|json) (default: markdown)

- **--hugo-sync** (bool) - Sync to Hugo site (default: true)

- **--no-hugo-sync** (bool) - Skip Hugo synchronization

- **--services-only** (bool) - Generate only services documentation




**Examples:**

```bash
dev-stack docs
```
Generate all documentation


```bash
dev-stack docs --commands-only
```
Generate only command reference


```bash
dev-stack docs --dry-run
```
Preview documentation changes







**See also:** init, validate


---


#### generate

**Usage:** `generate <type> [options]`

Generate configuration files and templates


Generate various configuration files, service definitions, and project
templates. Supports custom templates and configuration inheritance.






**Flags:**

- **--output**, **-o** (string) - Output file or directory

- **--overwrite** (bool) - Overwrite existing files

- **--services**, **-s** (string) - Comma-separated list of services

- **--template**, **-t** (string) - Template to use (default: default)




**Examples:**

```bash
dev-stack generate config
```
Generate base configuration file


```bash
dev-stack generate service postgres
```
Generate PostgreSQL service configuration


```bash
dev-stack generate compose --services postgres,redis
```
Generate docker-compose.yml for specific services







**See also:** init, validate


---


#### validate

**Usage:** `validate [file...]`

Validate configurations and manifests


Validate dev-stack configurations, service definitions, and YAML
manifests. Checks for syntax errors, missing dependencies, and
configuration inconsistencies.






**Flags:**

- **--fix** (bool) - Attempt to fix validation errors

- **--format**, **-f** (string) - Output format (table|json) (default: table)

- **--strict**, **-s** (bool) - Use strict validation rules




**Examples:**

```bash
dev-stack validate
```
Validate all configuration files


```bash
dev-stack validate dev-stack-config.yaml
```
Validate specific configuration file


```bash
dev-stack validate --strict
```
Use strict validation rules







**See also:** doctor, docs


---




### 🚀 Lifecycle Management

Commands for starting, stopping, and managing service lifecycles


#### up

**Usage:** `up [service...]`

Start development stack services


Start one or more services in the development stack. Services are started
with their configured dependencies and health checks. Use profiles to start
predefined service combinations.




**Aliases:** start, run



**Flags:**

- **--build**, **-b** (bool) - Build images before starting services

- **--detach**, **-d** (bool) - Run services in background (detached mode)

- **--force-recreate** (bool) - Recreate containers even if config hasn't changed

- **--no-deps** (bool) - Don't start linked services

- **--profile**, **-p** (string) - Use a specific service profile

- **--timeout**, **-t** (duration) - Timeout for service startup (default: 30s)




**Examples:**

```bash
dev-stack up
```
Start all configured services


```bash
dev-stack up postgres redis
```
Start specific services


```bash
dev-stack up --profile web
```
Start services using the 'web' profile


```bash
dev-stack up --detach --build
```
Build images and start services in background





**Tips:**

- Use --profile to quickly start predefined service combinations

- Add --build if you've made changes to Dockerfiles

- Use --detach to free up your terminal while services run




**See also:** down, restart, status


---


#### down

**Usage:** `down [service...]`

Stop development stack services


Stop one or more services in the development stack. By default, containers
are removed but volumes are preserved. Use --volumes to also remove data.




**Aliases:** stop



**Flags:**

- **--remove-images** (string) - Remove images (all|local)

- **--remove-orphans** (bool) - Remove containers for services not in compose file

- **--timeout**, **-t** (int) - Shutdown timeout in seconds (default: 10)

- **--volumes**, **-v** (bool) - Remove named volumes and anonymous volumes




**Examples:**

```bash
dev-stack down
```
Stop all running services


```bash
dev-stack down postgres redis
```
Stop specific services


```bash
dev-stack down --volumes
```
Stop services and remove volumes


```bash
dev-stack down --timeout 5
```
Stop services with custom timeout





**Tips:**

- Use --volumes carefully as it will delete all data

- Add --remove-orphans to clean up unused containers




**See also:** up, cleanup, status


---


#### restart

**Usage:** `restart [service...]`

Restart development stack services


Restart one or more services. This is equivalent to running down followed
by up, but more efficient for quick restarts.






**Flags:**

- **--no-deps** (bool) - Don't restart linked services

- **--timeout**, **-t** (int) - Restart timeout in seconds (default: 10)




**Examples:**

```bash
dev-stack restart
```
Restart all services


```bash
dev-stack restart postgres
```
Restart a specific service


```bash
dev-stack restart --timeout 5
```
Restart with custom timeout







**See also:** up, down, status


---


#### scale

**Usage:** `scale <service=replicas>...`

Scale services up or down


Scale the number of running instances for one or more services.
Useful for load testing and development scenarios requiring multiple
service instances.






**Flags:**

- **--no-recreate** (bool) - Don't recreate existing containers

- **--timeout**, **-t** (int) - Timeout for scaling operation (default: 30)




**Examples:**

```bash
dev-stack scale postgres=2
```
Scale postgres to 2 instances


```bash
dev-stack scale redis=3 postgres=1
```
Scale multiple services


```bash
dev-stack scale --timeout 60 postgres=2
```
Scale with custom timeout







**See also:** up, down, status


---




### 🧹 Maintenance & Cleanup

Commands for cleanup, initialization, and maintenance


#### cleanup

**Usage:** `cleanup [options]`

Clean up unused resources and data


Clean up unused Docker resources, temporary files, and orphaned data
created by dev-stack services. Helps reclaim disk space and maintain
a clean development environment.






**Flags:**

- **--all**, **-a** (bool) - Clean up all resources (containers, volumes, images)

- **--dry-run** (bool) - Show what would be cleaned without doing it

- **--force**, **-f** (bool) - Don't prompt for confirmation

- **--images**, **-i** (bool) - Remove unused images

- **--networks**, **-n** (bool) - Remove unused networks

- **--volumes**, **-v** (bool) - Remove unused volumes




**Examples:**

```bash
dev-stack cleanup
```
Interactive cleanup with confirmations


```bash
dev-stack cleanup --all --force
```
Clean up everything without prompts


```bash
dev-stack cleanup --dry-run
```
Preview what would be cleaned up





**Tips:**

- Use --dry-run first to see what will be removed

- Be careful with --volumes as it removes all data




**See also:** down, doctor


---


#### init

**Usage:** `init [project-type]`

Initialize a new dev-stack project


Initialize a new dev-stack project in the current directory. Creates
configuration files, directory structure, and optional service
configurations based on project type.






**Flags:**

- **--force**, **-f** (bool) - Overwrite existing files

- **--minimal** (bool) - Create minimal configuration

- **--name**, **-n** (string) - Project name

- **--template**, **-t** (string) - Project template to use (default: basic)




**Examples:**

```bash
dev-stack init
```
Interactive project initialization


```bash
dev-stack init web
```
Initialize with web development template


```bash
dev-stack init --name myproject microservices
```
Initialize microservices project with custom name







**See also:** docs, validate


---


#### version

**Usage:** `version`

Show version information


Display version information for dev-stack CLI, Docker, and managed
services. Includes build information and dependency versions.






**Flags:**

- **--check-updates** (bool) - Check for available updates

- **--format** (string) - Output format (table|json|yaml) (default: table)

- **--full**, **-f** (bool) - Show detailed version information




**Examples:**

```bash
dev-stack version
```
Show basic version information


```bash
dev-stack version --full
```
Show detailed version and build info


```bash
dev-stack version --check-updates
```
Check for available updates







**See also:** doctor


---




### 📊 Monitoring & Observability

Commands for monitoring services and viewing logs


#### status

**Usage:** `status [service...]`

Show status of development stack services


Display comprehensive status information for services including running
state, health checks, resource usage, and port mappings. Supports multiple
output formats and real-time monitoring.




**Aliases:** ps, ls



**Flags:**

- **--filter** (string) - Filter services by status

- **--format**, **-f** (string) - Output format (table|json|yaml) (default: table)

- **--no-trunc** (bool) - Don't truncate output

- **--quiet**, **-q** (bool) - Only show service names and basic status

- **--watch**, **-w** (bool) - Watch for status changes




**Examples:**

```bash
dev-stack status
```
Show status of all services


```bash
dev-stack status postgres redis
```
Show status of specific services


```bash
dev-stack status --format json
```
Output status in JSON format


```bash
dev-stack status --watch
```
Watch for status changes in real-time


```bash
dev-stack status --filter running
```
Show only running services





**Tips:**

- Use --watch to monitor services in real-time

- Try --format json for programmatic access

- Use --filter to focus on specific service states




**See also:** logs, monitor, doctor


---


#### logs

**Usage:** `logs [service...]`

View logs from services


View and follow logs from one or more services. Supports filtering,
timestamps, and real-time following. Logs from multiple services are
color-coded for easy identification.






**Flags:**

- **--follow**, **-f** (bool) - Follow log output in real-time

- **--no-color** (bool) - Disable colored output

- **--no-prefix** (bool) - Don't show service name prefix

- **--since** (string) - Show logs since timestamp or relative time

- **--tail**, **-t** (string) - Number of lines to show from end of logs (default: all)

- **--timestamps** (bool) - Show timestamps in log output




**Examples:**

```bash
dev-stack logs
```
Show logs from all services


```bash
dev-stack logs postgres redis
```
Show logs from specific services


```bash
dev-stack logs --follow postgres
```
Follow logs from postgres in real-time


```bash
dev-stack logs --tail 100 --since 1h
```
Show last 100 lines from the past hour





**Tips:**

- Use --follow to see logs in real-time

- Combine --tail and --since for targeted log viewing

- Use --timestamps to correlate events across services




**See also:** status, monitor


---


#### monitor

**Usage:** `monitor [service...]`

Real-time monitoring dashboard for services


Launch an interactive monitoring dashboard showing real-time metrics,
logs, and status for all services. Provides a unified view of your
development stack health.






**Flags:**

- **--compact** (bool) - Use compact display mode

- **--no-logs** (bool) - Don't show log streams

- **--refresh**, **-r** (int) - Refresh interval in seconds (default: 2)




**Examples:**

```bash
dev-stack monitor
```
Monitor all services


```bash
dev-stack monitor postgres redis
```
Monitor specific services


```bash
dev-stack monitor --refresh 5
```
Monitor with custom refresh interval







**See also:** status, logs, doctor


---


#### doctor

**Usage:** `doctor [service...]`

Diagnose and troubleshoot stack health


Run comprehensive health checks on your development stack. Identifies
common issues, provides troubleshooting suggestions, and validates
service configurations.






**Flags:**

- **--fix** (bool) - Attempt to automatically fix issues

- **--format**, **-f** (string) - Output format (table|json) (default: table)

- **--verbose**, **-v** (bool) - Show detailed diagnostic information




**Examples:**

```bash
dev-stack doctor
```
Run health checks on all services


```bash
dev-stack doctor postgres
```
Diagnose a specific service


```bash
dev-stack doctor --fix
```
Attempt to fix detected issues





**Tips:**

- Run doctor when services aren't behaving as expected

- Use --fix to attempt automatic resolution of common issues




**See also:** status, logs


---




## Help and Support


## Common Tasks

**Start services for web development:**
dev-stack up --profile web

**Connect to database:**
dev-stack connect postgres

**View real-time logs:**
dev-stack logs --follow

**Backup your data:**
dev-stack backup

**Clean up resources:**
dev-stack cleanup --dry-run




## Troubleshooting

**Services won't start:**
- Run: dev-stack doctor
- Check: dev-stack logs <service>

**Port conflicts:**
- Use: dev-stack status
- Modify port mappings in config

**Out of disk space:**
- Run: dev-stack cleanup --dry-run
- Then: dev-stack cleanup --all

**Performance issues:**
- Check: dev-stack monitor
- Scale services: dev-stack scale <service>=<count>



<!-- AUTO-GENERATED-END -->