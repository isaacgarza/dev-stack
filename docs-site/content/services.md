---
title: "Services"
description: "Available services and configuration options for dev-stack"
lead: "Explore all the services you can use with dev-stack and how to configure them"
date: 2025-10-07T12:20:14-05:00
lastmod: 2025-10-07T12:20:14-05:00
draft: false
weight: 30
toc: true
---

<!-- AUTO-GENERATED-START -->
# Available Services

7 services available for your development stack.

## Service Categories


### General

- **[jaeger](#jaeger)** - Distributed tracing system for observability.

- **[kafka](#kafka)** - Event streaming platform for messaging and pub/sub.

- **[localstack](#localstack)** - AWS cloud services emulator for local development.

- **[mysql](#mysql)** - Relational database (MySQL) as an alternative to PostgreSQL.

- **[postgres](#postgres)** - Relational database (PostgreSQL) for structured data.

- **[prometheus](#prometheus)** - Metrics collection and monitoring system.

- **[redis](#redis)** - In-memory data store for caching and session storage.




## Service Reference


### jaeger

Distributed tracing system for observability.










**Configuration Options:**

- ui_port

- otlp_grpc_port

- otlp_http_port

- memory_limit

- sampling_strategy




**Examples:**

```bash
curl http://localhost:16686
```

```bash
spring.sleuth.enabled=true
```




**Usage Notes:**

Use Jaeger to trace requests across microservices. Access the UI at the configured port.



**Links:**

- [Documentation](https://www.jaegertracing.io/docs/)

- [Documentation](https://docs.spring.io/spring-cloud-sleuth/docs/current/reference/html/)



---


### kafka

Event streaming platform for messaging and pub/sub.










**Configuration Options:**

- port

- ui_port

- zookeeper_port

- memory_limit

- auto_create_topics

- num_partitions

- replication_factor

- topics




**Examples:**

```bash
kafka-topics --bootstrap-server localhost:9092 --list
```

```bash
spring.kafka.bootstrap-servers=localhost:9092
```




**Usage Notes:**

Use Kafka for event-driven architectures. Configure topics and partitions as needed.



**Links:**

- [Documentation](https://kafka.apache.org/documentation/)

- [Documentation](https://docs.spring.io/spring-kafka/docs/current/reference/html/)



---


### localstack

AWS cloud services emulator for local development.










**Configuration Options:**

- port

- dashboard_port

- memory_limit

- services

- sqs_queues

- sns_topics

- dynamodb_tables




**Examples:**

```bash
aws --endpoint-url=http://localhost:4566 sqs list-queues
```

```bash
spring.cloud.aws.sqs.endpoint=http://localhost:4566
```




**Usage Notes:**

Emulates AWS APIs for local testing. Enable only the services you need for faster startup.



**Links:**

- [Documentation](https://docs.localstack.cloud/)

- [Documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-endpoints.html)



---


### mysql

Relational database (MySQL) as an alternative to PostgreSQL.










**Configuration Options:**

- port

- database

- username

- password

- root_password

- memory_limit

- character_set

- collation

- sql_mode

- innodb_buffer_pool_size




**Examples:**

```bash
mysql -h localhost -u root -e "SELECT VERSION();"
```

```bash
spring.datasource.url=jdbc:mysql://localhost:3306/my_app_dev
```




**Usage Notes:**

Use MySQL for compatibility with legacy systems or when required by application stack.



**Links:**

- [Documentation](https://dev.mysql.com/doc/)

- [Documentation](https://spring.io/projects/spring-data-jpa)



---


### postgres

Relational database (PostgreSQL) for structured data.










**Configuration Options:**

- port

- database

- username

- password

- memory_limit

- shared_preload_libraries

- log_statement

- log_duration

- shared_buffers

- effective_cache_size

- work_mem




**Examples:**

```bash
psql -h localhost -U postgres -c "SELECT version();"
```

```bash
spring.datasource.url=jdbc:postgresql://localhost:5432/my_app_dev
```




**Usage Notes:**

Ideal for structured data and transactional workloads. Use overrides to set custom database/user.



**Links:**

- [Documentation](https://www.postgresql.org/docs/)

- [Documentation](https://spring.io/projects/spring-data-jpa)



---


### prometheus

Metrics collection and monitoring system.










**Configuration Options:**

- port

- scrape_interval

- memory_limit

- retention_time

- scrape_configs




**Examples:**

```bash
curl http://localhost:9090
```

```bash
spring.metrics.export.prometheus.enabled=true
```




**Usage Notes:**

Prometheus scrapes metrics from configured endpoints. Use for monitoring and alerting.



**Links:**

- [Documentation](https://prometheus.io/docs/)

- [Documentation](https://micrometer.io/docs/registry/prometheus)



---


### redis

In-memory data store for caching and session storage.










**Configuration Options:**

- port

- password

- memory_limit

- persistence

- config




**Examples:**

```bash
redis-cli -h localhost -p 6379 ping
```

```bash
spring.data.redis.host=localhost
```




**Usage Notes:**

Use Redis for caching, session storage, and pub/sub. Set a password for production-like security.



**Links:**

- [Documentation](https://redis.io/documentation)

- [Documentation](https://spring.io/projects/spring-data-redis)



---



<!-- AUTO-GENERATED-END -->