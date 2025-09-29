# Setup & Installation Guide (dev-stack)

> **Quick Checklist**
> - Docker installed and running
> - Sufficient RAM, disk, and CPU
> - Framework copied or linked to your project
> - Initial configuration created and edited
> - Setup script run and services verified
> - See troubleshooting below for common issues


This guide covers everything you need to get **dev-stack** up and running on your system.

> For a quick start, main configuration example, and command reference, see the [README](../README.md).
> For troubleshooting and advanced help, see [Troubleshooting Guide](troubleshooting.md).

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
brew install colima docker docker-compose

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
sudo apt-get install docker.io docker-compose-plugin

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group (requires logout/login)
sudo usermod -aG docker $USER
```

**CentOS/RHEL/Fedora:**
```bash
# Install Docker
sudo dnf install docker docker-compose-plugin

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
    // Start framework: ./scripts/setup.sh
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
cp -r /path/to/local-dev-framework /path/to/your/project/

# Make scripts executable
chmod +x /path/to/your/project/local-dev-framework/scripts/*.sh
```

### Option 2: Git Submodule (Recommended)

```bash
# Add framework as a git submodule
cd /path/to/your/project
git submodule add <framework-repo-url> local-dev-framework

# Initialize and update submodule
git submodule update --init --recursive
```

### Option 3: Symbolic Link

```bash
# Create symbolic link to shared framework
ln -s /shared/path/to/local-dev-framework /path/to/your/project/local-dev-framework
```

## 🚀 Initial Setup

See the [README](../README.md) for the main configuration example and command reference.

### 1. Initialize Configuration

```bash
./scripts/setup.sh --init
```

This creates a sample `local-dev-config.yaml` file in your project root.

### 2. Edit Configuration

Edit `local-dev-config.yaml` to customize your stack.  
See the [Configuration Guide](configuration.md) for all options.

### 3. Run Setup

```bash
./scripts/setup.sh
```

This will:
- Validate your configuration
- Pull required Docker images
- Generate Docker Compose and environment files
- Start services

### 4. Verify Installation

```bash
./scripts/manage.sh status      # Check service status
./scripts/manage.sh info        # View connection information
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
./local-dev-framework/scripts/setup.sh
# Services start on standard ports
```

**Second Repository (Conflict Detection):**
```bash
cd /path/to/repo2
./local-dev-framework/scripts/setup.sh

# Framework detects existing instances and prompts:
# 1) Clean up existing instances and start fresh
# 2) Connect to existing instances (reuse repo1's services) 
# 3) Cancel setup

# Choose option 2 to reuse services, or 1 to start fresh
```

### Automatic Options

```bash
# Auto-cleanup existing and start fresh
./local-dev-framework/scripts/setup.sh --cleanup-existing

# Auto-connect to existing instances
./local-dev-framework/scripts/setup.sh --connect-existing  

# Force cleanup without prompts
./local-dev-framework/scripts/setup.sh --force
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
sudo ./local-dev-framework/scripts/setup.sh
```

### Port Conflicts

```bash
# Find what's using the port
lsof -i :6379

# Kill the process
kill -9 PID

# Or let framework handle it
./local-dev-framework/scripts/setup.sh --cleanup-existing
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
- [ ] Framework services start: `./scripts/manage.sh status`
- [ ] Configuration is valid: No errors during setup
- [ ] Generated files exist: `docker-compose.generated.yml`, `.env.generated`
- [ ] Services are accessible: Check ports with `./scripts/manage.sh info`
- [ ] IDE integration works: Database connections, Redis access
- [ ] Application connects: Spring Boot can connect to services

## 🗂️ See Also

- [README](../README.md)
- [Configuration Guide](configuration.md)
- [Services Guide](services.md)
- [Usage Guide](usage.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Integration Guide](integration.md)
- [Reference](reference.md)