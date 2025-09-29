# Usage & Management Guide

This guide covers daily usage patterns, service management commands, and common workflows for the Local Development Framework.

## 📋 Overview

The framework provides two main scripts for different purposes:
- **`setup.sh`**: Initial configuration and environment setup
- **`manage.sh`**: Ongoing service management and operations

## 🚀 Daily Workflow

### Morning Startup

```bash
# 1. Start Docker (if not running)
docker info                              # Check Docker status
colima start                            # Start Colima (macOS)

# 2. Start development services
cd /path/to/your/project
./local-dev-framework/scripts/setup.sh  # Uses existing config
# or
./local-dev-framework/scripts/manage.sh start  # Start existing services

# 3. Verify services are running
./local-dev-framework/scripts/manage.sh info

# 4. Start your application
./gradlew bootRun
```

### During Development

```bash
# View service information
./scripts/manage.sh info                 # Connection details
./scripts/manage.sh status              # Service health

# Debug issues
./scripts/manage.sh logs postgres -f    # Follow PostgreSQL logs
./scripts/manage.sh logs redis          # View Redis logs
./scripts/manage.sh monitor            # Real-time resource usage

# Interact with services
./scripts/manage.sh connect postgres    # PostgreSQL CLI
./scripts/manage.sh connect redis       # Redis CLI
./scripts/manage.sh connect mysql       # MySQL CLI

# Backup data before changes
./scripts/manage.sh backup postgres
```

### End of Day

```bash
# Stop services (keeps data)
./scripts/manage.sh stop

# Or cleanup everything (removes data)
./scripts/manage.sh cleanup

# Stop Docker (optional)
colima stop                             # macOS with Colima
```

## 🛠 Setup Commands

### Initial Setup

```bash
# Create sample configuration
./scripts/setup.sh --init

# Interactive configuration wizard
./scripts/setup.sh --interactive

# List available services
./scripts/setup.sh --list-services
```

### Configuration Options

```bash
# Use default configuration file
./scripts/setup.sh

# Override services from command line
./scripts/setup.sh --services=redis,postgres,jaeger

# Override project name
./scripts/setup.sh --project=my-custom-name

# Use custom configuration file
./scripts/setup.sh --config=local-dev-config.test.yaml

# Preview changes without executing
./scripts/setup.sh --dry-run

# Skip validation warnings
./scripts/setup.sh --skip-validation --force
```

### Instance Management

```bash
# Handle existing instances automatically
./scripts/setup.sh --cleanup-existing   # Remove existing, start fresh
./scripts/setup.sh --connect-existing   # Connect to existing services
./scripts/setup.sh --force             # Force cleanup without prompts

# System-wide instance management
./scripts/setup.sh --list-all          # List all framework instances
./scripts/setup.sh --cleanup-all       # Remove all instances (all repos)
```

### Advanced Setup Options

```bash
# Enable debug logging
./scripts/setup.sh --debug

# Validate configuration only
./scripts/setup.sh --validate-only

# Pull latest Docker images
./scripts/setup.sh --pull-images

# Generate files without starting services
./scripts/setup.sh --generate-only
```

## 🎛 Management Commands

### Service Control

```bash
# Start all services
./scripts/manage.sh start

# Stop all services (keeps data)
./scripts/manage.sh stop

# Restart all services
./scripts/manage.sh restart

# Start specific service
./scripts/manage.sh start redis

# Stop specific service
./scripts/manage.sh stop postgres
```

### Service Information

```bash
# Show all service connection information
./scripts/manage.sh info

# Show service status
./scripts/manage.sh status

# Show detailed service information
./scripts/manage.sh info --detailed

# Show only specific service info
./scripts/manage.sh info redis
```

### Logging and Monitoring

```bash
# View all service logs
./scripts/manage.sh logs

# Follow logs in real-time
./scripts/manage.sh logs -f

# View specific service logs
./scripts/manage.sh logs postgres
./scripts/manage.sh logs redis --tail=100
./scripts/manage.sh logs jaeger --since=1h

# Monitor resource usage
./scripts/manage.sh monitor

# Show Docker container stats
./scripts/manage.sh stats
```

### Service Interaction

```bash
# Connect to service CLI
./scripts/manage.sh connect postgres    # PostgreSQL psql
./scripts/manage.sh connect redis       # Redis CLI
./scripts/manage.sh connect mysql       # MySQL CLI

# Execute commands in containers
./scripts/manage.sh exec postgres psql -U postgres -c "SELECT version();"
./scripts/manage.sh exec redis redis-cli INFO
./scripts/manage.sh exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Copy files to/from containers
./scripts/manage.sh cp backup.sql postgres:/tmp/
./scripts/manage.sh cp postgres:/tmp/export.sql ./
```

### Data Management

```bash
# Backup databases
./scripts/manage.sh backup postgres
./scripts/manage.sh backup mysql
./scripts/manage.sh backup postgres --output=custom-backup.sql

# Restore databases
./scripts/manage.sh restore postgres backup.sql
./scripts/manage.sh restore mysql backup.sql

# Clear Redis cache
./scripts/manage.sh exec redis redis-cli FLUSHALL

# Reset Kafka topics
./scripts/manage.sh exec kafka kafka-topics --bootstrap-server localhost:9092 --delete --topic my-topic
```

### Maintenance

```bash
# Update service images and recreate containers
./scripts/manage.sh update

# Clean up unused Docker resources
./scripts/manage.sh cleanup-docker

# Remove all framework resources (destructive!)
./scripts/manage.sh cleanup

# System-wide cleanup (all repositories)
./scripts/manage.sh cleanup-all
```

## 📊 Multi-Repository Workflows

### Scenario 1: Shared Services

When working on multiple related projects that can share services:

```bash
# In first repository
cd /path/to/repo1
./scripts/setup.sh
# Services start on standard ports

# In second repository - reuse services
cd /path/to/repo2  
./scripts/setup.sh --connect-existing
# Connects to repo1's running services
```

### Scenario 2: Isolated Testing

When you need isolated environments for testing:

```bash
# In second repository - clean isolation
cd /path/to/repo2
./scripts/setup.sh --cleanup-existing
# Stops repo1 services, starts fresh for repo2
```

### Scenario 3: Resource Management

Monitor and manage all instances across repositories:

```bash
# List all running instances
./scripts/manage.sh list-all

# Output example:
# Repository: /path/to/repo1
#   Project: api-service-dev
#   Services: redis, postgres, jaeger
#   Status: running
#
# Repository: /path/to/repo2  
#   Project: user-service-dev
#   Services: redis, mysql, kafka
#   Status: running

# Clean up everything at end of day
./scripts/manage.sh cleanup-all
```

## 🔧 Configuration Management

### Runtime Configuration Changes

```bash
# Modify configuration
vim local-dev-config.yaml

# Apply changes (recreates affected services)
./scripts/setup.sh

# Apply changes to specific service only
./scripts/setup.sh --services=postgres
```

### Environment-Specific Configurations

```bash
# Create different configurations
cp local-dev-config.yaml local-dev-config.test.yaml
cp local-dev-config.yaml local-dev-config.integration.yaml

# Use specific configuration
./scripts/setup.sh --config=local-dev-config.test.yaml

# Switch between configurations
./scripts/manage.sh stop
./scripts/setup.sh --config=local-dev-config.integration.yaml
```

### Configuration Validation

```bash
# Validate configuration syntax
./scripts/setup.sh --validate-only

# Check for warnings
./scripts/setup.sh --dry-run

# Validate and show resolved configuration
./scripts/setup.sh --debug --dry-run
```

## 🧪 Testing Workflows

### Integration Testing

```bash
# Start services for testing
./scripts/setup.sh --config=local-dev-config.test.yaml

# Run integration tests
./gradlew integrationTest

# Clean up after tests
./scripts/manage.sh cleanup
```

### CI/CD Integration

```bash
# Lightweight CI configuration
./scripts/setup.sh --services=postgres,redis --force

# Run tests
./gradlew test integrationTest

# Cleanup in CI
./scripts/manage.sh cleanup
```

### Database Testing

```bash
# Backup production-like data
./scripts/manage.sh backup postgres

# Reset to clean state
./scripts/manage.sh exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS my_app_dev;"
./scripts/manage.sh exec postgres psql -U postgres -c "CREATE DATABASE my_app_dev;"

# Load test data
./scripts/manage.sh exec postgres psql -U postgres -d my_app_dev -f /path/to/test-data.sql

# Run tests
./gradlew test

# Restore backup if needed
./scripts/manage.sh restore postgres backup.sql
```

## 🔍 Debugging and Troubleshooting

### Service Health Checks

```bash
# Check all services
./scripts/manage.sh status

# Detailed health check
./scripts/manage.sh info --health

# Service-specific checks
curl http://localhost:16686/         # Jaeger UI
curl http://localhost:9090/          # Prometheus UI  
curl http://localhost:4566/health    # LocalStack health
redis-cli -h localhost -p 6379 ping  # Redis
```

### Log Analysis

```bash
# View recent errors
./scripts/manage.sh logs --since=1h | grep -i error

# Follow logs for specific service
./scripts/manage.sh logs postgres -f

# Export logs for analysis
./scripts/manage.sh logs > framework-logs.txt

# View container resource usage
./scripts/manage.sh monitor
```

### Network Debugging

```bash
# Test port connectivity
telnet localhost 5432               # PostgreSQL
telnet localhost 6379               # Redis
telnet localhost 9092               # Kafka

# Check port usage
lsof -i :5432
netstat -tulpn | grep :6379

# Inspect Docker networks
docker network ls
docker network inspect local-dev-framework_default
```

### Performance Debugging

```bash
# Monitor resource usage
./scripts/manage.sh monitor

# Check Docker system usage
docker system df

# Analyze container performance  
docker stats

# Memory usage by service
./scripts/manage.sh exec postgres ps aux
./scripts/manage.sh exec redis redis-cli INFO memory
```

## 📈 Performance Optimization

### Resource Tuning

```bash
# Monitor resource usage
./scripts/manage.sh monitor

# Adjust memory limits in configuration
vim local-dev-config.yaml
# overrides:
#   postgres:
#     memory_limit: "1g"
#   redis:
#     memory_limit: "512m"

# Apply changes
./scripts/setup.sh
```

### Service Optimization

```bash
# Disable unnecessary services
vim local-dev-config.yaml
# services:
#   enabled:
#     - redis
#     - postgres
#     # - jaeger     # Disable if not needed
#     # - prometheus # Disable if not needed

# Use faster startup options
# validation:
#   skip_warnings: true
#   auto_start: true
```

### Development Speed Tips

```bash
# Skip image pulls for faster startup
./scripts/setup.sh --no-pull

# Use cached data between sessions
# Don't run cleanup, just stop services
./scripts/manage.sh stop

# Prestart services in background
./scripts/manage.sh start &
# Continue with other work while services start
```

## 🔄 Update and Maintenance

### Framework Updates

```bash
# Update framework (if using git submodule)
git submodule update --remote local-dev-framework

# Update Docker images
./scripts/manage.sh update

# Recreate services with new images
./scripts/setup.sh --force
```

### Data Maintenance

```bash
# Regular database backups
./scripts/manage.sh backup postgres
./scripts/manage.sh backup mysql

# Clean up old Docker resources
./scripts/manage.sh cleanup-docker

# Rotate logs (if needed)
docker system prune
```

### Health Monitoring

```bash
# Daily health check
./scripts/manage.sh status

# Weekly resource cleanup
./scripts/manage.sh cleanup-docker

# Monitor disk usage
docker system df
df -h
```

## 📚 Integration Examples

### Spring Boot Application

```java
// Application startup check
@Component
public class ServiceHealthChecker {
    
    @EventListener(ApplicationReadyEvent.class)
    public void checkServices() {
        // Verify Redis connection
        try {
            redisTemplate.opsForValue().set("health-check", "ok");
            log.info("Redis connection: OK");
        } catch (Exception e) {
            log.error("Redis connection failed", e);
        }
        
        // Verify database connection
        try {
            jdbcTemplate.queryForObject("SELECT 1", Integer.class);
            log.info("Database connection: OK");
        } catch (Exception e) {
            log.error("Database connection failed", e);
        }
    }
}
```

### Development Profile

```yaml
# application-local.yml (generated by framework)
spring:
  profiles:
    active: local
  datasource:
    url: jdbc:postgresql://localhost:5432/my_app_dev
    username: app_user
    password: dev-password
  data:
    redis:
      host: localhost
      port: 6379
      password: dev-password

logging:
  level:
    org.springframework.web: DEBUG
    org.hibernate.SQL: DEBUG
```

## 🆘 Getting Help

### Self-Help Commands

```bash
# Show help for setup script
./scripts/setup.sh --help

# Show help for management script  
./scripts/manage.sh --help

# List available services
./scripts/setup.sh --list-services

# Validate current configuration
./scripts/setup.sh --validate-only
```

### Common Commands Reference

```bash
# Quick reference card
./scripts/manage.sh --help
./scripts/setup.sh --help

# Service information
./scripts/manage.sh info
./scripts/manage.sh status

# Troubleshooting
./scripts/manage.sh logs
./scripts/manage.sh monitor
```

## 📚 Next Steps

- **[Configuration Guide](configuration.md)** - Advanced configuration options
- **[Services Guide](services.md)** - Detailed service information  
- **[Integration Guide](integration.md)** - Application integration patterns
- **[Troubleshooting](troubleshooting.md)** - Detailed problem solving
- **[Quick Reference](reference.md)** - Commands cheatsheet