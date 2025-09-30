# Configuration Guide (dev-stack)

> For troubleshooting configuration issues, see [Troubleshooting Guide](troubleshooting.md).

This guide covers all configuration options for **dev-stack**, from basic setups to advanced configurations with multiple services.

## 📋 Overview

**dev-stack** uses a single `local-dev-config.yaml` file to define your entire development stack. This configuration-driven approach ensures consistency across team members and projects.

> For a quick checklist of configuration best practices, see the end of this guide.

## 🚀 Quick Start

### 1. Create Configuration File

See the [README](../README.md) for a quick start and command reference.

For a full configuration schema and examples, continue below.

## 🏗️ Configuration Schema

### Project Configuration

```yaml
project:
  name: "my-project"        # Used for container names and network naming
  environment: "local"      # Development environment identifier
```

**Properties:**
- `name` (required): Project identifier used in container and network names
- `environment` (optional): Environment label (default: "local")

### Services Configuration

```yaml
services:
  enabled:
    - redis              # In-memory data store
    - postgres           # Primary database
    - mysql              # Alternative database
    - jaeger             # Distributed tracing
    - prometheus         # Metrics collection
    - localstack         # AWS services emulation
    - kafka              # Event streaming platform
```

**Available Services:**
- **redis**: In-memory data structure store
- **postgres**: PostgreSQL relational database
- **mysql**: MySQL relational database  
- **jaeger**: Distributed tracing system
- **prometheus**: Metrics collection and monitoring
- **localstack**: AWS services emulation (SQS, SNS, DynamoDB, S3, etc.)
- **kafka**: Apache Kafka event streaming platform

### Validation Configuration

```yaml
validation:
  skip_warnings: false              # Skip resource and compatibility warnings
  allow_multiple_databases: false   # Permit both MySQL and PostgreSQL
  auto_start: true                  # Start services after setup
  pull_latest_images: true          # Pull latest Docker images
  cleanup_on_recreate: false        # Keep data when recreating services
```

## ⚙️ Service Overrides

Customize any service configuration using the `overrides` section:

### Redis Configuration

```yaml
overrides:
  redis:
    port: 6379
    password: "dev-password"
    memory_limit: "256m"
    persistence: true
    config: |
      maxmemory-policy allkeys-lru
      save 900 1
```

**Properties:**
- `port`: Redis port (default: 6379)
- `password`: Redis password (default: auto-generated)
- `memory_limit`: Container memory limit
- `persistence`: Enable RDB persistence (default: true)
- `config`: Additional Redis configuration

### PostgreSQL Configuration

```yaml
overrides:
  postgres:
    port: 5432
    database: "my_app_dev"
    username: "app_user"
    password: "dev-password"
    memory_limit: "512m"
    shared_preload_libraries: "pg_stat_statements"
    log_statement: "all"
```

**Properties:**
- `port`: PostgreSQL port (default: 5432)
- `database`: Database name (default: based on project name)
- `username`: Database user (default: based on project name)
- `password`: Database password (default: auto-generated)
- `memory_limit`: Container memory limit
- `shared_preload_libraries`: PostgreSQL extensions
- `log_statement`: SQL logging level

### MySQL Configuration

```yaml
overrides:
  mysql:
    port: 3306
    database: "my_app_dev"
    username: "app_user"
    password: "dev-password"
    root_password: "root-password"
    memory_limit: "512m"
    character_set: "utf8mb4"
    collation: "utf8mb4_unicode_ci"
```

**Properties:**
- `port`: MySQL port (default: 3306)
- `database`: Database name
- `username`: Database user
- `password`: User password
- `root_password`: Root password
- `memory_limit`: Container memory limit
- `character_set`: Default character set
- `collation`: Default collation

### Jaeger Configuration

```yaml
overrides:
  jaeger:
    ui_port: 16686
    otlp_grpc_port: 4317
    otlp_http_port: 4318
    memory_limit: "256m"
    sampling_strategy: |
      {
        "service_strategies": [
          {
            "service": "my-service",
            "type": "probabilistic",
            "param": 1.0
          }
        ],
        "default_strategy": {
          "type": "probabilistic",
          "param": 0.1
        }
      }
```

**Properties:**
- `ui_port`: Jaeger UI port (default: 16686)
- `otlp_grpc_port`: OTLP gRPC receiver port (default: 4317)
- `otlp_http_port`: OTLP HTTP receiver port (default: 4318)
- `memory_limit`: Container memory limit
- `sampling_strategy`: Jaeger sampling configuration

### Prometheus Configuration

```yaml
overrides:
  prometheus:
    port: 9090
    scrape_interval: "15s"
    memory_limit: "256m"
    retention_time: "15d"
    scrape_configs: |
      - job_name: 'my-app'
        static_configs:
          - targets: ['host.docker.internal:8080']
        scrape_interval: 5s
        metrics_path: '/actuator/prometheus'
```

**Properties:**
- `port`: Prometheus port (default: 9090)
- `scrape_interval`: Global scrape interval
- `memory_limit`: Container memory limit
- `retention_time`: Metrics retention period
- `scrape_configs`: Additional scrape configurations

### LocalStack Configuration

```yaml
overrides:
  localstack:
    port: 4566
    dashboard_port: 8055
    memory_limit: "512m"
    services:
      - sqs
      - sns
      - dynamodb
      - s3
    
    # SQS queues to create automatically
    sqs_queues:
      - name: "user-events"
        visibility_timeout: 30
        message_retention_period: 1209600  # 14 days
        max_receive_count: 3
        dead_letter_queue: true  # Creates "user-events-dlq"
      - name: "notifications"
        dead_letter_queue: "notifications-dlq"  # Custom DLQ name
    
    # SNS topics to create automatically
    sns_topics:
      - name: "user-notifications"
        display_name: "User Notifications"
        subscriptions:
          - protocol: "sqs"
            endpoint: "user-events"
            raw_message_delivery: true
    
    # DynamoDB tables to create automatically
    dynamodb_tables:
      - name: "users"
        attribute_definitions:
          - AttributeName: "userId"
            AttributeType: "S"
          - AttributeName: "email"
            AttributeType: "S"
        key_schema:
          - AttributeName: "userId"
            KeyType: "HASH"
        provisioned_throughput:
          ReadCapacityUnits: 5
          WriteCapacityUnits: 5
        global_secondary_indexes:
          - IndexName: "EmailIndex"
            KeySchema:
              - AttributeName: "email"
                KeyType: "HASH"
            Projection:
              ProjectionType: "ALL"
            ProvisionedThroughput:
              ReadCapacityUnits: 5
              WriteCapacityUnits: 5
```

**LocalStack Properties:**
- `port`: Main LocalStack port (default: 4566)
- `dashboard_port`: LocalStack Web UI port (default: 8055)
- `services`: AWS services to enable
- `sqs_queues`: SQS queues to auto-create
- `sns_topics`: SNS topics to auto-create
- `dynamodb_tables`: DynamoDB tables to auto-create

**SQS Queue Properties:**
- `name`: Queue name (required)
- `visibility_timeout`: Message visibility timeout in seconds
- `message_retention_period`: Message retention in seconds
- `max_receive_count`: Max receives before DLQ
- `dead_letter_queue`: `true` for auto-naming, string for custom name

**SNS Topic Properties:**
- `name`: Topic name (required)
- `display_name`: Human-readable name
- `subscriptions`: Array of subscriptions

**DynamoDB Table Properties:**
- `name`: Table name (required)
- `attribute_definitions`: Column definitions
- `key_schema`: Primary key definition
- `provisioned_throughput`: Read/write capacity
- `global_secondary_indexes`: GSI definitions

### Kafka Configuration

```yaml
overrides:
  kafka:
    port: 9092
    ui_port: 8080
    zookeeper_port: 2181
    memory_limit: "1g"
    auto_create_topics: true
    num_partitions: 3
    replication_factor: 1
    
    # Custom topics to create
    topics:
      - name: "user-events"
        partitions: 3
        replication_factor: 1
        cleanup_policy: "delete"
        retention_ms: 604800000  # 7 days
      - name: "order-processing"
        partitions: 6
        replication_factor: 1
      - name: "user-profiles"
        partitions: 2
        cleanup_policy: "compact"
```

**Kafka Properties:**
- `port`: Kafka broker port (default: 9092)
- `ui_port`: Kafka UI port (default: 8080)
- `zookeeper_port`: Zookeeper port (default: 2181)
- `auto_create_topics`: Enable automatic topic creation
- `num_partitions`: Default partitions for auto-created topics
- `topics`: Custom topics to create

**Topic Properties:**
- `name`: Topic name (required)
- `partitions`: Number of partitions
- `replication_factor`: Replication factor
- `cleanup_policy`: `delete`, `compact`, or `compact,delete`
- `retention_ms`: Message retention in milliseconds

## 📚 Common Configuration Examples

### Minimal Setup (Caching + Tracing)

```yaml
project:
  name: "minimal-api"
  environment: "local"

services:
  enabled:
    - redis
    - jaeger

overrides:
  redis:
    memory_limit: "128m"
```

### Database Development

```yaml
project:
  name: "data-api"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger

overrides:
  postgres:
    database: "data_api_dev"
    username: "data_user"
    log_statement: "all"  # Log all SQL statements
```

### Full Observability Stack

```yaml
project:
  name: "monitored-api"
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
  name: "cloud-api"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger
    - localstack

overrides:
  localstack:
    services:
      - sqs
      - sns
      - s3
      - dynamodb
    sqs_queues:
      - name: "user-events"
        dead_letter_queue: true
      - name: "notifications"
    sns_topics:
      - name: "user-notifications"
        subscriptions:
          - protocol: "sqs"
            endpoint: "user-events"
```

### Event-Driven Architecture

```yaml
project:
  name: "event-api"
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
      - name: "user-profiles"
        cleanup_policy: "compact"
```

### High-Performance Setup

```yaml
project:
  name: "high-perf-api"
  environment: "local"

services:
  enabled:
    - redis
    - postgres

overrides:
  redis:
    memory_limit: "1g"
    config: |
      maxmemory-policy allkeys-lru
      tcp-keepalive 60
  postgres:
    memory_limit: "1g"
    shared_preload_libraries: "pg_stat_statements"
    shared_buffers: "256MB"
    effective_cache_size: "1GB"

validation:
  skip_warnings: true  # Skip resource warnings
```

## 🔧 Advanced Configuration

### Environment-Specific Configurations

You can create different configurations for different environments:

```bash
# Development configuration
cp local-dev-config.yaml local-dev-config.dev.yaml

# Testing configuration  
cp local-dev-config.yaml local-dev-config.test.yaml

# Use specific config
./scripts/setup.sh --config=local-dev-config.test.yaml
```

### Configuration Validation

The framework validates your configuration and provides warnings:

```yaml
validation:
  skip_warnings: false              # Show all warnings
  allow_multiple_databases: true    # Allow both MySQL and PostgreSQL
  auto_start: true                  # Start services after setup
  strict_mode: false                # Strict validation mode
```

### Resource Management

```yaml
# Global resource settings
resources:
  memory_limit: "8g"        # Total memory limit for all services
  cpu_limit: "4"            # Total CPU limit
  disk_limit: "50g"         # Total disk limit

# Apply to all services
overrides:
  global:
    memory_limit: "512m"    # Default memory per service
    restart_policy: "unless-stopped"
```

### Custom Networks

```yaml
# Custom Docker network configuration
network:
  name: "my-app-network"
  driver: "bridge"
  subnet: "172.20.0.0/16"
  ip_range: "172.20.240.0/20"
```

## 🚨 Configuration Best Practices

### 1. Resource Allocation

- **Development**: Allocate 6-8GB RAM total
- **CI/CD**: Use minimal configurations
- **Team Sharing**: Use consistent configurations

### 2. Security

```yaml
# Use strong passwords in team configurations
overrides:
  postgres:
    password: "${POSTGRES_PASSWORD:-dev-password}"
  redis:
    password: "${REDIS_PASSWORD:-dev-password}"
```

### 3. Performance

```yaml
# Optimize for development speed
overrides:
  postgres:
    fsync: "off"                    # Faster writes (development only)
    synchronous_commit: "off"       # Async commits
  redis:
    save: ""                        # Disable persistence for speed
```

### 4. Debugging

```yaml
# Enable detailed logging for debugging
overrides:
  postgres:
    log_statement: "all"
    log_duration: "on"
  kafka:
    log_level: "DEBUG"
```

## 🔄 Configuration Migration

### From Template-Based Setup

If migrating from the old template system:

```bash
# Old way
./scripts/setup.sh --template=web-api

# New way - create equivalent config
cat > local-dev-config.yaml << EOF
services:
  enabled:
    - redis
    - postgres
    - jaeger
EOF
```

### Version Updates

When updating the framework:

```bash
# Backup current config
cp local-dev-config.yaml local-dev-config.yaml.bak

# Generate new sample
./scripts/setup.sh --init --force

# Merge changes manually
diff local-dev-config.yaml.bak local-dev-config.yaml
```

## 📋 Configuration Reference

### Complete Example

```yaml
# Complete configuration example
project:
  name: "my-awesome-api"
  environment: "local"

services:
  enabled:
    - redis
    - postgres
    - jaeger
    - prometheus
    - localstack
    - kafka

overrides:
  redis:
    port: 6379
    password: "dev-redis-password"
    memory_limit: "256m"
    
  postgres:
    port: 5432
    database: "awesome_api_dev"
    username: "api_user"
    password: "dev-db-password"
    memory_limit: "512m"
    
  jaeger:
    ui_port: 16686
    memory_limit: "256m"
    
  prometheus:
    port: 9090
    scrape_interval: "15s"
    
  localstack:
    services: ["sqs", "sns", "s3"]
    sqs_queues:
      - name: "events"
        dead_letter_queue: true
        
  kafka:
    auto_create_topics: true
    topics:
      - name: "user-events"
        partitions: 3

validation:
  skip_warnings: false
  allow_multiple_databases: false
  auto_start: true
```

## 🆘 Troubleshooting Configuration

### Common Issues

**Invalid YAML syntax:**
```bash
# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('local-dev-config.yaml'))"
```

**Service conflicts:**
```bash
# Check for port conflicts
./scripts/setup.sh --dry-run
```

**Resource warnings:**
```bash
# Skip warnings
./scripts/setup.sh --skip-validation
```

**Configuration not found:**
```bash
# Create default configuration
./scripts/setup.sh --init
```

### Debug Mode

```bash
# Run setup with debug information
./scripts/setup.sh --debug

# Validate configuration only
./scripts/setup.sh --validate-only
```

## 🧭 See Also

- [README](../README.md)
- [Services Guide](services.md)
- [Usage Guide](usage.md)
- [Integration Guide](integration.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Contributing Guide](contributing.md)