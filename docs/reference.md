# Quick Reference & Cheatsheet

This is a quick reference guide for the Local Development Framework, providing essential commands, configurations, and troubleshooting tips.

<!-- AUTO-GENERATED-START -->
# Command Reference (dev-stack)

This section is auto-generated from `scripts/commands.yaml`.

## setup
- `--init`
- `--services`
- `--project`
- `--cleanup-existing`
- `--connect-existing`
- `--force`
- `--dry-run`
- `--skip-validation`
- `--validate-only`
- `--debug`
- `--config`
- `--list-services`
- `--interactive`

## manage
- `start`
- `stop`
- `restart`
- `info`
- `status`
- `logs`
- `connect`
- `exec`
- `backup`
- `restore`
- `update`
- `cleanup`
- `list-all`
- `cleanup-all`
- `monitor`
- `scale`
- `cp`
- `help`
<!-- AUTO-GENERATED-END -->

> **Note:** The section above is auto-generated and lists all available commands and flags. For practical usage, workflows, and troubleshooting, see below.

## 🚀 Usage Patterns

### Common Workflows

#### Starting All Services
```bash
./scripts/manage.sh start
```
See the auto-generated section above for all available flags.

#### Stopping All Services
```bash
./scripts/manage.sh stop
```

#### Viewing Logs
```bash
./scripts/manage.sh logs
./scripts/manage.sh logs -f
```

#### Service-Specific Actions
```bash
./scripts/manage.sh start redis
./scripts/manage.sh stop postgres
./scripts/manage.sh connect postgres
```

#### Data Management
```bash
./scripts/manage.sh backup postgres
./scripts/manage.sh restore postgres backup.sql
```

#### Multi-Repository Management
```bash
./scripts/manage.sh list-all
./scripts/manage.sh cleanup-all
```

#### Maintenance
```bash
./scripts/manage.sh update
./scripts/manage.sh cleanup-docker
```

> See the auto-generated section above for a complete list of commands and flags.


## 🌐 Default Service Ports

| Service    | Port  | UI/Dashboard Port | Description                |
|------------|-------|-------------------|----------------------------|
| Redis      | 6379  | -                 | Redis CLI                  |
| PostgreSQL | 5432  | -                 | Database connection        |
| MySQL      | 3306  | -                 | Database connection        |
| Jaeger     | 4317  | 16686             | OTLP gRPC / Jaeger UI      |
| Jaeger     | 4318  | -                 | OTLP HTTP                  |
| Prometheus | 9090  | 9090              | Prometheus UI              |
| LocalStack | 4566  | 8055              | AWS API / Dashboard        |
| Kafka      | 9092  | 8080              | Kafka broker / Kafka UI    |
| Zookeeper  | 2181  | -                 | Zookeeper for Kafka        |

## ⚙️ Configuration Templates

### Minimal Setup
```yaml
project:
  name: "my-app"
  environment: "local"

services:
  enabled:
    - redis
    - jaeger
```

### Database Development
```yaml
project:
  name: "data-app"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger

overrides:
  postgres:
    database: "my_app_dev"
    username: "app_user"
```

### Full Observability
```yaml
project:
  name: "monitored-app"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger
    - prometheus

overrides:
  prometheus:
    scrape_configs: |
      - job_name: 'my-app'
        static_configs:
          - targets: ['host.docker.internal:8080']
        metrics_path: '/actuator/prometheus'
```

### AWS Development
```yaml
project:
  name: "cloud-app"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger
    - localstack

overrides:
  localstack:
    services: ["sqs", "sns", "s3", "dynamodb"]
    sqs_queues:
      - name: "events"
        dead_letter_queue: true
    sns_topics:
      - name: "notifications"
        subscriptions:
          - protocol: "sqs"
            endpoint: "events"
```

### Event-Driven Architecture
```yaml
project:
  name: "event-app"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger
    - kafka

overrides:
  kafka:
    auto_create_topics: true
    topics:
      - name: "user-events"
        partitions: 3
      - name: "order-events"
        partitions: 6
```

## 🔧 Common Configuration Overrides

### Redis Configuration
```yaml
overrides:
  redis:
    port: 6379
    password: "dev-password"
    memory_limit: "256m"
    config: |
      maxmemory-policy allkeys-lru
      save 900 1
```

### PostgreSQL Configuration
```yaml
overrides:
  postgres:
    port: 5432
    database: "my_app_dev"
    username: "app_user"
    password: "dev-password"
    memory_limit: "512m"
    log_statement: "all"
```

### LocalStack Configuration
```yaml
overrides:
  localstack:
    services: ["sqs", "sns", "dynamodb"]
    sqs_queues:
      - name: "user-events"
        dead_letter_queue: true
    dynamodb_tables:
      - name: "users"
        attribute_definitions:
          - AttributeName: "id"
            AttributeType: "S"
        key_schema:
          - AttributeName: "id"
            KeyType: "HASH"
        provisioned_throughput:
          ReadCapacityUnits: 5
          WriteCapacityUnits: 5
```

### Kafka Configuration
```yaml
overrides:
  kafka:
    auto_create_topics: true
    topics:
      - name: "events"
        partitions: 3
        replication_factor: 1
        cleanup_policy: "delete"
```

## 🔗 Connection Strings

### PostgreSQL
```
# Standard connection
postgresql://username:password@localhost:5432/database_name

# JDBC URL
jdbc:postgresql://localhost:5432/database_name

# Spring Boot configuration
spring.datasource.url=jdbc:postgresql://localhost:5432/my_app_dev
spring.datasource.username=app_user
spring.datasource.password=dev-password
```

### MySQL
```
# Standard connection
mysql://username:password@localhost:3306/database_name

# JDBC URL
jdbc:mysql://localhost:3306/database_name

# Spring Boot configuration
spring.datasource.url=jdbc:mysql://localhost:3306/my_app_dev
spring.datasource.username=app_user
spring.datasource.password=dev-password
```

### Redis
```
# Standard connection
redis://localhost:6379

# With password
redis://:password@localhost:6379

# Spring Boot configuration
spring.data.redis.host=localhost
spring.data.redis.port=6379
spring.data.redis.password=dev-password
```

### LocalStack AWS Services
```
# AWS CLI configuration
aws --endpoint-url=http://localhost:4566 sqs list-queues

# Spring Boot configuration
cloud.aws.credentials.access-key=test
cloud.aws.credentials.secret-key=test
cloud.aws.sqs.endpoint=http://localhost:4566
```

### Kafka
```
# Bootstrap servers
localhost:9092

# Spring Boot configuration
spring.kafka.bootstrap-servers=localhost:9092
```

## 🐛 Quick Troubleshooting

### Service Won't Start
```bash
# Check Docker
docker info

# Check service status
./scripts/manage.sh status

# View logs
./scripts/manage.sh logs SERVICE_NAME

# Restart service
./scripts/manage.sh restart SERVICE_NAME

# Recreate service
./scripts/setup.sh --services=SERVICE_NAME --force
```

### Port Conflicts
```bash
# Find what's using the port
lsof -i :PORT_NUMBER

# Kill the process
kill -9 PID

# Use different port in config
# overrides:
#   service_name:
#     port: NEW_PORT
```

### Cannot Connect to Service
```bash
# Test connection
telnet localhost PORT_NUMBER

# Check service info
./scripts/manage.sh info SERVICE_NAME

# Check container status
docker ps | grep SERVICE_NAME

# Check container logs
./scripts/manage.sh logs SERVICE_NAME
```

### Memory Issues
```bash
# Check memory usage
docker stats

# Reduce memory limits in config
# overrides:
#   global:
#     memory_limit: "256m"

# Clean up Docker resources
./scripts/manage.sh cleanup-docker
```

### Configuration Issues
```bash
# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('local-dev-config.yaml'))"

# Validate framework configuration
./scripts/setup.sh --validate-only

# Show resolved configuration
./scripts/setup.sh --debug --dry-run

# Reset to defaults
./scripts/setup.sh --init --force
```

## 🔍 Useful Commands

### Docker Commands
```bash
# List containers
docker ps

# Container resource usage
docker stats

# Clean up Docker
docker system prune -a

# Remove unused volumes
docker volume prune

# Check Docker disk usage
docker system df
```

### Service Testing
```bash
# Test PostgreSQL
psql -h localhost -U postgres -c "SELECT version();"

# Test MySQL
mysql -h localhost -u root -e "SELECT VERSION();"

# Test Redis
redis-cli -h localhost -p 6379 ping

# Test Kafka
kafka-topics --bootstrap-server localhost:9092 --list

# Test LocalStack
curl http://localhost:4566/health
```

### Network Debugging
```bash
# Check port usage
lsof -i :PORT
netstat -tulpn | grep :PORT

# Test connectivity
telnet localhost PORT

# Check DNS resolution
nslookup SERVICE_NAME
```

### File Operations
```bash
# View generated files
cat docker-compose.generated.yml
cat .env.generated
cat application-local.yml.generated

# Backup configuration
cp local-dev-config.yaml local-dev-config.yaml.bak

# Compare configurations
diff config1.yaml config2.yaml
```

## 🌟 Best Practices

### Configuration Management
- Use version control for `local-dev-config.yaml`
- Create team-specific configuration templates
- Document custom configurations
- Regular validation with `--validate-only`

### Resource Management
- Only enable services you need
- Set appropriate memory limits
- Monitor resource usage regularly
- Clean up Docker resources weekly

### Development Workflow
- Start services before application
- Use `--connect-existing` for shared development
- Backup databases before major changes
- Stop services when not developing

### Troubleshooting Approach
1. Check Docker daemon status
2. Verify service status
3. Review service logs
4. Test network connectivity
5. Validate configuration
6. Try simple restart first

## 📁 File Structure Reference
```
your-project/
├── local-dev-framework/           # Framework directory
│   ├── scripts/
│   │   ├── setup.sh              # Setup script
│   │   └── manage.sh             # Management script
│   ├── services/                  # Service definitions
│   ├── config/                    # Framework configuration
│   └── local-dev-config.sample.yaml
├── local-dev-config.yaml         # Your configuration
├── docker-compose.generated.yml  # Generated Docker Compose
├── .env.generated                 # Generated environment variables
├── application-local.yml.generated # Generated Spring config
└── backups/                       # Database backups
```

## 🚀 Performance Tips

### Fast Startup
- Disable unnecessary services
- Use `--no-pull` to skip image updates
- Pre-start services with `./scripts/manage.sh start`
- Use SSD for Docker storage

### Memory Optimization
- Set conservative memory limits
- Disable Redis persistence for speed
- Use minimal service configurations
- Monitor with `docker stats`

### Development Speed
- Use `--connect-existing` for shared services
- Keep data between sessions (use `stop` not `cleanup`)
- Pre-pull images during setup
- Use fast file systems (avoid network mounts)

## 📞 Getting Help

### Self-Diagnosis
```bash
# Health check
./scripts/manage.sh status
./scripts/manage.sh info
docker info

# View help
./scripts/setup.sh --help
./scripts/manage.sh --help

# Debug mode
./scripts/setup.sh --debug --dry-run
```

### Information for Support
```bash
# System info
uname -a
docker --version

# Framework status
./scripts/manage.sh info
cat local-dev-config.yaml

# Recent logs
./scripts/manage.sh logs --since=1h
```

Remember: Most issues can be resolved with `./scripts/manage.sh restart` or `./scripts/setup.sh --force`. When in doubt, check the logs first!