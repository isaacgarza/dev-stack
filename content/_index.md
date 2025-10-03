---
title: "dev-stack"
description: "A powerful development stack management tool built in Go for streamlined local development automation"
---

# Welcome to dev-stack

A powerful CLI tool built in Go that helps you quickly set up, manage, and tear down development environments with consistent configurations across your team.

## 🎉 Migration Complete

**dev-stack** has successfully migrated from Python to a pure Go implementation with enhanced performance, comprehensive CLI features, and no external language dependencies.

## What is dev-stack?

**dev-stack** is a modern CLI tool built in Go that provides:

- **🚀 Quick Setup**: Initialize development environments in seconds
- **🐳 Docker Integration**: Seamless container management for services
- **⚙️ Configuration Management**: Consistent setups across teams
- **🔧 Extensible**: Support for multiple project types and services
- **📊 Monitoring**: Built-in health checks and status monitoring
- **🛠️ Developer Experience**: Intuitive commands and helpful diagnostics

## Quick Start

### Installation

**Download Binary (Recommended)**
```bash
# Download the latest release for your platform
curl -L -o dev-stack "https://github.com/isaacgarza/dev-stack/releases/latest/download/dev-stack-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
chmod +x dev-stack
sudo mv dev-stack /usr/local/bin/
```

**Build from Source**
```bash
git clone https://github.com/isaacgarza/dev-stack.git
cd dev-stack
make build
sudo cp build/dev-stack /usr/local/bin/
```

**Go Install**
```bash
go install github.com/isaacgarza/dev-stack/cmd/dev-stack@latest
```

### Basic Usage

```bash
# Check system health
dev-stack doctor

# Initialize a new project
dev-stack init go --name my-app

# Start your development stack
dev-stack up

# Check status
dev-stack status

# Stop when done
dev-stack down
```

## Key Features

### Project Templates

Initialize projects with pre-configured templates:

- **Go**: Modern Go applications with hot reload
- **Node.js**: JavaScript/TypeScript applications
- **Python**: Python applications with virtual environments
- **Full-Stack**: Complete frontend/backend setups

### Service Management

Easily manage common development services:

- **Databases**: PostgreSQL, MySQL, MongoDB, Redis
- **Message Queues**: RabbitMQ, Apache Kafka
- **Monitoring**: Prometheus, Grafana
- **Search**: Elasticsearch, OpenSearch
- **And many more...**

### Health Monitoring

Built-in health checks ensure your services are running correctly:

```bash
# Check overall system health
dev-stack doctor

# Monitor service status
dev-stack status --watch

# View detailed service logs
dev-stack logs <service-name>
```

## Example Workflows

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

### Adding Services to Existing Project

```bash
# Add Redis for caching
dev-stack service add redis

# Add monitoring stack
dev-stack service add prometheus grafana

# Restart to apply changes
dev-stack restart
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

## Community & Support

- **GitHub**: [isaacgarza/dev-stack](https://github.com/isaacgarza/dev-stack)
- **Issues**: Report bugs and feature requests
- **Discussions**: Share ideas and get help
- **Contributing**: See our [contribution guide]({{< ref "/contributing" >}})

## What's Next?

1. **[Get started with installation]({{< ref "/getting-started" >}})**
2. **[Learn basic usage]({{< ref "/usage" >}})**
3. **[Explore CLI commands]({{< ref "/cli-reference" >}})**
4. **[Discover available services]({{< ref "/services" >}})**

---

> **Need Help?** If you encounter any issues, check our [troubleshooting section]({{< ref "/usage#troubleshooting" >}}) or [open an issue](https://github.com/isaacgarza/dev-stack/issues) on GitHub.