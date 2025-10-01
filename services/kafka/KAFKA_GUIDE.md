# Kafka Development Guide

This guide shows how to use Apache Kafka with the Local Development Framework for building event-driven applications.

## 🚀 Quick Start

### 1. Enable Kafka in Configuration

```yaml
# dev-stack-config.yaml
services:
  enabled:
    - redis
    - postgres
    - jaeger
    - kafka

overrides:
  kafka:
    auto_create_topics: true
    num_partitions: 1
    topics:
      - name: "user-events"
        partitions: 3
        replication_factor: 1
      - name: "order-processing"
        partitions: 6
        replication_factor: 1
      - name: "notifications"
        # Uses defaults: partitions: 1, replication_factor: 1
```

### 2. Start Services

```bash
./scripts/setup.sh
```

This will start:
- **Zookeeper** on port 2181 (Kafka coordination)
- **Kafka broker** on port 9092 (message streaming)
- **Kafka UI** on port 8080 (web management interface)

### 3. Access Kafka

- **Kafka UI**: http://localhost:8080
- **Bootstrap servers**: localhost:9092
- **Default topics**: `test`, `events`, `user-events`, `notifications`

## 📋 Spring Boot Integration

### Dependencies

Add to your `build.gradle`:

```gradle
dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.kafka:spring-kafka'
    implementation 'org.apache.kafka:kafka-clients'
}
```

### Configuration

The framework auto-generates this Spring Boot configuration:

```yaml
spring:
  kafka:
    bootstrap-servers: localhost:9092
    consumer:
      group-id: ${spring.application.name:local-app}
      auto-offset-reset: earliest
      key-deserializer: org.apache.kafka.common.serialization.StringDeserializer
      value-deserializer: org.apache.kafka.common.serialization.StringDeserializer
    producer:
      key-serializer: org.apache.kafka.common.serialization.StringSerializer
      value-serializer: org.apache.kafka.common.serialization.StringSerializer
```

## 🔧 Custom Topics Configuration

### Basic Topic Configuration

```yaml
overrides:
  kafka:
    topics:
      - name: "user-events"
        partitions: 3
        replication_factor: 1
      - name: "simple-topic"
        # Uses defaults: partitions: 1, replication_factor: 1
```

### Advanced Topic Configuration

```yaml
overrides:
  kafka:
    topics:
      - name: "order-events"
        partitions: 6
        replication_factor: 1
        cleanup_policy: "delete"
        retention_ms: 604800000 # 7 days
      - name: "user-profiles"
        partitions: 3
        replication_factor: 1
        cleanup_policy: "compact" # For key-based compaction
      - name: "metrics"
        partitions: 1
        replication_factor: 1
        retention_ms: 86400000 # 1 day
```

### Topic Properties

- **name** (required): Topic name
- **partitions**: Number of partitions (default: 1)
- **replication_factor**: Replication factor (default: 1, should not exceed broker count)
- **cleanup_policy**: Message cleanup policy - `delete`, `compact`, or `compact,delete` (default: delete)
- **retention_ms**: Message retention time in milliseconds (optional)
- **segment_ms**: Segment file time threshold in milliseconds (optional)

### Default vs Custom Topics

If no topics are configured, the framework creates these default topics:
- `test` (3 partitions)
- `events` (6 partitions)
- `user-events` (3 partitions)
- `notifications` (3 partitions)

When you specify custom topics, only your configured topics are created.

## 💻 Code Examples

### Basic Producer

```java
@Service
@Slf4j
public class EventPublisher {

    @Autowired
    private KafkaTemplate<String, String> kafkaTemplate;

    public void publishUserEvent(String userId, String action) {
        String message = String.format("{\"userId\":\"%s\",\"action\":\"%s\",\"timestamp\":\"%s\"}",
                                     userId, action, Instant.now());

        kafkaTemplate.send("user-events", userId, message)
            .addCallback(
                result -> log.info("Sent message=[{}] with offset=[{}]",
                                 message, result.getRecordMetadata().offset()),
                failure -> log.error("Unable to send message=[{}] due to : {}",
                                    message, failure.getMessage())
            );
    }
}
```

### Basic Consumer

```java
@Component
@Slf4j
public class EventConsumer {

    @KafkaListener(topics = "user-events")
    public void handleUserEvent(@Payload String message,
                               @Header(KafkaHeaders.RECEIVED_TOPIC) String topic,
                               @Header(KafkaHeaders.RECEIVED_PARTITION_ID) int partition,
                               @Header(KafkaHeaders.OFFSET) long offset) {

        log.info("Received: Topic=[{}] Partition=[{}] Offset=[{}] Message=[{}]",
                topic, partition, offset, message);

        // Process the event
        processUserEvent(message);
    }

    private void processUserEvent(String message) {
        // Your business logic here
        log.info("Processing user event: {}", message);
    }
}
```

### Advanced Configuration

```java
@Configuration
@EnableKafka
public class KafkaConfig {

    @Bean
    public ProducerFactory<String, Object> producerFactory() {
        Map<String, Object> configProps = new HashMap<>();
        configProps.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, "localhost:9092");
        configProps.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer.class);
        configProps.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, JsonSerializer.class);
        configProps.put(ProducerConfig.ACKS_CONFIG, "all"); // Wait for all replicas
        configProps.put(ProducerConfig.RETRIES_CONFIG, 3);
        return new DefaultKafkaProducerFactory<>(configProps);
    }

    @Bean
    public KafkaTemplate<String, Object> kafkaTemplate() {
        return new KafkaTemplate<>(producerFactory());
    }

    @Bean
    public ConsumerFactory<String, Object> consumerFactory() {
        Map<String, Object> props = new HashMap<>();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, "localhost:9092");
        props.put(ConsumerConfig.GROUP_ID_CONFIG, "my-app-group");
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, JsonDeserializer.class);
        props.put(JsonDeserializer.TRUSTED_PACKAGES, "*");
        return new DefaultKafkaConsumerFactory<>(props);
    }
}
```

## 🛠 Management Commands

### Using Framework Management Scripts

```bash
# Connect to Kafka console consumer
./scripts/manage.sh connect kafka

# List topics
./scripts/manage.sh exec kafka kafka-topics --list --bootstrap-server localhost:9092

# Create a new topic (manual)
./scripts/manage.sh exec kafka kafka-topics --create --bootstrap-server localhost:9092 --topic my-topic --partitions 6 --replication-factor 1

# Describe topic
./scripts/manage.sh exec kafka kafka-topics --describe --bootstrap-server localhost:9092 --topic my-topic

# Produce test messages
./scripts/manage.sh exec kafka kafka-console-producer --bootstrap-server localhost:9092 --topic my-topic

# Consume messages from beginning
./scripts/manage.sh exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic my-topic --from-beginning
```

### Direct Docker Commands

```bash
# List consumer groups
docker exec dev-stack-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list

# Check consumer group lag
docker exec dev-stack-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group my-app-group --describe

# Delete topic (be careful!)
docker exec dev-stack-kafka kafka-topics --delete --bootstrap-server localhost:9092 --topic my-topic
```

## 📊 Common Patterns

### Event Sourcing Pattern

```java
@Service
public class UserService {

    @Autowired
    private EventPublisher eventPublisher;

    @Autowired
    private UserRepository userRepository;

    @Transactional
    public User createUser(CreateUserRequest request) {
        // Create user
        User user = new User(request.getName(), request.getEmail());
        user = userRepository.save(user);

        // Publish event
        eventPublisher.publishUserEvent(user.getId(), "USER_CREATED");

        return user;
    }
}
```

### CQRS with Kafka

```java
// Command side
@RestController
public class UserCommandController {

    @PostMapping("/users")
    public ResponseEntity<String> createUser(@RequestBody CreateUserCommand command) {
        String eventId = UUID.randomUUID().toString();
        kafkaTemplate.send("user-commands", eventId, command);
        return ResponseEntity.accepted().body(eventId);
    }
}

// Query side
@Component
public class UserProjectionHandler {

    @KafkaListener(topics = "user-events")
    public void handleUserEvent(UserEvent event) {
        // Update read model/projection
        updateUserProjection(event);
    }
}
```

### Saga Pattern for Distributed Transactions

```java
@Component
public class OrderSagaOrchestrator {

    @KafkaListener(topics = "order-created")
    public void handleOrderCreated(OrderCreatedEvent event) {
        // Step 1: Reserve inventory
        kafkaTemplate.send("inventory-commands", new ReserveInventoryCommand(event.getOrderId()));
    }

    @KafkaListener(topics = "inventory-reserved")
    public void handleInventoryReserved(InventoryReservedEvent event) {
        // Step 2: Process payment
        kafkaTemplate.send("payment-commands", new ProcessPaymentCommand(event.getOrderId()));
    }

    @KafkaListener(topics = "payment-failed")
    public void handlePaymentFailed(PaymentFailedEvent event) {
        // Compensate: Release inventory
        kafkaTemplate.send("inventory-commands", new ReleaseInventoryCommand(event.getOrderId()));
    }
}
```

## 🔧 Configuration Options

### Framework Configuration

```yaml
# dev-stack-config.yaml
overrides:
  kafka:
    port: 9092                    # Kafka broker port
    ui_port: 8080                 # Kafka UI port
    zookeeper_port: 2181          # Zookeeper port
    memory_limit: "1024m"         # Memory limit for Kafka
    zookeeper_memory_limit: "256m" # Memory limit for Zookeeper
    auto_create_topics: true      # Auto-create topics when produced to
    num_partitions: 6             # Default partitions for new topics
    log_retention_hours: 168      # 7 days retention
    log_retention_bytes: 1073741824 # 1GB max log size
```

### Performance Tuning

For high-throughput scenarios:

```yaml
overrides:
  kafka:
    memory_limit: "2048m"
    environment:
      KAFKA_HEAP_OPTS: "-Xmx1024M -Xms512M"
      KAFKA_LOG_SEGMENT_BYTES: 536870912  # 512MB segments
      KAFKA_NUM_NETWORK_THREADS: 8
      KAFKA_NUM_IO_THREADS: 16
```

## 🐛 Troubleshooting

### Common Issues

**Kafka Won't Start**:
```bash
# Check Zookeeper is healthy first
./scripts/manage.sh logs zookeeper

# Check Kafka logs
./scripts/manage.sh logs kafka

# Restart services
./scripts/manage.sh restart
```

**Topics Not Auto-Created**:
```bash
# Create topic manually
./scripts/manage.sh exec kafka kafka-topics --create --bootstrap-server localhost:9092 --topic my-topic --partitions 3 --replication-factor 1
```

**Consumer Lag Issues**:
```bash
# Check consumer group status
docker exec dev-stack-kafka kafka-consumer-groups --bootstrap-server localhost:9092 --group my-group --describe
```

**Connection Refused**:
- Ensure Kafka is fully started (can take 60+ seconds)
- Check bootstrap servers configuration: `localhost:9092`
- Verify no port conflicts on 9092

### Monitoring

```bash
# Real-time monitoring
./scripts/manage.sh monitor

# Kafka-specific monitoring via UI
open http://localhost:8080

# Check topic metrics
curl http://localhost:8080/api/clusters/dev-stack/topics
```

## 📚 Best Practices

### Topic Design
- **Use meaningful names**: `user-events`, `order-processing`, `notifications`
- **Plan partitions**: More partitions = better parallelism, but more overhead
- **Consider retention**: Set appropriate retention based on use case

### Producer Best Practices
- **Use appropriate serializers**: String, JSON, Avro, etc.
- **Handle failures**: Implement retry logic and error handling
- **Batch messages**: Use batching for better throughput
- **Use keys wisely**: Keys determine partitioning

### Consumer Best Practices
- **Process idempotently**: Handle duplicate messages gracefully
- **Use appropriate group IDs**: Different services should use different groups
- **Handle errors**: Implement error handling and dead letter queues
- **Monitor lag**: Keep consumer lag low

### Development Tips
- **Start simple**: Begin with string messages, evolve to JSON/Avro
- **Use Kafka UI**: Great for debugging and topic management
- **Test locally**: Use framework's Kafka setup for development
- **Plan schema evolution**: Consider backward/forward compatibility

## 🎯 Production Considerations

When moving to production:

1. **External Kafka Cluster**: Use managed Kafka (Confluent Cloud, AWS MSK, etc.)
2. **Schema Registry**: Use Confluent Schema Registry or similar for schema management
3. **Monitoring**: Implement comprehensive monitoring (Prometheus + Grafana)
4. **Security**: Enable SASL/SSL authentication and encryption
5. **Backup**: Implement proper backup and disaster recovery
6. **Multi-broker**: Use multiple brokers with replication factor > 1

## 🔗 Additional Resources

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Spring for Apache Kafka](https://docs.spring.io/spring-kafka/docs/current/reference/html/)
- [Kafka UI Documentation](https://docs.kafka-ui.provectus.io/)
- [Kafka Design Patterns](https://kafka.apache.org/documentation/#design)
- [Event-Driven Architecture with Kafka](https://www.confluent.io/learn/event-driven-architecture/)

## 🎯 Topic Configuration Best Practices

### Partition Planning
- **Start small**: Begin with 1-3 partitions, scale as needed
- **Consider parallelism**: More partitions = more parallel consumers
- **Avoid over-partitioning**: Too many partitions can hurt performance

### Retention Policies
- **Delete policy**: Good for event streams, logs, metrics
- **Compact policy**: Good for key-based data, configuration, user profiles
- **Mixed policy**: `compact,delete` for both benefits

### Development vs Production
```yaml
# Development (local)
topics:
  - name: "user-events"
    partitions: 1        # Simple setup
    replication_factor: 1 # Single broker
    retention_ms: 86400000 # 1 day

# Production considerations
topics:
  - name: "user-events"
    partitions: 12       # Higher throughput
    replication_factor: 3 # Multiple brokers
    retention_ms: 604800000 # 7 days
```

### Topic Naming Conventions
- Use kebab-case: `user-events`, `order-processing`
- Be descriptive: `payment-confirmations` vs `payments`
- Include entity: `user-profile-updates`, `order-status-changes`
- Avoid spaces and special characters

Happy event streaming! 🚀
