---
title: "Services Reference"
description: "Complete reference of available services for dev-stack"
weight: 40
---

# Services Reference

This page provides a comprehensive reference for all services available in dev-stack. Services are containerized applications that your project can depend on, such as databases, message queues, and monitoring tools.

## Service Categories

### Databases

#### PostgreSQL
**Service ID:** `postgres`  
**Description:** PostgreSQL relational database  
**Default Port:** 5432

```bash
# Add PostgreSQL to your project
dev-stack service add postgres

# With custom configuration
dev-stack service add postgres --port 5433 --database myapp
```

**Configuration Options:**
- `port` - Database port (default: 5432)
- `database` - Database name (default: project name)
- `username` - Database username (default: postgres)
- `password` - Database password (default: auto-generated)
- `version` - PostgreSQL version (default: 15)

**Environment Variables:**
- `POSTGRES_DB` - Database name
- `POSTGRES_USER` - Database user
- `POSTGRES_PASSWORD` - Database password
- `DATABASE_URL` - Full connection string

**Health Check:** `pg_isready -U postgres`

#### MySQL
**Service ID:** `mysql`  
**Description:** MySQL relational database  
**Default Port:** 3306

```bash
# Add MySQL to your project
dev-stack service add mysql

# With custom configuration
dev-stack service add mysql --port 3307 --database myapp
```

**Configuration Options:**
- `port` - Database port (default: 3306)
- `database` - Database name (default: project name)
- `username` - Database username (default: root)
- `password` - Root password (default: auto-generated)
- `version` - MySQL version (default: 8.0)

**Environment Variables:**
- `MYSQL_DATABASE` - Database name
- `MYSQL_USER` - Database user
- `MYSQL_PASSWORD` - User password
- `MYSQL_ROOT_PASSWORD` - Root password

#### MongoDB
**Service ID:** `mongodb`  
**Description:** MongoDB NoSQL document database  
**Default Port:** 27017

```bash
# Add MongoDB to your project
dev-stack service add mongodb

# With authentication
dev-stack service add mongodb --username admin --password secret
```

**Configuration Options:**
- `port` - Database port (default: 27017)
- `database` - Database name (default: project name)
- `username` - Database username (optional)
- `password` - Database password (optional)
- `version` - MongoDB version (default: 6.0)

**Environment Variables:**
- `MONGO_INITDB_DATABASE` - Initial database
- `MONGO_INITDB_ROOT_USERNAME` - Root username
- `MONGO_INITDB_ROOT_PASSWORD` - Root password

#### Redis
**Service ID:** `redis`  
**Description:** Redis in-memory data store for caching and session storage  
**Default Port:** 6379

```bash
# Add Redis to your project
dev-stack service add redis

# With password protection
dev-stack service add redis --password mypassword
```

**Configuration Options:**
- `port` - Redis port (default: 6379)
- `password` - Redis password (optional)
- `version` - Redis version (default: 7.0)
- `persistent` - Enable data persistence (default: true)

**Environment Variables:**
- `REDIS_PASSWORD` - Redis password
- `REDIS_URL` - Full connection string

**Health Check:** `redis-cli ping`

### Message Queues

#### RabbitMQ
**Service ID:** `rabbitmq`  
**Description:** RabbitMQ message broker for reliable messaging  
**Default Port:** 5672 (AMQP), 15672 (Management UI)

```bash
# Add RabbitMQ to your project
dev-stack service add rabbitmq

# With custom credentials
dev-stack service add rabbitmq --username admin --password secret
```

**Configuration Options:**
- `port` - AMQP port (default: 5672)
- `management_port` - Management UI port (default: 15672)
- `username` - RabbitMQ username (default: guest)
- `password` - RabbitMQ password (default: guest)
- `version` - RabbitMQ version (default: 3.11)

**Environment Variables:**
- `RABBITMQ_DEFAULT_USER` - Default username
- `RABBITMQ_DEFAULT_PASS` - Default password
- `RABBITMQ_URL` - Full connection string

**Management UI:** http://localhost:15672

#### Apache Kafka
**Service ID:** `kafka`  
**Description:** Apache Kafka distributed streaming platform  
**Default Port:** 9092

```bash
# Add Kafka to your project (includes Zookeeper)
dev-stack service add kafka
```

**Configuration Options:**
- `port` - Kafka port (default: 9092)
- `zookeeper_port` - Zookeeper port (default: 2181)
- `topics` - Initial topics to create
- `version` - Kafka version (default: 3.4)

**Environment Variables:**
- `KAFKA_BOOTSTRAP_SERVERS` - Kafka broker list
- `KAFKA_ZOOKEEPER_CONNECT` - Zookeeper connection

#### NATS
**Service ID:** `nats`  
**Description:** NATS messaging system for cloud native applications  
**Default Port:** 4222

```bash
# Add NATS to your project
dev-stack service add nats
```

**Configuration Options:**
- `port` - NATS port (default: 4222)
- `monitoring_port` - Monitoring port (default: 8222)
- `cluster` - Enable clustering (default: false)
- `version` - NATS version (default: 2.9)

### Monitoring and Observability

#### Prometheus
**Service ID:** `prometheus`  
**Description:** Prometheus monitoring and alerting toolkit  
**Default Port:** 9090

```bash
# Add Prometheus to your project
dev-stack service add prometheus
```

**Configuration Options:**
- `port` - Prometheus port (default: 9090)
- `retention` - Data retention period (default: 15d)
- `scrape_interval` - Metrics collection interval (default: 15s)
- `version` - Prometheus version (default: 2.42)

**Web UI:** http://localhost:9090

#### Grafana
**Service ID:** `grafana`  
**Description:** Grafana analytics and monitoring platform  
**Default Port:** 3000

```bash
# Add Grafana to your project
dev-stack service add grafana

# Often used with Prometheus
dev-stack service add prometheus grafana
```

**Configuration Options:**
- `port` - Grafana port (default: 3000)
- `username` - Admin username (default: admin)
- `password` - Admin password (default: admin)
- `version` - Grafana version (default: 9.4)

**Web UI:** http://localhost:3000  
**Default Credentials:** admin/admin

#### Jaeger
**Service ID:** `jaeger`  
**Description:** Jaeger distributed tracing system  
**Default Port:** 16686 (UI), 14268 (HTTP)

```bash
# Add Jaeger to your project
dev-stack service add jaeger
```

**Configuration Options:**
- `ui_port` - Jaeger UI port (default: 16686)
- `http_port` - HTTP collector port (default: 14268)
- `udp_port` - UDP collector port (default: 14267)
- `version` - Jaeger version (default: 1.42)

**Web UI:** http://localhost:16686

### Search Engines

#### Elasticsearch
**Service ID:** `elasticsearch`  
**Description:** Elasticsearch search and analytics engine  
**Default Port:** 9200

```bash
# Add Elasticsearch to your project
dev-stack service add elasticsearch

# With Kibana for visualization
dev-stack service add elasticsearch kibana
```

**Configuration Options:**
- `port` - Elasticsearch port (default: 9200)
- `cluster_name` - Cluster name (default: dev-stack)
- `heap_size` - JVM heap size (default: 1g)
- `version` - Elasticsearch version (default: 8.6)

**Environment Variables:**
- `ELASTICSEARCH_URL` - Elasticsearch endpoint
- `ES_JAVA_OPTS` - JVM options

#### Kibana
**Service ID:** `kibana`  
**Description:** Kibana data visualization for Elasticsearch  
**Default Port:** 5601

```bash
# Add Kibana (requires Elasticsearch)
dev-stack service add elasticsearch kibana
```

**Configuration Options:**
- `port` - Kibana port (default: 5601)
- `elasticsearch_url` - Elasticsearch endpoint
- `version` - Kibana version (default: 8.6)

**Web UI:** http://localhost:5601

#### OpenSearch
**Service ID:** `opensearch`  
**Description:** OpenSearch search and analytics suite  
**Default Port:** 9200

```bash
# Add OpenSearch to your project
dev-stack service add opensearch
```

**Configuration Options:**
- `port` - OpenSearch port (default: 9200)
- `dashboard_port` - Dashboard port (default: 5601)
- `cluster_name` - Cluster name (default: dev-stack)
- `version` - OpenSearch version (default: 2.5)

### Web Servers and Proxies

#### Nginx
**Service ID:** `nginx`  
**Description:** Nginx web server and reverse proxy  
**Default Port:** 80

```bash
# Add Nginx to your project
dev-stack service add nginx

# With custom configuration
dev-stack service add nginx --port 8080
```

**Configuration Options:**
- `port` - HTTP port (default: 80)
- `ssl_port` - HTTPS port (default: 443)
- `config_file` - Custom nginx.conf path
- `version` - Nginx version (default: 1.23)

**Configuration File:** `./nginx/nginx.conf`

#### Traefik
**Service ID:** `traefik`  
**Description:** Traefik reverse proxy and load balancer  
**Default Port:** 80 (HTTP), 8080 (Dashboard)

```bash
# Add Traefik to your project
dev-stack service add traefik
```

**Configuration Options:**
- `port` - HTTP port (default: 80)
- `dashboard_port` - Dashboard port (default: 8080)
- `api_enabled` - Enable API/Dashboard (default: true)
- `version` - Traefik version (default: 2.9)

**Dashboard:** http://localhost:8080

### Development Tools

#### MailHog
**Service ID:** `mailhog`  
**Description:** Email testing tool for development  
**Default Port:** 8025 (UI), 1025 (SMTP)

```bash
# Add MailHog to your project
dev-stack service add mailhog
```

**Configuration Options:**
- `ui_port` - Web UI port (default: 8025)
- `smtp_port` - SMTP port (default: 1025)
- `version` - MailHog version (default: 1.0.1)

**Web UI:** http://localhost:8025  
**SMTP:** localhost:1025

#### Adminer
**Service ID:** `adminer`  
**Description:** Database management tool  
**Default Port:** 8080

```bash
# Add Adminer to your project
dev-stack service add adminer
```

**Configuration Options:**
- `port` - Adminer port (default: 8080)
- `design` - UI theme (default: galkaev)
- `version` - Adminer version (default: 4.8.1)

**Web UI:** http://localhost:8080

#### MinIO
**Service ID:** `minio`  
**Description:** MinIO object storage server  
**Default Port:** 9000 (API), 9001 (Console)

```bash
# Add MinIO to your project
dev-stack service add minio
```

**Configuration Options:**
- `api_port` - API port (default: 9000)
- `console_port` - Console port (default: 9001)
- `access_key` - Access key (default: minioadmin)
- `secret_key` - Secret key (default: minioadmin)
- `version` - MinIO version (default: latest)

**Console:** http://localhost:9001  
**Credentials:** minioadmin/minioadmin

## Service Operations

### Adding Services

```bash
# Add a single service
dev-stack service add postgres

# Add multiple services
dev-stack service add postgres redis nginx

# Add with custom configuration
dev-stack service add postgres --port 5433 --database myapp

# Add with specific version
dev-stack service add redis --version 6.2
```

### Listing Services

```bash
# List all available services
dev-stack service list

# List services by category
dev-stack service list --category database

# Show only installed services
dev-stack service list --installed

# Show detailed service information
dev-stack service info postgres
```

### Removing Services

```bash
# Remove a service
dev-stack service remove redis

# Remove service and its data
dev-stack service remove redis --volumes

# Force removal without confirmation
dev-stack service remove redis --force
```

### Service Configuration

Services can be configured in `dev-stack-config.yaml`:

```yaml
services:
  postgres:
    port: 5432
    database: myapp
    username: developer
    password: secret
    version: "15"
    
  redis:
    port: 6379
    password: myredispassword
    persistent: true
    
  nginx:
    port: 8080
    config_file: ./nginx/custom.conf
```

## Custom Services

You can create custom service definitions by adding YAML files to the `services/` directory:

```yaml
# services/myservice.yaml
name: myservice
description: My custom service
category: custom

container:
  image: myregistry/myservice:latest
  ports:
    - "9000:9000"
  environment:
    - SERVICE_PORT=9000
    - SERVICE_ENV=development
  volumes:
    - ./data:/app/data
  
health_check:
  endpoint: /health
  port: 9000
  interval: 30s
  timeout: 10s
  retries: 3

configuration:
  port:
    description: Service port
    default: 9000
    type: integer
  env:
    description: Service environment
    default: development
    type: string
```

Then use it like any built-in service:

```bash
dev-stack service add myservice
```

## Networking

All services are automatically connected to a shared Docker network, allowing them to communicate using service names as hostnames:

```bash
# From your application, connect to:
postgres://postgres:5432/myapp    # PostgreSQL
redis://redis:6379                # Redis
http://nginx:80                   # Nginx
```

## Data Persistence

Services with persistent data automatically create Docker volumes:

- **PostgreSQL**: `/var/lib/postgresql/data`
- **MySQL**: `/var/lib/mysql`
- **MongoDB**: `/data/db`
- **Redis**: `/data` (if persistence enabled)
- **Elasticsearch**: `/usr/share/elasticsearch/data`

To manage volumes:

```bash
# List volumes
dev-stack volume list

# Backup volume
dev-stack volume backup postgres-data

# Remove volumes when stopping
dev-stack down --volumes
```

## Service Dependencies

Some services have dependencies that are automatically managed:

- **Kafka** → Requires Zookeeper (auto-added)
- **Kibana** → Requires Elasticsearch
- **Grafana** → Works best with Prometheus

dev-stack will automatically add required dependencies when you add a service.

## Troubleshooting Services

### Common Issues

**Service won't start:**
```bash
# Check service logs
dev-stack logs servicename

# Check service configuration
dev-stack service info servicename
```

**Port conflicts:**
```bash
# Check what's using the port
lsof -i :5432

# Change service port
dev-stack service config postgres --port 5433
```

**Connection refused:**
```bash
# Check if service is healthy
dev-stack health servicename

# Restart the service
dev-stack restart servicename
```

### Getting Help

For service-specific help:

1. Use `dev-stack service info <service>` for detailed information
2. Check service logs with `dev-stack logs <service>`
3. Verify health with `dev-stack health <service>`
4. [Open an issue](https://github.com/isaacgarza/dev-stack/issues) for bugs

---

> **Contributing**: Want to add a new service? Check our [contributing guide]({{< ref "/contributing" >}}) to learn how to create service definitions.