---
title: "Usage & Management"
description: "Daily usage patterns, service management commands, and common workflows for dev-stack"
lead: "Learn how to effectively use dev-stack for your daily development workflow"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 20
toc: true

---

# Usage & Management Guide (dev-stack)

This guide covers daily usage patterns, service management commands, and common workflows for **dev-stack**.

---

## ✅ Quick Checklist

- [ ] Setup your environment ([Setup Guide](setup.md))
- [ ] Configure your stack ([Configuration Guide](configuration.md))
- [ ] Start services ([README](../README.md))
- [ ] Manage services ([reference.md](reference.md))
- [ ] Troubleshoot issues ([Troubleshooting Guide](troubleshooting.md))
- [ ] Integrate with your app ([Integration Guide](integration.md))

---

## 📋 Overview

**dev-stack** provides two main scripts for different purposes:
- **`setup.sh`**: Initial configuration and environment setup
- **`manage.sh`**: Ongoing service management and operations

For a quick start and full command reference, see the [README](../README.md).

## 🚀 Common Workflows

### Starting a New Go Project

```bash
# Initialize with Go template
dev-stack init go --name my-api --with-database postgres

# Start the development environment
dev-stack up

# Your Go application is now running with:
# - Hot reload enabled
# - PostgreSQL database
# - Health checks configured
```

**Project structure created:**
```
my-api/
├── main.go
├── go.mod
├── dev-stack-config.yaml
├── docker-compose.override.yml
└── .env.local
```

### Adding Services to Existing Project

```bash
# Add Redis for caching
dev-stack service add redis

# Add monitoring stack
dev-stack service add prometheus grafana

# Add message queue
dev-stack service add kafka

# Restart to apply changes
dev-stack restart
```

**Verify services are running:**
```bash
dev-stack status
dev-stack health redis
```

### Team Collaboration

Share your development environment configuration:

```bash
# Export current configuration
dev-stack config export > dev-stack.yaml

# Team members can import it
dev-stack config import dev-stack.yaml
dev-stack up
```

**Team workflow best practices:**
- Commit `dev-stack-config.yaml` to version control
- Use `.env.local` for personal overrides (don't commit)
- Document service dependencies in README
- Use `dev-stack doctor` to verify team setup

### Database Development Workflow

```bash
# Start with database services
dev-stack init --services postgres,redis
dev-stack up

# Run migrations
dev-stack exec postgres psql -U postgres -d myapp < migrations/001_initial.sql

# Backup data for testing
dev-stack backup postgres

# Reset database for clean testing
dev-stack reset postgres
dev-stack restore postgres backup-20241006.sql
```

### Microservices Development

```bash
# Initialize with multiple services
dev-stack init --name order-service --services postgres,kafka,jaeger
dev-stack service add prometheus grafana

# Start everything
dev-stack up

# Monitor distributed traces
open http://localhost:16686  # Jaeger UI

# View metrics
open http://localhost:3000   # Grafana UI
```

### Daily Development Workflow

**Start of day:**
```bash
# Quick health check
dev-stack doctor

# Start your stack
dev-stack up

# Check service status
dev-stack status
```

**During development:**
```bash
# View logs for debugging
dev-stack logs api
dev-stack logs postgres

# Monitor resource usage
dev-stack stats

# Reset data for testing
dev-stack reset redis
```

**End of day:**
```bash
# Stop services
dev-stack down

# Or pause (keeps data)
dev-stack pause
```

## 🛠 Setup Commands

See [README](../README.md) and [Configuration Guide](configuration.md) for setup and configuration commands.

### Configuration Options

See [Configuration Guide](configuration.md) for all available options and overrides.

### Instance Management

See [README](../README.md) for instance management commands.

### Advanced Setup Options

See [Configuration Guide](configuration.md) for advanced setup options.

## 🎛 Management Commands

See [README](../README.md) and [reference.md](reference.md) for all management commands.

### Service Information

See [README](../README.md) and [services.md](services.md) for service info and status commands.

### Logging and Monitoring

See [README](../README.md) and [troubleshooting.md](troubleshooting.md) for logging and monitoring commands.

### Service Interaction

See [services.md](services.md) for service CLI and exec commands.

### Data Management

See [usage.md](usage.md) and [reference.md](reference.md) for backup, restore, and data management commands.

### Maintenance

See [usage.md](usage.md) and [reference.md](reference.md) for update and cleanup commands.

## 📊 Multi-Repository Workflows

See [README](../README.md) and [setup.md](setup.md) for multi-repo usage and resource management workflows.

## 🔧 Configuration Management

See [configuration.md](configuration.md) for runtime config changes, environment-specific configs, and validation.

## 🧪 Testing Workflows

See [integration.md](integration.md) and [configuration.md](configuration.md) for integration, CI/CD, and database testing workflows.

## 🔍 Debugging and Troubleshooting

See [troubleshooting.md](troubleshooting.md) for health checks, log analysis, network debugging, and performance tips.

## 📈 Performance Optimization

See [configuration.md](configuration.md) and [usage.md](usage.md) for resource tuning, service optimization, and speed tips.

## 🔄 Update and Maintenance

See [contributing.md](contributing.md) for update and maintenance workflows.

## 📚 Integration Examples

See [integration.md](integration.md) for application integration patterns and Spring Boot examples.

## 🆘 Getting Help

See [README](../README.md) and [reference.md](reference.md) for help commands and quick reference.

## 🎯 What's Next?

After mastering these workflows, explore advanced dev-stack features:

1. **[Configure advanced settings](configuration.md)** - Custom ports, environment variables, service options
2. **[Integrate with your applications](integration.md)** - Spring Boot, Node.js, and other framework examples
3. **[Set up monitoring and observability](services.md#monitoring-stack)** - Prometheus, Grafana, Jaeger
4. **[Troubleshoot common issues](troubleshooting.md)** - Debug problems and optimize performance

**Pro tips:**
- Use `dev-stack config validate` to check your configuration
- Set up shell completion: `dev-stack completion bash`
- Create project templates for your team's common stacks

**Share your setup:** Export configurations with `dev-stack config export` for team collaboration.

## 🗂️ See Also

- [README](../README.md)
- [Setup Guide](setup.md)
- [Configuration Guide](configuration.md)
- [Services Guide](services.md)
- [Integration Guide](integration.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Quick Reference](reference.md)