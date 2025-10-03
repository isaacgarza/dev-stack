# dev-stack
> A powerful development stack management tool built in Go for streamlined local development automation.

> **🎉 Migration Complete**: dev-stack has successfully migrated from Python to a pure Go implementation with enhanced performance, comprehensive CLI features, and no external language dependencies.

## Overview

**dev-stack** is a modern CLI tool that helps you quickly set up, manage, and tear down development environments with consistent configurations across your team. Built in Go for performance and reliability, it provides a unified interface for managing Docker-based development stacks.

## 🚀 Quick Start

### Installation

#### Prerequisites

**Install Task (Build Tool)**
```bash
# macOS
brew install go-task/tap/go-task

# Linux/macOS (direct download)
curl -sL https://taskfile.dev/install.sh | sh -s -- -d -b ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"

# Verify installation
task --version
```

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
task build
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
| `dev-stack docs` | Generate documentation from YAML manifests |
| `dev-stack logs` | View service logs |
| `dev-stack exec` | Execute commands in running containers |
| `dev-stack scale` | Scale services up or down |
| `dev-stack backup` | Create backups of service data |
| `dev-stack restore` | Restore service data from backups |
| `dev-stack cleanup` | Clean up stopped containers and resources |
| `dev-stack connect` | Connect to running services |
| `dev-stack monitor` | Monitor service health and performance |

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
├── Taskfile.yml          # Build system
└── README.md
```

## 🔧 Development

### Prerequisites
- Go 1.21 or later (for development)
- Docker and Docker Compose
- Make (build system)

> **Note**: End users only need Docker - the dev-stack binary is self-contained with no runtime dependencies.

### Build Commands
```bash
# Build for current platform
task build

# Build for all platforms
task build-all

# Run tests
task test

# Run linting
task lint

# Install locally
task install

# Development mode
task dev

# Watch for changes
task watch
```

### Running Tests
```bash
# Unit tests
task test

# Integration tests
task test-integration

# All tests
task test
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

dev-stack provides project scaffolding templates for different application types. All templates are managed by the Go-based CLI:

- **go**: Go application with Docker
- **node**: Node.js application with Docker  
- **python**: Python application with Docker
- **fullstack**: Multi-service full-stack application

Each template includes:
- Dockerfile and docker-compose.yml
- Development and production configurations
- Health checks and monitoring
- Best practice directory structure

## 📚 Documentation Generation

dev-stack includes a built-in documentation generation system that creates comprehensive reference docs from YAML manifests.

### Generating Documentation
```bash
# Generate all documentation from YAML manifests
dev-stack docs

# Generate only command reference
dev-stack docs --commands-only

# Generate only services guide
dev-stack docs --services-only

# Preview changes without writing files
dev-stack docs --dry-run

# Show detailed progress
dev-stack docs --verbose
```

The documentation system automatically generates:
- **Command Reference** (`docs/reference.md`) from `scripts/commands.yaml`
- **Services Guide** (`docs/services.md`) from `services/services.yaml`

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `task test`
5. Run linting: `task lint`
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

- [x] **Complete Go Migration** - Python-free implementation ✅
- [x] **Comprehensive CLI** - Full feature parity achieved ✅
- [x] **Documentation Generation** - Native Go implementation ✅
- [x] **Modern Build System** - Task-based build system with intelligent caching ✅
- [ ] Plugin system for extensibility
- [ ] Advanced version management
- [ ] Team collaboration features
- [ ] Cloud deployment integration
- [ ] Performance monitoring
- [ ] Auto-update system

---

> **Built with ❤️ by the dev-stack team**  
> Making local development environments simple, consistent, and powerful.