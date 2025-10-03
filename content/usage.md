---
title: "Usage Guide"
description: "Complete guide to using dev-stack for development environment management"
weight: 20
---

# Usage Guide

This comprehensive guide covers how to use dev-stack effectively for managing your development environments.

## Core Concepts

### Projects
A **project** in dev-stack is a directory containing your application code and a `dev-stack-config.yaml` file that defines your development environment.

### Services
**Services** are containerized applications like databases, message queues, or monitoring tools that your project depends on.

### Templates
**Templates** are pre-configured project structures that include common setups for different technology stacks.

### Stacks
A **stack** is the combination of your project and all its running services.

## Project Lifecycle

### 1. Initialize a New Project

Create a new project from a template:

```bash
# Initialize with Go template
dev-stack init go --name my-api

# Initialize with specific options
dev-stack init go \
  --name my-api \
  --with-database postgres \
  --with-cache redis \
  --port 8080
```

Available templates:
- `go` - Go applications with hot reload
- `node` - Node.js/TypeScript applications  
- `python` - Python applications with virtual environments
- `fullstack` - Complete frontend/backend setup
- `minimal` - Basic Docker setup

### 2. Understanding Project Structure

After initialization, your project will have:

```
my-api/
├── dev-stack-config.yaml    # Main configuration
├── docker-compose.yml       # Generated compose file
├── .env                     # Environment variables
├── src/                     # Your application code
├── scripts/                 # Helper scripts
└── docs/                    # Project documentation
```

### 3. Start Your Development Environment

```bash
# Start all services
dev-stack up

# Start specific services only
dev-stack up database cache

# Start in background
dev-stack up --detach

# Start with fresh containers
dev-stack up --force-recreate
```

### 4. Monitor Your Environment

```bash
# Check status of all services
dev-stack status

# Watch status in real-time
dev-stack status --watch

# Get detailed information
dev-stack info
```

### 5. Stop Your Environment

```bash
# Stop all services
dev-stack down

# Stop and remove volumes
dev-stack down --volumes

# Stop specific services
dev-stack down database
```

## Service Management

### Adding Services

Add services to your existing project:

```bash
# Add a database
dev-stack service add postgres

# Add multiple services
dev-stack service add redis rabbitmq

# Add with custom configuration
dev-stack service add postgres --port 5433 --database myapp
```

### Listing Available Services

```bash
# List all available services
dev-stack service list

# List only database services
dev-stack service list --category database

# Show service details
dev-stack service info postgres
```

### Removing Services

```bash
# Remove a service
dev-stack service remove redis

# Remove service and its data
dev-stack service remove redis --volumes
```

### Service Categories

**Databases:**
- `postgres` - PostgreSQL database
- `mysql` - MySQL database  
- `mongodb` - MongoDB database
- `redis` - Redis cache/database

**Message Queues:**
- `rabbitmq` - RabbitMQ message broker
- `kafka` - Apache Kafka
- `nats` - NATS messaging

**Monitoring:**
- `prometheus` - Metrics collection
- `grafana` - Metrics visualization
- `jaeger` - Distributed tracing

**Search:**
- `elasticsearch` - Elasticsearch
- `opensearch` - OpenSearch

**Web Servers:**
- `nginx` - Nginx web server
- `traefik` - Traefik reverse proxy

## Configuration Management

### Project Configuration

Edit `dev-stack-config.yaml` to customize your environment:

```yaml
# Project metadata
name: my-api
version: 1.0.0
description: My awesome API

# Application settings
app:
  port: 8080
  environment: development
  hot_reload: true

# Services configuration
services:
  postgres:
    port: 5432
    database: myapp
    username: developer
    password: secret
  
  redis:
    port: 6379
    
# Environment variables
env:
  DATABASE_URL: postgres://developer:secret@postgres:5432/myapp
  REDIS_URL: redis://redis:6379
  
# Port mappings
ports:
  - "8080:8080"  # App port
  - "5432:5432" # PostgreSQL
  - "6379:6379" # Redis
```

### Environment Variables

Manage environment variables:

```bash
# Set environment variable
dev-stack env set DATABASE_URL "postgres://localhost:5432/myapp"

# Get environment variable
dev-stack env get DATABASE_URL

# List all environment variables
dev-stack env list

# Load from .env file
dev-stack env load .env.local
```

### Global Configuration

Configure dev-stack globally:

```bash
# Set default template
dev-stack config set default-template go

# Set default registry
dev-stack config set registry myregistry.com

# View current configuration
dev-stack config show

# Reset to defaults
dev-stack config reset
```

## Health Checks and Monitoring

### System Health

```bash
# Run comprehensive health check
dev-stack doctor

# Check specific components
dev-stack doctor --check docker
dev-stack doctor --check network
dev-stack doctor --check disk
```

### Service Health

```bash
# Check service health
dev-stack health

# Check specific service
dev-stack health postgres

# Set up health check endpoints
dev-stack health setup
```

### Logs and Debugging

```bash
# View logs for all services
dev-stack logs

# View logs for specific service
dev-stack logs postgres

# Follow logs in real-time
dev-stack logs --follow

# View last N lines
dev-stack logs --tail 100

# Show timestamps
dev-stack logs --timestamps
```

## Advanced Usage

### Custom Services

Create custom service definitions:

```bash
# Generate custom service template
dev-stack service create myservice

# Edit the generated service file
# Edit services/myservice.yaml
```

Example custom service (`services/myservice.yaml`):

```yaml
name: myservice
description: My custom service
category: custom

container:
  image: myregistry/myservice:latest
  ports:
    - "9000:9000"
  environment:
    - SERVICE_PORT=9000
  volumes:
    - ./data:/data
  
health_check:
  endpoint: /health
  interval: 30s
  retries: 3
```

### Multi-Environment Support

Manage different environments:

```bash
# Switch to development environment
dev-stack env switch development

# Switch to staging environment  
dev-stack env switch staging

# Create new environment
dev-stack env create testing
```

### Networking

Configure service networking:

```bash
# List networks
dev-stack network list

# Create custom network
dev-stack network create mynetwork

# Connect service to network
dev-stack network connect mynetwork postgres
```

### Volume Management

Manage persistent data:

```bash
# List volumes
dev-stack volume list

# Create volume
dev-stack volume create mydata

# Backup volume
dev-stack volume backup postgres-data

# Restore volume
dev-stack volume restore postgres-data backup.tar.gz
```

## Workflows and Best Practices

### Daily Development Workflow

1. **Start your day:**
   ```bash
   dev-stack status  # Check if anything is running
   dev-stack up      # Start your environment
   ```

2. **During development:**
   ```bash
   dev-stack logs --follow app  # Monitor application logs
   dev-stack health             # Check service health
   ```

3. **End of day:**
   ```bash
   dev-stack down  # Stop services
   ```

### Team Collaboration

Share your environment with team members:

```bash
# Export configuration
dev-stack config export > team-config.yaml

# Team member imports it
dev-stack config import team-config.yaml
```

### Continuous Integration

Use dev-stack in CI/CD:

```bash
# In your CI script
dev-stack up --detach
dev-stack health --wait-timeout 60s
# Run your tests
dev-stack down
```

## Troubleshooting

### Common Issues

**Services won't start:**
```bash
# Check Docker status
docker ps

# Check logs for errors
dev-stack logs

# Verify configuration
dev-stack config validate
```

**Port conflicts:**
```bash
# Check what's using the port
lsof -i :8080

# Change port in configuration
dev-stack config set app.port 8081
```

**Out of disk space:**
```bash
# Clean up old containers and volumes
dev-stack cleanup

# Remove unused Docker resources
docker system prune
```

**Network issues:**
```bash
# Reset networks
dev-stack network reset

# Check connectivity
dev-stack doctor --check network
```

### Performance Optimization

**Speed up startup:**
```bash
# Use cached images
dev-stack config set use-cache true

# Parallelize service startup
dev-stack config set parallel-start true
```

**Reduce resource usage:**
```bash
# Limit service resources
dev-stack service config postgres --memory 512m --cpus 0.5
```

### Getting Help

```bash
# Show help for any command
dev-stack help [command]

# Show detailed command options
dev-stack [command] --help

# Enable debug mode
dev-stack --debug [command]

# Report issues
dev-stack issue create
```

## Integration Examples

### With VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Start dev-stack",
      "type": "shell",
      "command": "dev-stack up",
      "group": "build"
    },
    {
      "label": "Stop dev-stack", 
      "type": "shell",
      "command": "dev-stack down",
      "group": "build"
    }
  ]
}
```

### With Task

Add to `Taskfile.yml`:

```yaml
tasks:
  dev-up:
    desc: Start development environment
    cmds:
      - dev-stack up

  dev-down:
    desc: Stop development environment
    cmds:
      - dev-stack down

  dev-status:
    desc: Check development environment status
    cmds:
      - dev-stack status

  dev-logs:
    desc: Follow development environment logs
    cmds:
      - dev-stack logs --follow
```

### With Git Hooks

Add to `.git/hooks/post-checkout`:

```bash
#!/bin/bash
if [ -f dev-stack-config.yaml ]; then
  echo "Starting dev-stack environment..."
  dev-stack up --detach
fi
```

## Next Steps

- **[Explore CLI Reference]({{< ref "/cli-reference" >}})** - Complete command documentation
- **[Browse Available Services]({{< ref "/services" >}})** - Discover all supported services  
- **[Contributing Guide]({{< ref "/contributing" >}})** - Help improve dev-stack
- **[Join Discussions](https://github.com/isaacgarza/dev-stack/discussions)** - Connect with the community

---

> **Pro Tip**: Use `dev-stack status --watch` to monitor your services in real-time during development!