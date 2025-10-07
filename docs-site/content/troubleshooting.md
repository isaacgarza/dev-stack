---
title: "Troubleshooting"
description: "Common issues and solutions for dev-stack problems"
lead: "Quick solutions to the most common dev-stack issues and problems"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 70
toc: true

---

# Troubleshooting Guide

## 🚨 Top 5 Issues (Quick Reference)

1. **Docker Not Running**
   - Run `docker info` to check status.
   - Start Docker/Colima if not running.

2. **Port Conflicts**
   - Run `lsof -i :PORT_NUMBER` to find conflicts.
   - Use `--cleanup-existing` to resolve.

3. **Service Won't Start**
   - Check logs: `dev-stack logs SERVICE_NAME`
   - Restart service: `dev-stack down && dev-stack up`

4. **Memory Issues**
   - Check usage: `docker stats`
   - Increase Docker memory limit.

5. **Invalid Configuration**
   - Validate YAML: `dev-stack doctor`
   - Run setup with debug: `dev-stack --verbose up`

---

This guide covers common issues, debugging techniques, and solutions for the Local Development Framework.

## 📋 Overview

Most issues with the framework fall into these categories:
- Docker and container issues
- Service connectivity problems
- Configuration errors
- Resource constraints
- Port conflicts

## 🚨 Quick Diagnosis

### Health Check Commands

```bash
# Quick system check
docker info                              # Docker daemon status
dev-stack status                       # Service status and connection information
dev-stack doctor                       # Comprehensive system health check

# Resource check
docker system df                        # Docker disk usage
free -h                                 # Available memory (Linux)
vm_stat                                 # Memory stats (macOS)

# Network check
lsof -i :5432                          # PostgreSQL port
lsof -i :6379                          # Redis port
lsof -i :9092                          # Kafka port
```

### Log Analysis

```bash
# View all service logs
dev-stack logs

# Recent errors only
dev-stack logs --since=1h | grep -i error

# Service-specific logs
dev-stack logs postgres -f
dev-stack logs redis --tail=50
```

## 🐳 Docker Issues

### Docker Not Running

**Symptoms:**
- "Cannot connect to the Docker daemon" error
- `docker info` fails

**Solutions:**

**macOS with Colima:**
```bash
# Check Colima status
colima status

# Start Colima
colima start --cpu 4 --memory 8

# If stuck, reset Colima
colima stop
colima delete
colima start --cpu 4 --memory 8 --vm-type=vz --mount-type=virtiofs
```

**macOS with Docker Desktop:**
```bash
# Restart Docker Desktop through the application
# Or via command line:
killall Docker && open /Applications/Docker.app
```

**Linux:**
```bash
# Check Docker service
sudo systemctl status docker

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# If permission denied
sudo usermod -aG docker $USER
# Logout and login again
```

### Docker Out of Space

**Symptoms:**
- "No space left on device" errors
- Container creation fails

**Solutions:**
```bash
# Check Docker disk usage
docker system df

# Clean up unused resources
docker system prune -a

# Remove unused volumes
docker volume prune

# Remove unused networks
docker network prune

# Clean up framework resources specifically
dev-stack cleanup
```

### Docker Memory Issues

**Symptoms:**
- Services crash randomly
- Slow performance
- "Cannot allocate memory" errors

**Solutions:**
```bash
# Check current memory usage
docker stats

# Increase Docker memory limit
# Docker Desktop: Settings > Resources > Memory (8GB+)
# Colima: colima start --memory 8

# Reduce service memory usage in config
vim dev-stack-config.yaml
# overrides:
#   postgres:
#     memory_limit: "256m"
#   redis:
#     memory_limit: "128m"
```

## 🔌 Service Connectivity Issues

### Cannot Connect to Database

**Symptoms:**
- Connection refused errors
- Timeout connecting to PostgreSQL/MySQL
- Application startup fails

**Diagnosis:**
```bash
# Check if service is running
dev-stack status

# Check if port is open
telnet localhost 5432               # PostgreSQL
telnet localhost 3306               # MySQL

# Check service logs
dev-stack logs postgres
dev-stack logs mysql

# Test connection directly
psql -h localhost -U postgres
mysql -h localhost -u root -p
```

**Solutions:**
```bash
# Restart database service
dev-stack down
dev-stack up

# Check for port conflicts
lsof -i :5432
# Kill conflicting process if found
kill -9 PID

# Verify configuration
dev-stack status

# Recreate service
dev-stack down
dev-stack up
```

### Redis Connection Issues

**Symptoms:**
- "Connection refused" to Redis
- Authentication failures
- Timeout errors

**Diagnosis:**
```bash
# Test Redis connection
redis-cli -h localhost -p 6379 ping

# With password
redis-cli -h localhost -p 6379 -a your-password ping

# Check Redis logs
dev-stack logs redis

# Check Redis info
dev-stack exec redis redis-cli INFO
```

**Solutions:**
```bash
# Restart Redis
./scripts/manage.sh restart redis

# Clear Redis data if corrupted
./scripts/manage.sh exec redis redis-cli FLUSHALL

# Check Redis configuration
./scripts/manage.sh exec redis redis-cli CONFIG GET "*"

# Verify password in configuration
./scripts/manage.sh info redis
```

### Kafka Connection Issues

**Symptoms:**
- Cannot connect to Kafka broker
- Topic creation fails
- Consumer/producer errors

**Diagnosis:**
```bash
# Check Kafka status
./scripts/manage.sh logs kafka
./scripts/manage.sh logs zookeeper

# Test Kafka connection
./scripts/manage.sh exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Check Kafka UI
curl http://localhost:8080
```

**Solutions:**
```bash
# Restart Kafka stack
./scripts/manage.sh restart kafka

# Clear Kafka data if needed
./scripts/manage.sh stop kafka
docker volume rm $(docker volume ls -q | grep kafka)
./scripts/setup.sh --services=kafka

# Check Zookeeper connectivity
./scripts/manage.sh exec zookeeper zkCli.sh -server localhost:2181
```

## 🌐 Network and Port Issues

### Port Already in Use

**Symptoms:**
- "Port is already allocated" errors
- "Address already in use" errors
- Services fail to start

**Diagnosis:**
```bash
# Find what's using the port
lsof -i :5432                          # PostgreSQL
lsof -i :6379                          # Redis
lsof -i :9092                          # Kafka
netstat -tulpn | grep :PORT

# Check for other framework instances
./scripts/manage.sh list-all
```

**Solutions:**
```bash
# Kill process using the port
kill -9 PID

# Use different ports in configuration
vim dev-stack-config.yaml
# overrides:
#   postgres:
#     port: 5433
#   redis:
#     port: 6380

# Let framework handle conflicts automatically
./scripts/setup.sh --cleanup-existing
./scripts/setup.sh --force
```

### DNS Resolution Issues

**Symptoms:**
- Cannot resolve service hostnames
- "Name or service not known" errors

**Solutions:**
```bash
# Use localhost instead of service names
# In application configuration:
# spring.datasource.url=jdbc:postgresql://localhost:5432/db

# Check Docker network
docker network ls
docker network inspect dev-stack-framework_default

# Recreate network
./scripts/manage.sh cleanup
./scripts/setup.sh
```

## ⚙️ Configuration Issues

### Invalid Configuration File

**Symptoms:**
- YAML parsing errors
- "Configuration file not found"
- Setup script fails with validation errors

**Diagnosis:**
```bash
# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('dev-stack-config.yaml'))"

# Check configuration with framework
./scripts/setup.sh --validate-only

# Show resolved configuration
./scripts/setup.sh --debug --dry-run
```

**Solutions:**
```bash
# Create new configuration from sample
./scripts/setup.sh --init --force

# Fix YAML syntax errors
vim dev-stack-config.yaml
# Common issues:
# - Incorrect indentation (use spaces, not tabs)
# - Missing quotes around strings with special characters
# - Incorrect list syntax

# Use online YAML validator
# Copy content to https://yaml-online-parser.appspot.com/
```

### Service Configuration Errors

**Symptoms:**
- Services start but behave incorrectly
- Authentication failures
- Wrong database/cache settings

**Solutions:**
```bash
# Check generated configuration
cat docker-compose.generated.yml
cat .env.generated
cat application-local.yml.generated

# Compare with sample configuration
diff dev-stack-config.yaml dev-stack-framework/dev-stack-config.sample.yaml

# Reset to defaults
./scripts/setup.sh --init
# Manually merge your changes
```

## 🔄 Service-Specific Issues

### PostgreSQL Issues

**Connection refused:**
```bash
# Check if PostgreSQL is ready
./scripts/manage.sh exec postgres pg_isready -h localhost

# Check PostgreSQL logs
./scripts/manage.sh logs postgres

# Reset PostgreSQL data
./scripts/manage.sh stop postgres
docker volume rm $(docker volume ls -q | grep postgres)
./scripts/setup.sh --services=postgres
```

**Database doesn't exist:**
```bash
# Create database manually
./scripts/manage.sh exec postgres createdb -U postgres my_app_dev

# Or recreate with correct configuration
vim dev-stack-config.yaml
# overrides:
#   postgres:
#     database: "my_app_dev"
./scripts/setup.sh --services=postgres
```

**Permission denied:**
```bash
# Check user and permissions
./scripts/manage.sh exec postgres psql -U postgres -c "\du"

# Create user if missing
./scripts/manage.sh exec postgres psql -U postgres -c "CREATE USER app_user WITH PASSWORD 'password';"
./scripts/manage.sh exec postgres psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE my_app_dev TO app_user;"
```

### Redis Issues

**Memory issues:**
```bash
# Check Redis memory usage
./scripts/manage.sh exec redis redis-cli INFO memory

# Clear Redis data
./scripts/manage.sh exec redis redis-cli FLUSHALL

# Increase memory limit
vim dev-stack-config.yaml
# overrides:
#   redis:
#     memory_limit: "512m"
```

**Persistence issues:**
```bash
# Check Redis persistence
./scripts/manage.sh exec redis redis-cli LASTSAVE

# Disable persistence for development
vim dev-stack-config.yaml
# overrides:
#   redis:
#     config: |
#       save ""
```

### LocalStack Issues

**Services not available:**
```bash
# Check LocalStack logs
./scripts/manage.sh logs localstack

# Check LocalStack health
curl http://localhost:4566/health

# Restart LocalStack
./scripts/manage.sh restart localstack

# Check enabled services
curl http://localhost:4566/_localstack/health | jq
```

**SQS/SNS issues:**
```bash
# List SQS queues
aws --endpoint-url=http://localhost:4566 sqs list-queues

# List SNS topics
aws --endpoint-url=http://localhost:4566 sns list-topics

# Recreate queues/topics
./scripts/manage.sh stop localstack
./scripts/setup.sh --services=localstack
```

**DynamoDB issues:**
```bash
# List DynamoDB tables
aws --endpoint-url=http://localhost:4566 dynamodb list-tables

# Check table status
aws --endpoint-url=http://localhost:4566 dynamodb describe-table --table-name my-table

# Recreate tables
./scripts/manage.sh exec localstack awslocal dynamodb delete-table --table-name my-table
./scripts/setup.sh --services=localstack
```

## 🚀 Performance Issues

### Slow Service Startup

**Symptoms:**
- Services take a long time to start
- Timeouts during startup
- Application fails to connect initially

**Solutions:**
```bash
# Check resource usage
./scripts/manage.sh monitor

# Reduce memory limits for faster startup
vim dev-stack-config.yaml
# overrides:
#   global:
#     memory_limit: "256m"

# Disable unnecessary services
# services:
#   enabled:
#     - redis
#     - postgres
#     # - kafka      # Comment out if not needed
#     # - localstack # Comment out if not needed

# Pre-pull images
docker pull postgres:15-alpine
docker pull redis:7-alpine
```

### High Memory Usage

**Symptoms:**
- System becomes slow
- Out of memory errors
- Services crash randomly

**Solutions:**
```bash
# Monitor memory usage
docker stats
./scripts/manage.sh monitor

# Reduce service memory limits
vim dev-stack-config.yaml
# overrides:
#   postgres:
#     memory_limit: "256m"
#   redis:
#     memory_limit: "128m"
#   kafka:
#     memory_limit: "512m"

# Increase system memory allocation
# Docker Desktop: Settings > Resources > Memory
# Colima: colima start --memory 8
```

### Slow Database Performance

**Symptoms:**
- Long query execution times
- Application timeouts
- High database CPU usage

**Solutions:**
```bash
# Check PostgreSQL performance
./scripts/manage.sh exec postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"

# Optimize PostgreSQL for development
vim dev-stack-config.yaml
# overrides:
#   postgres:
#     config: |
#       shared_buffers = 256MB
#       effective_cache_size = 1GB
#       work_mem = 4MB
#       maintenance_work_mem = 64MB
#       wal_buffers = 16MB
#       checkpoint_completion_target = 0.9
#       random_page_cost = 1.1

# For development only (data safety disabled):
#       fsync = off
#       synchronous_commit = off
#       full_page_writes = off
```

## 🔍 Advanced Debugging

### Container Inspection

```bash
# Inspect running containers
docker ps
docker inspect CONTAINER_ID

# Check container resource usage
docker stats CONTAINER_NAME

# Execute shell in container
./scripts/manage.sh exec postgres bash
./scripts/manage.sh exec redis sh

# Check container logs with timestamps
docker logs --timestamps CONTAINER_NAME
```

### Network Debugging

```bash
# Check Docker networks
docker network ls
docker network inspect dev-stack-framework_default

# Test network connectivity between containers
./scripts/manage.sh exec app-container ping postgres
./scripts/manage.sh exec app-container telnet redis 6379

# Check DNS resolution
./scripts/manage.sh exec app-container nslookup postgres
```

### File System Debugging

```bash
# Check Docker volumes
docker volume ls
docker volume inspect VOLUME_NAME

# Check file permissions
./scripts/manage.sh exec postgres ls -la /var/lib/postgresql/data

# Copy files for debugging
./scripts/manage.sh cp postgres:/var/log/postgresql/ ./postgres-logs/
```

## 🔧 Advanced Solutions

### Complete Reset

When all else fails, perform a complete reset:

```bash
# Stop all services
./scripts/manage.sh stop

# Remove all framework resources
./scripts/manage.sh cleanup

# Clean Docker system
docker system prune -a
docker volume prune

# Remove configuration and start fresh
rm dev-stack-config.yaml
rm docker-compose.generated.yml
rm .env.generated
rm application-local.yml.generated

# Initialize new configuration
./scripts/setup.sh --init
vim dev-stack-config.yaml
./scripts/setup.sh
```

### Framework Recovery

If the framework itself is corrupted:

```bash
# Update framework (if using git submodule)
git submodule update --remote dev-stack-framework

# Or re-copy framework files
rm -rf dev-stack-framework
cp -r /path/to/fresh/dev-stack-framework ./

# Make scripts executable
chmod +x dev-stack-framework/scripts/*.sh

# Regenerate configuration
scripts/setup.sh --init
```

### System Resource Recovery

```bash
# Free up system resources
docker system prune -a --volumes

# Clear system caches (Linux)
sudo sync && echo 3 | sudo tee /proc/sys/vm/drop_caches

# Restart Docker daemon (Linux)
sudo systemctl restart docker

# Reset Colima completely (macOS)
colima stop
colima delete
rm -rf ~/.colima
colima start --cpu 4 --memory 8
```

## 📊 Monitoring and Prevention

### Health Monitoring

Create a health check script:

```bash
#!/bin/bash
# health-check.sh

echo "=== Framework Health Check ==="
echo "Docker Status:"
docker info > /dev/null 2>&1 && echo "✓ Docker running" || echo "✗ Docker not running"

echo "Services Status:"
./scripts/manage.sh status

echo "Resource Usage:"
./scripts/manage.sh monitor | head -10

echo "Disk Usage:"
docker system df

echo "=== End Health Check ==="
```

### Preventive Maintenance

Weekly maintenance routine:

```bash
#!/bin/bash
# weekly-maintenance.sh

# Backup databases
./scripts/manage.sh backup postgres
./scripts/manage.sh backup mysql

# Clean up Docker resources
docker system prune

# Update service images
./scripts/manage.sh update

# Validate configuration
./scripts/setup.sh --validate-only

echo "Maintenance complete"
```

## 📞 Getting Help

### Self-Help Checklist

Before seeking help, try these steps:

1. **Check the basics:**
   - [ ] Docker is running: `docker info`
   - [ ] Services are running: `./scripts/manage.sh status`
   - [ ] No port conflicts: `lsof -i :5432 :6379 :9092`
   - [ ] Sufficient resources: `docker stats`

2. **Review logs:**
   - [ ] Framework logs: `./scripts/manage.sh logs`
   - [ ] Service-specific logs: `./scripts/manage.sh logs SERVICE_NAME`
   - [ ] System logs: `dmesg | tail` (Linux)

3. **Validate configuration:**
   - [ ] YAML syntax: `python -c "import yaml; yaml.safe_load(open('dev-stack-config.yaml'))"`
   - [ ] Framework validation: `./scripts/setup.sh --validate-only`

4. **Try simple fixes:**
   - [ ] Restart services: `./scripts/manage.sh restart`
   - [ ] Recreate services: `./scripts/setup.sh --force`
   - [ ] Clear caches: `./scripts/manage.sh cleanup-docker`

### Information to Collect

When reporting issues, include:

```bash
# System information
uname -a
docker --version
docker compose version

# Framework status
./scripts/manage.sh info
./scripts/manage.sh status

# Configuration
cat dev-stack-config.yaml

# Recent logs
./scripts/manage.sh logs --since=1h

# Resource usage
docker stats --no-stream
docker system df
```

### Debug Mode

Enable debug mode for detailed information:

```bash
# Debug setup
./scripts/setup.sh --debug

# Debug with dry run
./scripts/setup.sh --debug --dry-run

# Verbose logging
export DEBUG=1
./scripts/setup.sh
```

## 📚 Related Documentation

- **[Setup Guide](setup.md)** - Initial installation and configuration
- **[Configuration Guide](configuration.md)** - Detailed configuration options
- **[Usage Guide](usage.md)** - Daily commands and workflows
- **[Services Guide](services.md)** - Service-specific information
- **[Quick Reference](reference.md)** - Commands cheatsheet

## 🏥 Emergency Procedures

---

## 📚 See Also

- [README](../README.md)
- [Setup Guide](setup.md)
- [Configuration Guide](configuration.md)
- [Services Guide](services.md)
- [Usage Guide](usage.md)
- [Integration Guide](integration.md)
- [Contributing Guide](contributing.md)

### Complete System Recovery

```bash
# 1. Stop everything
./scripts/manage.sh cleanup-all

# 2. Clean Docker completely
docker system prune -a --volumes

# 3. Restart Docker
# macOS: colima stop && colima start
# Linux: sudo systemctl restart docker

# 4. Start fresh
./scripts/setup.sh --init
./scripts/setup.sh

# 5. Verify
./scripts/manage.sh status
```

### Data Recovery

```bash
# If you have backups
./scripts/manage.sh restore postgres backup.sql

# If no backups, check Docker volumes
docker volume ls | grep postgres
# Mount volume to recover data
docker run --rm -v VOLUME_NAME:/data -v $(pwd):/backup alpine cp -r /data /backup/
```

Remember: Most issues can be resolved by restarting services or recreating them with `./scripts/setup.sh --force`. When in doubt, start with the simplest solutions first.
