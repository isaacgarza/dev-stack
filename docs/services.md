---
title: "Services"
description: "Available services and configuration options for dev-stack"
lead: "Explore all the services you can use with dev-stack and how to configure them"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 30
toc: true

---

## Service Management Overview

dev-stack makes it easy to manage common development services with simple commands and consistent configurations.

### Managing Services

**Add services to your stack:**
```bash
# Add individual services
dev-stack service add redis
dev-stack service add postgres

# Add monitoring stack
dev-stack service add prometheus grafana

# Add message queue with UI
dev-stack service add kafka
```

**Remove or modify services:**
```bash
# Remove a service
dev-stack service remove redis

# List available services
dev-stack service list

# View service configuration
dev-stack service config postgres
```

### Common Service Combinations

**Database Stack:**
- PostgreSQL + Redis + Adminer

**Monitoring Stack:**
- Prometheus + Grafana + Jaeger

**Message Queue Stack:**
- Kafka + Kafka UI + Schema Registry

**AWS Local Development:**
- LocalStack + DynamoDB + S3

### Health Monitoring

Built-in health checks ensure your services are running correctly:

```bash
# Check overall system health
dev-stack doctor

# Monitor service status in real-time
dev-stack status --watch

# View detailed service logs
dev-stack logs <service-name>

# Check specific service health
dev-stack health postgres
```

**Health check features:**
- **Automatic service discovery**: Detects running containers
- **Connection validation**: Tests database connections, API endpoints
- **Resource monitoring**: CPU, memory, disk usage per service
- **Dependency checking**: Verifies service startup order

### Service Categories

The following sections detail all available services organized by category:

<!-- AUTO-GENERATED-START -->
# Services Guide (dev-stack)

This section is auto-generated from `services/services.yaml`.

## jaeger

Distributed tracing system for observability.

**Options:**
- `ui_port`
- `otlp_grpc_port`
- `otlp_http_port`
- `memory_limit`
- `sampling_strategy`

**Examples:**
- `curl http://localhost:16686`
- `spring.sleuth.enabled=true`

**Usage Notes:** Use Jaeger to trace requests across microservices. Access the UI at the configured port.

**Links:**
- [https://www.jaegertracing.io/docs/](https://www.jaegertracing.io/docs/)
- [https://docs.spring.io/spring-cloud-sleuth/docs/current/reference/html/](https://docs.spring.io/spring-cloud-sleuth/docs/current/reference/html/)

## kafka

Event streaming platform for messaging and pub/sub.

**Options:**
- `port`
- `ui_port`
- `zookeeper_port`
- `memory_limit`
- `auto_create_topics`
- `num_partitions`
- `replication_factor`
- `topics`

**Examples:**
- `kafka-topics --bootstrap-server localhost:9092 --list`
- `spring.kafka.bootstrap-servers=localhost:9092`

**Usage Notes:** Use Kafka for event-driven architectures. Configure topics and partitions as needed.

**Links:**
- [https://kafka.apache.org/documentation/](https://kafka.apache.org/documentation/)
- [https://docs.spring.io/spring-kafka/docs/current/reference/html/](https://docs.spring.io/spring-kafka/docs/current/reference/html/)

## localstack

AWS cloud services emulator for local development.

**Options:**
- `port`
- `dashboard_port`
- `memory_limit`
- `services`
- `sqs_queues`
- `sns_topics`
- `dynamodb_tables`

**Examples:**
- `aws --endpoint-url=http://localhost:4566 sqs list-queues`
- `spring.cloud.aws.sqs.endpoint=http://localhost:4566`

**Usage Notes:** Emulates AWS APIs for local testing. Enable only the services you need for faster startup.

**Links:**
- [https://docs.localstack.cloud/](https://docs.localstack.cloud/)
- [https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-endpoints.html](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-endpoints.html)

## mysql

Relational database (MySQL) as an alternative to PostgreSQL.

**Options:**
- `port`
- `database`
- `username`
- `password`
- `root_password`
- `memory_limit`
- `character_set`
- `collation`
- `sql_mode`
- `innodb_buffer_pool_size`

**Examples:**
- `mysql -h localhost -u root -e "SELECT VERSION();"`
- `spring.datasource.url=jdbc:mysql://localhost:3306/my_app_dev`

**Usage Notes:** Use MySQL for compatibility with legacy systems or when required by application stack.

**Links:**
- [https://dev.mysql.com/doc/](https://dev.mysql.com/doc/)
- [https://spring.io/projects/spring-data-jpa](https://spring.io/projects/spring-data-jpa)

## postgres

Relational database (PostgreSQL) for structured data.

**Options:**
- `port`
- `database`
- `username`
- `password`
- `memory_limit`
- `shared_preload_libraries`
- `log_statement`
- `log_duration`
- `shared_buffers`
- `effective_cache_size`
- `work_mem`

**Examples:**
- `psql -h localhost -U postgres -c "SELECT version();"`
- `spring.datasource.url=jdbc:postgresql://localhost:5432/my_app_dev`

**Usage Notes:** Ideal for structured data and transactional workloads. Use overrides to set custom database/user.

**Links:**
- [https://www.postgresql.org/docs/](https://www.postgresql.org/docs/)
- [https://spring.io/projects/spring-data-jpa](https://spring.io/projects/spring-data-jpa)

## prometheus

Metrics collection and monitoring system.

**Options:**
- `port`
- `scrape_interval`
- `memory_limit`
- `retention_time`
- `scrape_configs`

**Examples:**
- `curl http://localhost:9090`
- `spring.metrics.export.prometheus.enabled=true`

**Usage Notes:** Prometheus scrapes metrics from configured endpoints. Use for monitoring and alerting.

**Links:**
- [https://prometheus.io/docs/](https://prometheus.io/docs/)
- [https://micrometer.io/docs/registry/prometheus](https://micrometer.io/docs/registry/prometheus)

## redis

In-memory data store for caching and session storage.

**Options:**
- `port`
- `password`
- `memory_limit`
- `persistence`
- `config`

**Examples:**
- `redis-cli -h localhost -p 6379 ping`
- `spring.data.redis.host=localhost`

**Usage Notes:** Use Redis for caching, session storage, and pub/sub. Set a password for production-like security.

**Links:**
- [https://redis.io/documentation](https://redis.io/documentation)
- [https://spring.io/projects/spring-data-redis](https://spring.io/projects/spring-data-redis)
<!-- AUTO-GENERATED-END -->

## 🎯 What's Next?

Now that you understand service management, take your setup further:

1. **[Advanced configuration](configuration.md)** - Fine-tune service settings, environment variables, and overrides
2. **[Application integration](integration.md)** - Connect your apps to these services with code examples
3. **[Production considerations](troubleshooting.md#performance-optimization)** - Optimize for performance and reliability
4. **[Contributing new services](contributing.md)** - Add services that aren't available yet

**Advanced service management:**
- Create service profiles for different environments
- Set up service dependencies and startup ordering
- Configure custom health checks and monitoring

**Need a service that's not listed?** Check the [Contributing Guide](contributing.md) to learn how to add new services.
