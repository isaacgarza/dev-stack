# Dev Stack - Local Development Environment Framework

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Docker](https://img.shields.io/badge/Docker-Compose-blue.svg)](https://www.docker.com/) [![GitHub issues](https://img.shields.io/github/issues/isaacgarza/dev-stack)]

A modular, configuration-driven framework for setting up local development environments with Docker Compose. Define your services in a YAML configuration file and get productive instantly.

## 🚀 Quick Start

```bash
# 1. Ensure Docker is running
docker info

# 2. Initialize configuration file
./scripts/setup.sh --init

# 3. Edit the configuration file
vim local-dev-config.yaml

# 4. Set up and start services
./scripts/setup.sh

# 5. View service information
./scripts/manage.sh info
```

## 🎯 Overview

This framework solves the common problem of setting up consistent local development environments across different projects and team members. Instead of maintaining separate Docker Compose files for each project, teams can:

- **Configure services declaratively**: Define your stack in `local-dev-config.yaml`
- **Choose services à la carte**: Redis, PostgreSQL, MySQL, Jaeger, Prometheus, LocalStack, etc.
- **Generate configurations**: Docker Compose, environment files, and Spring Boot configs
- **Manage everything**: Start, stop, backup, restore, and monitor services
- **Stay consistent**: Same setup across all team members and projects

### Key Benefits

- ✅ **Configuration-driven**: Single YAML file defines your entire stack
- ✅ **Plug-and-play**: Add to any project in minutes
- ✅ **Consistent**: Same environment across team members
- ✅ **Flexible**: Mix and match services as needed
- ✅ **Production-ready**: Services configured with best practices
- ✅ **Spring Boot integration**: Auto-generated configuration
- ✅ **Smart validation**: Warns about resource usage and conflicts

## 🛠 Available Services

### Databases
- **PostgreSQL** - Primary database with custom DB/user creation
- **MySQL** - Alternative database option

### Caching & Storage
- **Redis** - In-memory data structure store

### Observability
- **Jaeger** - Distributed tracing
- **Prometheus** - Metrics collection and monitoring

### Cloud Services
- **LocalStack** - AWS services emulation (SQS, SNS, DynamoDB, S3, etc.)

### Messaging
- **Kafka** - Event streaming platform with custom topic creation

## 📖 Documentation

This documentation is organized into focused sections for easy navigation:

### 🏗️ Getting Started
- **[Setup & Installation](docs/setup.md)** - Docker setup, prerequisites, and installation
- **[Configuration Guide](docs/configuration.md)** - Complete configuration reference and examples
- **[Available Services](docs/services.md)** - Detailed service configurations and options

### 💻 Development
- **[Usage & Management](docs/usage.md)** - Daily commands, service management, and workflows
- **[Integration Guide](docs/integration.md)** - Spring Boot, IDE, and framework integrations
- **[Advanced Workflows](docs/advanced.md)** - Multi-repository usage, CI/CD, and team collaboration

### 🔧 Support
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues, debugging, and performance tips
- **[Contributing](docs/contributing.md)** - How to add services and contribute to the framework
- **[Quick Reference](docs/reference.md)** - Commands cheatsheet and port reference

## ⚙️ Basic Configuration

The framework uses a `local-dev-config.yaml` file to define your development stack:

```yaml
# Basic configuration example
project:
  name: "my-api"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger
    - localstack

overrides:
  postgres:
    database: "my_api_dev"
    username: "api_user"
  localstack:
    services: ["sqs", "sns"]
```

## 🎛 Management Commands

```bash
# Configuration
./scripts/setup.sh --init                    # Create sample config
./scripts/setup.sh                          # Setup from config
./scripts/setup.sh --services=redis,postgres # Override services

# Service Management
./scripts/manage.sh start                    # Start all services
./scripts/manage.sh stop                     # Stop all services
./scripts/manage.sh restart                  # Restart services
./scripts/manage.sh info                     # Show service info

# Instance Management
./scripts/manage.sh list-all                 # List all instances
./scripts/manage.sh cleanup-all              # Cleanup all instances
```

## 🔄 Multi-Repository Usage

The framework intelligently detects when you're running multiple instances across different repositories and provides options to:

- **Connect** to existing running services
- **Cleanup** existing instances and start fresh
- **Cancel** and resolve conflicts manually

This prevents resource conflicts and ensures smooth development across multiple projects.

## 📄 Generated Files

When you run setup, the framework generates:

- `docker-compose.generated.yml` - Complete Docker Compose configuration
- `.env.generated` - Environment variables for services
- `application-local.yml.generated` - Spring Boot configuration (if applicable)

## 🤝 Contributing

We welcome contributions! Please see the [Contributing Guide](docs/contributing.md) for details on:

- Adding new services
- Improving existing configurations
- Documentation updates
- Bug fixes and enhancements

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙋 Support

- **Issues**: Report bugs or request features via GitHub Issues
- **Documentation**: Check the [docs](docs/) directory for detailed guides
- **Team**: Reach out to the development tools team for support

---

**Next Steps**: Start with the [Setup Guide](docs/setup.md) to get Docker configured, then follow the [Configuration Guide](docs/configuration.md) to customize your development environment.
