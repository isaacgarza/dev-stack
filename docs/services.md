# Services Guide (dev-stack)

This guide provides detailed information about each service available in the Local Development Framework, including configuration options, use cases, and integration examples.

<!-- AUTO-GENERATED-START -->
# Services Guide (dev-stack)

This section is auto-generated from `services/services.yaml`.

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
<!-- AUTO-GENERATED-END -->

## 🗂️ Overview

This guide describes all services available in **dev-stack**, their integration patterns, advanced configuration, and troubleshooting. For a full configuration schema, see [configuration.md](configuration.md).

The framework supports multiple services that can be mixed and matched based on your project needs. Each service is containerized and configured with sensible defaults for development environments.

## 🛠️ Integration Patterns

Refer to the auto-generated section above for a summary of available services, options, examples, and links.

## 📄 Database Services

### PostgreSQL

**Description**: PostgreSQL is the primary relational database option, offering excellent performance and feature set for development.

**Default Configuration**:
- **Port**: 5432
- **Database**: `{project_name}_dev`
- **Username**: `{project_name}_user`
- **Password**: Auto-generated
- **Version**: PostgreSQL 15

<!-- Service options and overrides are now covered in the auto-generated section above. -->

**Spring Boot Integration**:
```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/my_app_dev
    username: app_user
    password: dev-password
    driver-class-name: org.postgresql.Driver
  jpa:
    database-platform: org.hibernate.dialect.PostgreSQLDialect
    hibernate:
      ddl-auto: update
    show-sql: true
```

**Useful Commands**:
```bash
# Connect to PostgreSQL
./scripts/manage.sh connect postgres

# Run SQL commands
psql -h localhost -U app_user -d my_app_dev

# Backup database
./scripts/manage.sh backup postgres

# View logs
./scripts/manage.sh logs postgres
```

**Best Practices**:
- Use separate databases for different environments
- Enable query logging during development
- Use connection pooling in applications
- Regular backups before schema changes

### MySQL

**Description**: MySQL is an alternative relational database option, popular for web applications.

**Default Configuration**:
- **Port**: 3306
- **Database**: `{project_name}_dev`
- **Username**: `{project_name}_user`
- **Password**: Auto-generated
- **Version**: MySQL 8.0

**Configuration Options**:
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
    sql_mode: "STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO"
    innodb_buffer_pool_size: "256M"
```

**Spring Boot Integration**:
```yaml
spring:
  datasource:
    url: jdbc:mysql://localhost:3306/my_app_dev?useSSL=false&allowPublicKeyRetrieval=true
    username: app_user
    password: dev-password
    driver-class-name: com.mysql.cj.jdbc.Driver
  jpa:
    database-platform: org.hibernate.dialect.MySQL8Dialect
    hibernate:
      ddl-auto: update
```

**Useful Commands**:
```bash
# Connect to MySQL
./scripts/manage.sh connect mysql

# Run SQL commands
mysql -h localhost -u app_user -p my_app_dev

# Backup database
./scripts/manage.sh backup mysql
```

## 🚀 Caching Services

### Redis

**Description**: Redis is an in-memory data structure store used for caching, session storage, and real-time data processing.

**Default Configuration**:
- **Port**: 6379
- **Password**: Auto-generated
- **Memory Limit**: 256MB
- **Persistence**: Enabled (RDB snapshots)
- **Version**: Redis 7

**Configuration Options**:
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
      save 300 10
      save 60 10000
      tcp-keepalive 60
      timeout 300
```

**Spring Boot Integration**:
```yaml
spring:
  data:
    redis:
      host: localhost
      port: 6379
      password: dev-password
      timeout: 2000ms
      lettuce:
        pool:
          max-active: 8
          max-idle: 8
          min-idle: 0
```

**Java Configuration**:
```java
@Configuration
@EnableCaching
public class RedisConfig {
    
    @Bean
    public RedisTemplate<String, Object> redisTemplate(RedisConnectionFactory factory) {
        RedisTemplate<String, Object> template = new RedisTemplate<>();
        template.setConnectionFactory(factory);
        template.setKeySerializer(new StringRedisSerializer());
        template.setValueSerializer(new GenericJackson2JsonRedisSerializer());
        return template;
    }
    
    @Bean
    public CacheManager cacheManager(RedisConnectionFactory factory) {
        RedisCacheConfiguration config = RedisCacheConfiguration.defaultCacheConfig()
            .entryTtl(Duration.ofMinutes(60))
            .serializeKeysWith(RedisSerializationContext.SerializationPair
                .fromSerializer(new StringRedisSerializer()))
            .serializeValuesWith(RedisSerializationContext.SerializationPair
                .fromSerializer(new GenericJackson2JsonRedisSerializer()));
        
        return RedisCacheManager.builder(factory)
            .cacheDefaults(config)
            .build();
    }
}
```

**Useful Commands**:
```bash
# Connect to Redis CLI
./scripts/manage.sh connect redis

# Redis commands
redis-cli -h localhost -p 6379 -a dev-password
PING
KEYS *
GET mykey
SET mykey "value"
FLUSHALL
```

**Use Cases**:
- Session storage
- Application caching
- Rate limiting
- Real-time analytics
- Message queuing

## 📊 Observability Services

### Jaeger

**Description**: Jaeger is a distributed tracing system for monitoring and troubleshooting microservices-based architectures.

**Default Configuration**:
- **UI Port**: 16686
- **OTLP gRPC Port**: 4317
- **OTLP HTTP Port**: 4318
- **Memory Limit**: 256MB

**Configuration Options**:
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

**Spring Boot Integration**:
```yaml
management:
  tracing:
    enabled: true
    sampling:
      probability: 1.0
  otlp:
    tracing:
      endpoint: http://localhost:4318/v1/traces
```

**Dependencies**:
```gradle
dependencies {
    implementation 'io.micrometer:micrometer-tracing-bridge-otel'
    implementation 'io.opentelemetry:opentelemetry-exporter-otlp'
}
```

**Accessing Jaeger UI**:
- Open http://localhost:16686
- Search for traces by service name
- Analyze request flows and performance bottlenecks

### Prometheus

**Description**: Prometheus is a monitoring system and time series database for collecting and querying metrics.

**Default Configuration**:
- **Port**: 9090
- **Scrape Interval**: 15s
- **Memory Limit**: 256MB
- **Retention**: 15 days

**Configuration Options**:
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
      - job_name: 'redis'
        static_configs:
          - targets: ['redis:6379']
```

**Spring Boot Integration**:
```yaml
management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus
  metrics:
    export:
      prometheus:
        enabled: true
```

**Dependencies**:
```gradle
dependencies {
    implementation 'io.micrometer:micrometer-registry-prometheus'
    implementation 'org.springframework.boot:spring-boot-starter-actuator'
}
```

**Accessing Prometheus**:
- Open http://localhost:9090
- Query metrics using PromQL
- Set up alerts and dashboards

## ☁️ Cloud Services

### LocalStack

**Description**: LocalStack provides local AWS cloud stack emulation for development and testing.

**Default Configuration**:
- **Main Port**: 4566
- **Dashboard Port**: 8055
- **Services**: SQS (default)
- **Memory Limit**: 512MB

**Configuration Options**:
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
      - lambda
    
    # Auto-create SQS queues
    sqs_queues:
      - name: "user-events"
        visibility_timeout: 30
        message_retention_period: 1209600  # 14 days
        max_receive_count: 3
        dead_letter_queue: true  # Creates "user-events-dlq"
      - name: "notifications"
        dead_letter_queue: "custom-dlq-name"
    
    # Auto-create SNS topics
    sns_topics:
      - name: "user-notifications"
        display_name: "User Notifications"
        subscriptions:
          - protocol: "sqs"
            endpoint: "user-events"
            raw_message_delivery: true
    
    # Auto-create DynamoDB tables
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

**Spring Boot Integration**:
```yaml
cloud:
  aws:
    credentials:
      access-key: test
      secret-key: test
    region:
      static: us-east-1
    sqs:
      endpoint: http://localhost:4566
    sns:
      endpoint: http://localhost:4566
    dynamodb:
      endpoint: http://localhost:4566
```

**Dependencies**:
```gradle
dependencies {
    implementation 'org.springframework.cloud:spring-cloud-starter-aws'
    implementation 'org.springframework.cloud:spring-cloud-starter-aws-messaging'
    implementation 'com.amazonaws:aws-java-sdk-dynamodb'
    implementation 'com.amazonaws:aws-java-sdk-s3'
}
```

**Java Configuration Examples**:
```java
// SQS Configuration
@Configuration
public class SqsConfig {
    
    @Bean
    @Primary
    public AmazonSQSAsync amazonSQS() {
        return AmazonSQSAsyncClientBuilder.standard()
            .withEndpointConfiguration(new AwsClientBuilder.EndpointConfiguration(
                "http://localhost:4566", "us-east-1"))
            .withCredentials(new AWSStaticCredentialsProvider(
                new BasicAWSCredentials("test", "test")))
            .build();
    }
}

// DynamoDB Configuration
@Configuration
public class DynamoDbConfig {
    
    @Bean
    @Primary
    public AmazonDynamoDB amazonDynamoDB() {
        return AmazonDynamoDBClientBuilder.standard()
            .withEndpointConfiguration(new AwsClientBuilder.EndpointConfiguration(
                "http://localhost:4566", "us-east-1"))
            .withCredentials(new AWSStaticCredentialsProvider(
                new BasicAWSCredentials("test", "test")))
            .build();
    }
}
```

**Accessing LocalStack**:
- Dashboard: http://localhost:8055
- AWS CLI: `aws --endpoint-url=http://localhost:4566 sqs list-queues`

## 📨 Messaging Services

### Kafka

**Description**: Apache Kafka is a distributed streaming platform for building real-time data pipelines and streaming applications.

**Default Configuration**:
- **Kafka Port**: 9092
- **UI Port**: 8080
- **Zookeeper Port**: 2181
- **Memory Limit**: 1GB

**Configuration Options**:
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
        cleanup_policy: "compact"  # For key-based compaction
```

**Spring Boot Integration**:
```yaml
spring:
  kafka:
    bootstrap-servers: localhost:9092
    consumer:
      group-id: my-app-group
      auto-offset-reset: earliest
      key-deserializer: org.apache.kafka.common.serialization.StringDeserializer
      value-deserializer: org.apache.kafka.common.serialization.StringDeserializer
    producer:
      key-serializer: org.apache.kafka.common.serialization.StringSerializer
      value-serializer: org.apache.kafka.common.serialization.StringSerializer
```

**Dependencies**:
```gradle
dependencies {
    implementation 'org.springframework.kafka:spring-kafka'
    implementation 'org.apache.kafka:kafka-clients'
}
```

**Java Configuration Example**:
```java
@Service
public class EventService {
    
    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;
    
    public void publishEvent(String topic, String key, String message) {
        kafkaTemplate.send(topic, key, message);
    }
    
    @KafkaListener(topics = "user-events")
    public void handleUserEvent(String message) {
        log.info("Received user event: {}", message);
        // Process event
    }
    
    @KafkaListener(topics = "order-processing")
    public void handleOrderEvent(
        @Payload String message,
        @Header("eventType") String eventType) {
        log.info("Processing order event: {} (type: {})", message, eventType);
    }
}
```

**Accessing Kafka UI**:
- Open http://localhost:8080
- View topics, partitions, and messages
- Monitor consumer groups and lag

**Useful Commands**:
```bash
# List topics
kafka-topics --bootstrap-server localhost:9092 --list

# Create topic
kafka-topics --bootstrap-server localhost:9092 --create --topic my-topic --partitions 3

# Produce messages
kafka-console-producer --bootstrap-server localhost:9092 --topic my-topic

# Consume messages
kafka-console-consumer --bootstrap-server localhost:9092 --topic my-topic --from-beginning
```

## 🔗 Service Integration Patterns

### Database + Caching
```yaml
services:
  enabled:
    - postgres
    - redis

# Use Redis for caching database queries
@Cacheable(value = "users", key = "#id")
public User findById(Long id) {
    return userRepository.findById(id).orElse(null);
}
```

### Observability Stack
```yaml
services:
  enabled:
    - jaeger
    - prometheus

# Automatic tracing and metrics collection
# No additional code needed with Spring Boot Auto-configuration
```

### Event-Driven Architecture
```yaml
services:
  enabled:
    - postgres
    - kafka
    - redis

# Database for persistence, Kafka for events, Redis for caching
```

### Cloud-Native Development
```yaml
services:
  enabled:
    - postgres
    - redis
    - localstack
    - jaeger

# Full AWS-compatible development environment
```

## 📈 Performance Considerations

### Resource Usage by Service

| Service    | Default Memory | CPU Usage | Startup Time |
|------------|----------------|-----------|--------------|
| Redis      | 256MB         | Low       | Fast (~5s)   |
| PostgreSQL | 512MB         | Medium    | Medium (~15s)|
| MySQL      | 512MB         | Medium    | Medium (~20s)|
| Jaeger     | 256MB         | Low       | Fast (~10s)  |
| Prometheus | 256MB         | Low       | Medium (~15s)|
| LocalStack | 512MB         | High      | Slow (~30s)  |
| Kafka      | 1GB           | High      | Slow (~45s)  |

### Optimization Tips

1. **Memory Allocation**: Adjust memory limits based on your system
2. **Service Selection**: Only enable services you actually need
3. **Persistence**: Disable Redis persistence for faster development
4. **Parallel Startup**: Services start in parallel to reduce total time
5. **Image Caching**: Docker images are cached after first pull

## 🔧 Service Management

### Health Checks
```bash
# Check all services
./scripts/manage.sh status

# Service-specific health checks
curl http://localhost:16686/         # Jaeger UI
curl http://localhost:9090/          # Prometheus UI
curl http://localhost:4566/health    # LocalStack health
redis-cli -h localhost -p 6379 ping  # Redis
psql -h localhost -U postgres -c "SELECT 1"  # PostgreSQL
```

### Logs and Debugging
```bash
# View all logs
./scripts/manage.sh logs

# Service-specific logs
./scripts/manage.sh logs postgres -f
./scripts/manage.sh logs redis --tail=100
./scripts/manage.sh logs localstack
```

### Data Management
```bash
# Backup databases
./scripts/manage.sh backup postgres
./scripts/manage.sh backup mysql

# Clear Redis cache
./scripts/manage.sh exec redis redis-cli FLUSHALL

# Reset Kafka topics
./scripts/manage.sh exec kafka kafka-topics --bootstrap-server localhost:9092 --delete --topic my-topic
```

## 🆘 Troubleshooting Services

### Common Issues

**Service Won't Start**:
```bash
# Check logs
./scripts/manage.sh logs service-name

# Check port conflicts
lsof -i :PORT_NUMBER

# Recreate service
./scripts/manage.sh restart service-name
```

**Connection Refused**:
```bash
# Verify service is running
./scripts/manage.sh status

# Check network connectivity
telnet localhost PORT_NUMBER

# Verify configuration
./scripts/manage.sh info
```

**Performance Issues**:
```bash
# Monitor resource usage
./scripts/manage.sh monitor

# Check Docker stats
docker stats

# Adjust memory limits in configuration
```

## See Also
- [README](../README.md)
- [Configuration Guide](configuration.md)
- [Usage Guide](usage.md)
- [Troubleshooting Guide](troubleshooting.md)

- **[Configuration Guide](configuration.md)** - Advanced service configuration
- **[Usage Guide](usage.md)** - Daily management commands
- **[Integration Guide](integration.md)** - Application integration patterns
- **[Troubleshooting](troubleshooting.md)** - Detailed problem solving