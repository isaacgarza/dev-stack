---
title: "Getting Started"
description: "Installation and setup guide for dev-stack"
weight: 10
---

# Getting Started with dev-stack

This guide will help you install and set up dev-stack on your system.

## Prerequisites

Before installing dev-stack, ensure you have:

- **Docker**: Required for container management
- **Git**: For version control and project initialization
- **Terminal/Command Line**: For running dev-stack commands

### System Requirements

- **Operating System**: Linux, macOS, or Windows (with WSL2)
- **Memory**: Minimum 4GB RAM (8GB recommended)
- **Disk Space**: At least 2GB free space
- **Network**: Internet connection for downloading images and dependencies

## Installation

### Option 1: Download Binary (Recommended)

Download the pre-compiled binary for your platform:

```bash
# Detect your platform and download
curl -L -o dev-stack "https://github.com/isaacgarza/dev-stack/releases/latest/download/dev-stack-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"

# Make it executable
chmod +x dev-stack

# Move to PATH
sudo mv dev-stack /usr/local/bin/

# Verify installation
dev-stack version
```

### Option 2: Build from Source

If you have Go installed and want the latest features:

```bash
# Clone the repository
git clone https://github.com/isaacgarza/dev-stack.git
cd dev-stack

# Build the binary
task build

# Install to PATH
sudo cp build/dev-stack /usr/local/bin/

# Verify installation
dev-stack version
```

### Option 3: Go Install

For Go developers:

```bash
# Install directly with Go
go install github.com/isaacgarza/dev-stack/cmd/dev-stack@latest

# Ensure $GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Verify installation
dev-stack version
```

## Initial Setup

### 1. Verify Docker Installation

dev-stack requires Docker to be running:

```bash
# Check Docker status
docker --version
docker ps

# If Docker isn't running, start it
# On macOS/Windows: Start Docker Desktop
# On Linux: sudo systemctl start docker
```

### 2. Run System Health Check

Check if your system is ready for dev-stack:

```bash
dev-stack doctor
```

This command will:
- ✅ Verify Docker is running
- ✅ Check available disk space
- ✅ Test network connectivity
- ✅ Validate required tools

### 3. Create Your First Project

Initialize a new project to test your setup:

```bash
# Create a new directory
mkdir my-first-project
cd my-first-project

# Initialize with Go template
dev-stack init go --name my-app

# Start the development stack
dev-stack up

# Check status
dev-stack status
```

## Configuration

### Global Configuration

dev-stack stores its global configuration in:

- **Linux/macOS**: `~/.config/dev-stack/config.yaml`
- **Windows**: `%APPDATA%\dev-stack\config.yaml`

### Project Configuration

Each project has its own configuration file:

- `dev-stack-config.yaml` in the project root

### Environment Variables

You can override configuration with environment variables:

```bash
# Set default project template
export DEV_STACK_DEFAULT_TEMPLATE=go

# Set custom Docker registry
export DEV_STACK_REGISTRY=myregistry.com

# Enable debug logging
export DEV_STACK_DEBUG=true
```

## Basic Commands

Here are the essential commands to get you started:

### Project Management

```bash
# Initialize a new project
dev-stack init [template] --name [project-name]

# Start services
dev-stack up

# Stop services
dev-stack down

# Restart services
dev-stack restart

# Check status
dev-stack status
```

### Service Management

```bash
# List available services
dev-stack service list

# Add a service to your project
dev-stack service add [service-name]

# Remove a service
dev-stack service remove [service-name]

# View service logs
dev-stack logs [service-name]
```

### Health and Diagnostics

```bash
# Run health checks
dev-stack doctor

# Show system information
dev-stack info

# View configuration
dev-stack config show
```

## Project Templates

dev-stack comes with several built-in templates:

### Go Template
```bash
dev-stack init go --name my-go-app
```
- Hot reload with Air
- Go modules support
- Health check endpoints
- Docker-ready

### Node.js Template
```bash
dev-stack init node --name my-node-app
```
- npm/yarn support
- Nodemon for hot reload
- TypeScript ready
- Express.js boilerplate

### Python Template
```bash
dev-stack init python --name my-python-app
```
- Virtual environment setup
- Flask/FastAPI support
- Hot reload with watchdog
- Requirements management

### Full-Stack Template
```bash
dev-stack init fullstack --name my-fullstack-app
```
- Frontend and backend separation
- API gateway configuration
- Database integration
- Authentication setup

## Troubleshooting

### Common Issues

**"dev-stack: command not found"**
```bash
# Check if binary is in PATH
which dev-stack

# If not found, add to PATH or reinstall
export PATH=$PATH:/usr/local/bin
```

**"Docker daemon not running"**
```bash
# Start Docker
sudo systemctl start docker  # Linux
# Or start Docker Desktop on macOS/Windows
```

**"Permission denied" errors**
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

**Port conflicts**
```bash
# Check what's using the port
lsof -i :8080

# Stop conflicting service or change port
dev-stack config set default-port 8081
```

### Getting Help

```bash
# Show help for any command
dev-stack help [command]

# Show version and build info
dev-stack version

# Enable verbose logging
dev-stack --verbose [command]
```

## Next Steps

Now that you have dev-stack installed and working:

1. **[Learn the basic usage patterns]({{< ref "/usage" >}})**
2. **[Explore the CLI reference]({{< ref "/cli-reference" >}})**
3. **[Discover available services]({{< ref "/services" >}})**
4. **[Join the community](https://github.com/isaacgarza/dev-stack/discussions)**

---

> **Pro Tip**: Use `dev-stack doctor` regularly to ensure your development environment stays healthy!