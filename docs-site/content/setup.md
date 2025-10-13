---
title: "Setup & Installation"
description: "Complete setup guide for dev-stack with Docker, dependencies, and initial configuration"
lead: "Get dev-stack up and running on your system with this comprehensive installation guide"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 10
toc: true
---

# Setup & Installation Guide (dev-stack)

> **Quick Checklist**
>
> - Docker installed and running
> - Sufficient RAM, disk, and CPU
> - Framework copied or linked to your project
> - Initial configuration created and edited
> - Setup script run and services verified
> - See troubleshooting below for common issues

This guide covers everything you need to get **dev-stack** up and running on your system.

> For a quick start, main configuration example, and command reference, see the [README](../README.md).
> For troubleshooting and advanced help, see [Troubleshooting Guide](troubleshooting.md).

## 📦 dev-stack CLI Installation

Choose your preferred installation method:

### Method 1: Download Binary (Recommended)

**macOS and Linux:**

```bash
# Download the latest release for your platform
curl -L -o dev-stack "https://github.com/isaacgarza/dev-stack/releases/latest/download/dev-stack-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
chmod +x dev-stack
sudo mv dev-stack /usr/local/bin/
```

**Windows (PowerShell):**

```powershell
# Download and install dev-stack for Windows
Invoke-WebRequest -Uri "https://github.com/isaacgarza/dev-stack/releases/latest/download/dev-stack-windows-amd64.exe" -OutFile "dev-stack.exe"
# Move to a directory in your PATH
```

### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/isaacgarza/dev-stack.git
cd dev-stack

# Build using Task (recommended)
task build
sudo cp build/dev-stack /usr/local/bin/

# Or build with Go directly
go build -o dev-stack ./cmd/dev-stack
sudo mv dev-stack /usr/local/bin/
```

### Method 3: Go Install

```bash
# Install directly with Go (requires Go 1.21+)
go install github.com/isaacgarza/dev-stack/cmd/dev-stack@latest
```

### Verify Installation

```bash
# Check version and basic functionality
dev-stack --version
dev-stack --help

# Run system health check
dev-stack doctor
```

### CLI Requirements

- **Go**: 1.21+ (only for building from source)
- **Docker**: Required for most dev-stack operations
- **Git**: For cloning repositories and version management

## 📋 Prerequisites

Before using this framework, you need Docker installed and running. Here are the recommended setups for different environments.

### System Requirements

- **Docker**: 20.0+ with Docker Compose 2.0+
- **RAM**: 8GB+ recommended (6GB minimum)
- **Disk**: 50GB+ available space
- **CPU**: 4+ cores recommended for multiple services

## 🐳 Docker Setup

### macOS Setup with Colima (Recommended)

[Colima](https://github.com/abiosoft/colima) is a lightweight Docker Desktop alternative for macOS that uses fewer resources and provides better performance.

```bash
# Install Colima and Docker CLI via Homebrew
brew install colima docker

# Start Colima with recommended settings for development
colima start --cpu 4 --memory 8 --disk 100

# Verify Docker is working
docker --version
docker compose version
```

**Colima Configuration for Framework:**

```bash
# For better performance with multiple services
colima start --cpu 4 --memory 8 --disk 100 --vm-type=vz --mount-type=virtiofs

# Enable Kubernetes (optional, for advanced use cases)
colima start --kubernetes --cpu 4 --memory 8
```

**Managing Colima:**

```bash
# Check status
colima status

# Stop Colima
colima stop

# Reset if needed
colima delete
colima start --cpu 4 --memory 8
```

### Docker Desktop (Alternative)

If you prefer Docker Desktop:

```bash
# Install via Homebrew
brew install --cask docker

# Or download from https://www.docker.com/products/docker-desktop
```

**Docker Desktop Configuration:**

- Go to Settings > Resources
- Set Memory to 8GB+
- Set CPU to 4+ cores
- Ensure sufficient disk space

### Linux Setup

**Ubuntu/Debian:**

```bash
# Update package index
sudo apt-get update

# Install Docker
sudo apt-get install docker.io

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group (requires logout/login)
sudo usermod -aG docker $USER
```

**CentOS/RHEL/Fedora:**

```bash
# Install Docker
sudo dnf install docker

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group
sudo usermod -aG docker $USER
```

**Verify Installation:**

```bash
# Test Docker installation
docker --version
docker compose version
docker run hello-world
```

## 🧪 IntelliJ IDEA Integration

### Docker Plugin Setup

1. Open IntelliJ IDEA
2. Go to Settings > Build, Execution, Deployment > Docker
3. Add Docker configuration:
   - **Name**: Local Docker
   - **Connect to Docker daemon with**: Docker for Mac/Colima
   - **Docker socket**: `unix:///var/run/docker.sock` (default)

### Testcontainers Configuration

Add to your `application-test.yml`:

```yaml
# Use framework services for integration tests
spring:
  datasource:
    url: jdbc:tc:postgresql:15:///test_db
    driver-class-name: org.testcontainers.jdbc.ContainerDatabaseDriver

  # Or connect to running framework services
  datasource:
    url: jdbc:postgresql://localhost:5432/local_dev
    username: postgres
    password: password

  data:
    redis:
      host: localhost
      port: 6379
      password: password

testcontainers:
  # Reuse containers across test runs
  reuse:
    enable: true
```

### Test Dependencies

Add to your `build.gradle`:

```gradle
dependencies {
    // Framework-compatible test dependencies
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
    testImplementation 'org.testcontainers:junit-jupiter'
    testImplementation 'org.testcontainers:postgresql'
    testImplementation 'org.testcontainers:kafka'
    testImplementation 'org.testcontainers:localstack'

    // Use framework services instead of embedded
    testImplementation 'redis.clients:jedis'
    testRuntimeOnly 'org.postgresql:postgresql'
}
```

### IDE Test Configuration

**IntelliJ Run Configuration VM Options:**

```bash
-Dspring.profiles.active=test
-Dtestcontainers.reuse.enable=true
-Dspring.datasource.url=jdbc:postgresql://localhost:5432/local_dev
-Dspring.data.redis.host=localhost
-Dspring.data.redis.port=6379
```

### Integration Test Strategies

**Option 1: Use Framework Services (Recommended)**

```java
@SpringBootTest
@TestPropertySource(properties = {
    "spring.datasource.url=jdbc:postgresql://localhost:5432/local_dev",
    "spring.data.redis.host=localhost"
})
class IntegrationTest {
    // Tests run against framework services
    // Start framework: dev-stack up
}
```

**Option 2: Testcontainers with Framework Images**

```java
@SpringBootTest
@Testcontainers
class ContainerizedIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:15-alpine")
            .withDatabaseName("test_db")
            .withUsername("test_user")
            .withPassword("test_password");

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
    }
}
```

### Recommended IntelliJ Plugins

Install these plugins for better Docker/framework integration:

- **Docker**: Built-in Docker support
- **Database Tools and SQL**: Connect to framework databases
- **Redis**: Redis client integration
- **Kafka**: Kafka topic browsing
- **AWS Toolkit**: LocalStack integration

## 🏗️ Framework Installation

### Option 1: Copy Framework Directory

```bash
# Copy the entire framework to your project
cp -r /path/to/dev-stack-framework /path/to/your/project/

# Make scripts executable
chmod +x /path/to/your/project/dev-stack-framework/scripts/*.sh
```

### Option 2: Git Submodule (Recommended)

```bash
# Add framework as a git submodule
cd /path/to/your/project
git submodule add <framework-repo-url> dev-stack-framework

# Initialize and update submodule
git submodule update --init --recursive
```

### Option 3: Symbolic Link

```bash
# Create symbolic link to shared framework
ln -s /shared/path/to/dev-stack-framework /path/to/your/project/dev-stack-framework
```

## 🚀 Initial Setup

See the [README](../README.md) for the main configuration example and command reference.

### 1. Initialize Configuration

```bash
dev-stack init
```

This creates a sample `dev-stack-config.yaml` file in your project root.

### 2. Edit Configuration

Edit `dev-stack-config.yaml` to customize your stack.
See the [Configuration Guide](configuration.md) for all options.

### 3. Run Setup

```bash
dev-stack up
```

This will:

- Validate your configuration
- Pull required Docker images
- Generate Docker Compose and environment files
- Start services

### 4. Verify Installation

```bash
dev-stack status             # Check service status and connection information
docker ps                      # See running containers
```

## 🔄 Multi-Repository Usage

The framework automatically detects existing instances from other repositories and provides options to:

1. **Clean up existing instances** and start fresh with your configuration
2. **Connect to existing instances** (reuse running services from another repo)
3. **Cancel setup** to avoid conflicts

### Workflow Example

**First Repository:**

```bash
cd /path/to/repo1
dev-stack up
# Services start on standard ports
```

**Second Repository (Conflict Detection):**

```bash
cd /path/to/repo2
dev-stack up

# Framework detects existing instances and may reuse them automatically
# Use dev-stack cleanup to remove existing instances if needed
```

### Automatic Options

```bash
# Cleanup existing and start fresh
dev-stack cleanup
dev-stack up

# Check system health before starting
dev-stack doctor
```

## 🐛 Common Setup Issues

### Docker Not Running

```bash
# Check Docker status
docker info

# Start Colima (macOS)
colima start

# Start Docker service (Linux)
sudo systemctl start docker
```

### Permission Denied (Linux)

```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout and login again

# Or run with sudo (temporary)
sudo dev-stack up
```

### Port Conflicts

```bash
# Find what's using the port
lsof -i :6379

# Kill the process
kill -9 PID

# Or let framework handle it
dev-stack cleanup
```

### Memory Issues

```bash
# Check available memory
free -h  # Linux
vm_stat  # macOS

# Increase Docker memory limit
# Docker Desktop: Settings > Resources > Memory
# Colima: colima start --memory 8
```

### Colima Issues (macOS)

```bash
# Reset Colima
colima stop
colima delete
colima start --cpu 4 --memory 8

# Check Colima status
colima status
colima list
```

## ✅ Verification Checklist

After setup, verify everything works:

- [ ] Docker is running: `docker info`
- [ ] Framework services start: `dev-stack status`
- [ ] Configuration is valid: No errors during setup
- [ ] Generated files exist: `docker-compose.generated.yml`, `.env.generated`
- [ ] Services are accessible: Check ports with `dev-stack status`
- [ ] IDE integration works: Database connections, Redis access
- [ ] Application connects: Spring Boot can connect to services

## 🎯 What's Next?

Now that you have dev-stack installed and configured:

1. **[Learn basic usage patterns](usage.md)** - Common workflows and daily commands
2. **[Explore available services](services.md)** - Add databases, monitoring, and more
3. **[Configure your stack](configuration.md)** - Customize settings for your needs
4. **[Integration examples](integration.md)** - Connect your applications

**Quick start:** Run `dev-stack init` to create your first project, then `dev-stack up` to start services.

**Need help?** Check the [Troubleshooting Guide](troubleshooting.md) or run `dev-stack doctor`.

## 🪝 Git Hooks Setup (Optional)

For contributors and team development, set up pre-commit and pre-push hooks to ensure code quality:

```bash
# Install Git hooks for automatic checks
task setup-hooks
```

This installs hooks that automatically run:

- **Pre-commit**: Code formatting, linting, module tidying
- **Pre-push**: All pre-commit checks + tests + build

### Manual Usage

You can also run these checks manually:

```bash
# Run pre-commit checks
task pre-commit

# Run pre-push checks
task pre-push
```

### Skip Hooks

To skip hooks temporarily (not recommended):

```bash
git commit --no-verify
git push --no-verify
```

## 🗂️ See Also

- [README](../README.md)
- [Configuration Guide](configuration.md)
- [Services Guide](services.md)
- [Usage Guide](usage.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Integration Guide](integration.md)
- [Reference](reference.md)
