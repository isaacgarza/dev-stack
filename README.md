# dev-stack
> A powerful development stack management tool built in Go for streamlined local development automation.

## Overview

**dev-stack** is a modern CLI tool that helps you quickly set up, manage, and tear down development environments with consistent configurations across your team. Built in Go for performance and reliability, it provides a unified interface for managing Docker-based development stacks.

## 🚀 Quick Start

### Installation

#### Option 1: Download Binary (Recommended)
```bash
# Download the latest release for your platform
curl -L -o dev-stack "https://github.com/isaacgarza/dev-stack/releases/latest/download/dev-stack-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
chmod +x dev-stack
sudo mv dev-stack /usr/local/bin/
```

#### Option 2: Build from Source
```bash
git clone https://github.com/isaacgarza/dev-stack.git
cd dev-stack
make build
sudo cp build/dev-stack /usr/local/bin/
```

#### Option 3: Go Install
```bash
go install github.com/isaacgarza/dev-stack/cmd/dev-stack@latest
```

### Quick Setup
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

## 🛠️ Features

- **🐳 Docker Integration**: Seamless Docker and Docker Compose management
- **🏗️ Project Templates**: Quick setup with pre-configured project types
- **📊 Health Monitoring**: Built-in health checks and system diagnostics
- **🔄 Version Management**: Intelligent version detection and management
- **🎨 Rich CLI**: Beautiful, interactive command-line interface
- **⚡ Fast**: Built in Go for optimal performance
- **🔧 Extensible**: Plugin architecture for custom functionality

## 📋 Commands

| Command | Description |
|---------|-------------|
| `dev-stack init` | Initialize a new development stack project |
| `dev-stack up` | Start development stack services |
| `dev-stack down` | Stop development stack services |
| `dev-stack status` | Show status of services |
| `dev-stack doctor` | Run system health checks |
| `dev-stack version` | Show version information |

### Examples

```bash
# Initialize different project types
dev-stack init go --name my-go-app
dev-stack init node --name my-node-app
dev-stack init python --name my-python-app
dev-stack init fullstack --name my-fullstack-app

# Start specific services
dev-stack up postgres redis
dev-stack up --build --detach

# Monitor services
dev-stack status --watch
dev-stack status --format json

# Health diagnostics
dev-stack doctor --verbose
dev-stack doctor --check docker
dev-stack doctor --fix
```

## 🏗️ Project Structure

```
dev-stack/
├── cmd/dev-stack/          # CLI entry point
├── internal/
│   ├── cli/               # CLI commands
│   ├── core/              # Core business logic
│   │   ├── config/        # Configuration management
│   │   ├── docker/        # Docker integration
│   │   ├── project/       # Project detection
│   │   ├── services/      # Service management
│   │   └── version/       # Version management
│   ├── pkg/               # Shared packages
│   └── templates/         # Project templates
├── scripts/               # Build and development scripts
├── .github/workflows/     # CI/CD pipelines
├── Makefile              # Build system
└── README.md
```

## 🔧 Development

### Prerequisites
- Go 1.21 or later
- Docker and Docker Compose
- Make

### Build Commands
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linting
make lint

# Install locally
make install

# Development mode
make dev

# Watch for changes
make watch
```

### Running Tests
```bash
# Unit tests
make test-go

# Integration tests
make test-go-integration

# All tests
make test
```

## 📚 Configuration

dev-stack uses YAML configuration files for project setup:

```yaml
# dev-stack-config.yaml
global:
  default_project_type: "go"
  log_level: "info"
  color_output: true

projects:
  my-app:
    type: "go"
    services:
      - name: "app"
        build: "./Dockerfile"
        ports: ["8080:8080"]
      - name: "postgres"
        image: "postgres:15"
        environment:
          POSTGRES_DB: "myapp"
          POSTGRES_USER: "user"
          POSTGRES_PASSWORD: "password"
```

## 🐳 Docker Integration

dev-stack seamlessly integrates with Docker and Docker Compose:

- **Service Management**: Start, stop, and monitor Docker services
- **Health Checks**: Built-in health monitoring for containers
- **Network Management**: Automatic network setup and configuration
- **Volume Management**: Persistent data management
- **Build Integration**: Automatic image building and caching

## 🎯 Project Templates

Supported project types:
- **go**: Go application with Docker
- **node**: Node.js application with Docker
- **python**: Python application with Docker
- **fullstack**: Multi-service full-stack application

Each template includes:
- Dockerfile and docker-compose.yml
- Development and production configurations
- Health checks and monitoring
- Best practice directory structure

## 📖 Legacy Python Support

The repository also contains the legacy Python implementation in the `scripts/` directory. While the Go implementation is the primary focus, Python scripts are maintained for backward compatibility.

### Python Setup (Legacy)
```bash
# Set up Python environment
make setup

# Generate documentation
make docs

# Run Python tests
make test-python
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linting: `make lint`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 [Documentation](docs/)
- 🐛 [Issues](https://github.com/isaacgarza/dev-stack/issues)
- 💬 [Discussions](https://github.com/isaacgarza/dev-stack/discussions)

## 🚀 Roadmap

- [ ] Plugin system for extensibility
- [ ] Advanced version management
- [ ] Team collaboration features
- [ ] Cloud deployment integration
- [ ] Performance monitoring
- [ ] Auto-update system

---

> **Built with ❤️ by the dev-stack team**  
> Making local development environments simple, consistent, and powerful.