# LocalStack Development Guide

This guide shows how to use LocalStack with the Local Development Framework for AWS service emulation with automatic SQS queue and SNS topic creation.

## 🚀 Quick Start

### 1. Enable LocalStack in Configuration

```yaml
# local-dev-config.yaml
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
      - dynamodb
    sqs_queues:
      - name: "user-events"
        visibility_timeout: 30
        dead_letter_queue: true # Creates "user-events-dlq"
      - name: "notifications"
        visibility_timeout: 60
        dead_letter_queue: true # Creates "notifications-dlq"
    sns_topics:
      - name: "user-notifications"
        display_name: "User Notifications"
        subscriptions:
          - protocol: "sqs"
            endpoint: "user-events"
            raw_message_delivery: true
          - protocol: "sqs"
            endpoint: "notifications"
            raw_message_delivery: false
      dynamodb_tables:
        - name: "event-store"
          attribute_definitions:
            - AttributeName: "aggregateId"
              AttributeType: "S"
            - AttributeName: "version"
              AttributeType: "N"
          key_schema:
            - AttributeName: "aggregateId"
              KeyType: "HASH"
            - AttributeName: "version"
              KeyType: "RANGE"
          provisioned_throughput:
            ReadCapacityUnits: 10
            WriteCapacityUnits: 10
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

### 2. Start Services

```bash
./scripts/setup.sh
```

This will start:
- **LocalStack** on port 4566 (AWS API endpoint)
- **LocalStack Dashboard** on port 8055 (web management interface)
- **Auto-create** your configured SQS queues, SNS topics, and DynamoDB tables

### 3. Access LocalStack

- **AWS Endpoint**: http://localhost:4566
- **Dashboard**: http://localhost:8055
- **AWS CLI**: `aws --endpoint-url=http://localhost:4566 --region=us-east-1`

## 📋 Spring Boot Integration

### Dependencies

Add to your `build.gradle`:

```gradle
dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.cloud:spring-cloud-starter-aws'
    implementation 'org.springframework.cloud:spring-cloud-starter-aws-messaging'
    implementation 'com.amazonaws:aws-java-sdk-dynamodb'
}
```

### Configuration

The framework auto-generates this Spring Boot configuration:

```yaml
cloud:
  aws:
    credentials:
      access-key: test
      secret-key: test
    region:
      static: us-east-1
    stack:
      auto: false
    sqs:
      endpoint: http://localhost:4566
    sns:
      endpoint: http://localhost:4566
```

## 🔧 SQS Queue Configuration

### Basic Queue Configuration

```yaml
overrides:
  localstack:
    sqs_queues:
      - name: "simple-queue"
      - name: "user-events"
        visibility_timeout: 30
        message_retention_period: 1209600 # 14 days
        receive_message_wait_time: 0 # Long polling disabled
        delay_seconds: 0 # No delivery delay
```

### Queue with Dead Letter Queue

```yaml
sqs_queues:
  # Automatic DLQ creation with boolean flag
  - name: "orders"
    visibility_timeout: 60
    max_receive_count: 3
    dead_letter_queue: true # Creates "orders-dlq" automatically
  
  # Custom DLQ name
  - name: "notifications"
    visibility_timeout: 30
    max_receive_count: 5
    dead_letter_queue: "custom-dlq-name" # Custom DLQ name
  
  # No DLQ (default behavior)
  - name: "simple-queue"
    visibility_timeout: 30
    # No dead_letter_queue specified
  
  # Explicitly disable DLQ
  - name: "no-dlq-queue"
    visibility_timeout: 30
    dead_letter_queue: false # No DLQ created
```

### Queue Properties

- **name** (required): Queue name
- **visibility_timeout**: Time in seconds message is invisible after being received (default: 30)
- **message_retention_period**: How long to keep messages in seconds (default: 14 days)
- **receive_message_wait_time**: Long polling wait time in seconds (default: 0)
- **delay_seconds**: Delay before message becomes available (default: 0)
- **max_receive_count**: Max receives before moving to DLQ (default: 3)
- **dead_letter_queue**: 
  - `true`: Auto-create DLQ as `{{queue-name}}-dlq`
  - `"custom-name"`: Create DLQ with custom name
  - `false` or omitted: No DLQ created

## 🗃️ DynamoDB Table Configuration

### Basic Table Configuration

```yaml
overrides:
  localstack:
    services:
      - dynamodb
    dynamodb_tables:
      - name: "users"
        attribute_definitions:
          - AttributeName: "userId"
            AttributeType: "S"
        key_schema:
          - AttributeName: "userId"
            KeyType: "HASH"
        provisioned_throughput:
          ReadCapacityUnits: 5
          WriteCapacityUnits: 5
```

### Table with Global Secondary Index

```yaml
dynamodb_tables:
  - name: "orders"
    attribute_definitions:
      - AttributeName: "orderId"
        AttributeType: "S"
      - AttributeName: "customerId"
        AttributeType: "S"
      - AttributeName: "status"
        AttributeType: "S"
    key_schema:
      - AttributeName: "orderId"
        KeyType: "HASH"
    provisioned_throughput:
      ReadCapacityUnits: 10
      WriteCapacityUnits: 10
    global_secondary_indexes:
      - IndexName: "CustomerIdIndex"
        KeySchema:
          - AttributeName: "customerId"
            KeyType: "HASH"
          - AttributeName: "status"
            KeyType: "RANGE"
        Projection:
          ProjectionType: "ALL"
        ProvisionedThroughput:
          ReadCapacityUnits: 5
          WriteCapacityUnits: 5
```

### Table Properties

- **name** (required): Table name
- **attribute_definitions** (required): Array of attribute definitions
  - **AttributeName**: Attribute name
  - **AttributeType**: S (String), N (Number), or B (Binary)
- **key_schema** (required): Array defining partition and sort keys
  - **AttributeName**: Attribute name (must be in attribute_definitions)
  - **KeyType**: HASH (partition key) or RANGE (sort key)
- **provisioned_throughput** (required): Read/write capacity units
  - **ReadCapacityUnits**: Read capacity (default: 5)
  - **WriteCapacityUnits**: Write capacity (default: 5)
- **global_secondary_indexes** (optional): Array of GSI definitions
- **table_class** (optional): STANDARD or STANDARD_INFREQUENT_ACCESS (default: STANDARD)

## 📢 SNS Topic Configuration

### Basic Topic Configuration

```yaml
overrides:
  localstack:
    sns_topics:
      - name: "notifications"
        display_name: "Application Notifications"
      - name: "user-events"
        display_name: "User Events"
```

### Topic with SQS Subscriptions

```yaml
sns_topics:
  - name: "order-events"
    display_name: "Order Events"
    subscriptions:
      - protocol: "sqs"
        endpoint: "order-processing"
        raw_message_delivery: true
      - protocol: "sqs"
        endpoint: "order-analytics"
        raw_message_delivery: false
      - protocol: "sqs"
        endpoint: "order-notifications"
        raw_message_delivery: true
```

### Topic with Multiple Subscription Types

```yaml
sns_topics:
  - name: "system-alerts"
    display_name: "System Alerts"
    subscriptions:
      - protocol: "sqs"
        endpoint: "alert-queue"
        raw_message_delivery: true
      - protocol: "email"
        endpoint: "admin@company.com"
        raw_message_delivery: false
      - protocol: "http"
        endpoint: "https://webhook.site/your-webhook"
        raw_message_delivery: false
```

### Subscription Properties

- **protocol** (required): `sqs`, `http`, `https`, `email`, `sms`, or `lambda`
- **endpoint** (required): 
  - SQS: Queue name (must exist in `sqs_queues`)
  - HTTP/HTTPS: Full URL
  - Email: Email address
- **raw_message_delivery**: 
  - `true`: Deliver message content directly
  - `false`: Wrap in SNS JSON envelope (default for non-SQS)
- **filter_policy**: JSON object for message filtering (optional)

## 💻 Code Examples

### SQS Producer (Spring Cloud AWS)

```java
@Service
@Slf4j
public class MessageProducer {
    
    @Autowired
    private QueueMessagingTemplate queueMessagingTemplate;
    
    public void sendMessage(String queueName, Object message) {
        try {
            queueMessagingTemplate.convertAndSend(queueName, message);
            log.info("Message sent to queue: {}", queueName);
        } catch (Exception e) {
            log.error("Failed to send message to queue: {}", queueName, e);
        }
    }
    
    public void sendUserEvent(String userId, String action) {
        UserEvent event = UserEvent.builder()
            .userId(userId)
            .action(action)
            .timestamp(Instant.now())
            .build();
            
        sendMessage("user-events", event);
    }
}
```

### SQS Consumer (Spring Cloud AWS)

```java
@Component
@Slf4j
public class MessageConsumer {
    
    @SqsListener("user-events")
    public void handleUserEvent(UserEvent event, 
                               @Header("SenderId") String senderId) {
        log.info("Received user event: {} for user: {}", event.getAction(), event.getUserId());
        
        try {
            processUserEvent(event);
        } catch (Exception e) {
            log.error("Failed to process user event: {}", event, e);
            throw e; // Will be retried and eventually sent to DLQ
        }
    }
    
    @SqsListener("notifications")
    public void handleNotification(NotificationMessage message) {
        log.info("Received notification: {}", message.getContent());
        sendNotification(message);
    }
    
    // Handle messages from dead letter queue
    @SqsListener("user-events-dlq")
    public void handleFailedUserEvent(UserEvent event) {
        log.error("Processing failed user event from DLQ: {}", event);
        // Handle failed messages (alert, manual processing, etc.)
        alertOperations("Failed to process user event", event);
    }
    
    private void processUserEvent(UserEvent event) {
        // Your business logic here
        switch (event.getAction()) {
            case "USER_CREATED":
                handleUserCreated(event.getUserId());
                break;
            case "USER_UPDATED":
                handleUserUpdated(event.getUserId());
                break;
            default:
                log.warn("Unknown user event action: {}", event.getAction());
        }
    }
}
```

### DynamoDB Operations (AWS SDK)

```java
@Service
@Slf4j
public class UserService {
    
    @Autowired
    private DynamoDBMapper dynamoDBMapper;
    
    public void saveUser(User user) {
        try {
            dynamoDBMapper.save(user);
            log.info("User saved to DynamoDB: {}", user.getUserId());
        } catch (Exception e) {
            log.error("Failed to save user to DynamoDB", e);
        }
    }
    
    public User getUser(String userId) {
        return dynamoDBMapper.load(User.class, userId);
    }
    
    public List<User> getUsersByEmail(String email) {
        DynamoDBScanExpression scanExpression = new DynamoDBScanExpression()
            .withFilterExpression("email = :email")
            .withExpressionAttributeValues(Map.of(":email", new AttributeValue().withS(email)));
        
        return dynamoDBMapper.scan(User.class, scanExpression);
    }
}

@DynamoDBTable(tableName = "users")
public class User {
    
    @DynamoDBHashKey
    private String userId;
    
    @DynamoDBIndexHashKey(globalSecondaryIndexName = "EmailIndex")
    private String email;
    
    private String name;
    private Instant createdTime;
    
    // getters and setters
}
```

### SNS Publisher (Spring Cloud AWS)

```java
@Service
@Slf4j
public class NotificationPublisher {
    
    @Autowired
    private NotificationMessagingTemplate notificationMessagingTemplate;
    
    public void publishNotification(String topicName, Object message) {
        try {
            notificationMessagingTemplate.convertAndSend(topicName, message);
            log.info("Message published to topic: {}", topicName);
        } catch (Exception e) {
            log.error("Failed to publish message to topic: {}", topicName, e);
        }
    }
    
    public void publishUserNotification(String userId, String message) {
        UserNotification notification = UserNotification.builder()
            .userId(userId)
            .message(message)
            .timestamp(Instant.now())
            .type("USER_NOTIFICATION")
            .build();
            
        publishNotification("user-notifications", notification);
    }
    
    public void publishSystemAlert(String level, String message) {
        SystemAlert alert = SystemAlert.builder()
            .level(level)
            .message(message)
            .timestamp(Instant.now())
            .source("application")
            .build();
            
        // This will be delivered to all subscribers of system-alerts topic
        publishNotification("system-alerts", alert);
    }
}
```

### Configuration Class

```java
@Configuration
@EnableSqs
@EnableSns
public class AwsConfig {
    
    @Bean
    @Primary
    public AmazonSQSAsync amazonSQS() {
        return AmazonSQSAsyncClientBuilder.standard()
            .withEndpointConfiguration(
                new AwsClientBuilder.EndpointConfiguration(
                    "http://localhost:4566", "us-east-1"))
            .withCredentials(new AWSStaticCredentialsProvider(
                new BasicAWSCredentials("test", "test")))
            .build();
    }
    
    @Bean
    @Primary
    public AmazonSNSAsync amazonSNS() {
        return AmazonSNSAsyncClientBuilder.standard()
            .withEndpointConfiguration(
                new AwsClientBuilder.EndpointConfiguration(
                    "http://localhost:4566", "us-east-1"))
            .withCredentials(new AWSStaticCredentialsProvider(
                new BasicAWSCredentials("test", "test")))
            .build();
    }
}
```

## 🛠 Management Commands

### Using Framework Management Scripts

```bash
# View LocalStack status
./scripts/manage.sh status

# View LocalStack logs
./scripts/manage.sh logs localstack

# Connect to LocalStack container
./scripts/manage.sh exec localstack bash
```

### AWS CLI Commands

```bash
# Set up AWS CLI alias for LocalStack
alias awslocal="aws --endpoint-url=http://localhost:4566 --region=us-east-1"

# List SQS queues
awslocal sqs list-queues

# Send message to queue
awslocal sqs send-message \
  --queue-url http://localhost:4566/000000000000/user-events \
  --message-body '{"userId":"123","action":"USER_CREATED"}'

# Receive messages from queue
awslocal sqs receive-message \
  --queue-url http://localhost:4566/000000000000/user-events

# List SNS topics
awslocal sns list-topics

# Publish to SNS topic
awslocal sns publish \
  --topic-arn arn:aws:sns:us-east-1:000000000000:user-notifications \
  --message '{"userId":"123","message":"Welcome!"}'

# List subscriptions
awslocal sns list-subscriptions

# List DynamoDB tables
awslocal dynamodb list-tables

# Describe DynamoDB table
awslocal dynamodb describe-table --table-name users

# Put item in DynamoDB
awslocal dynamodb put-item \
  --table-name users \
  --item '{"userId": {"S": "user123"}, "email": {"S": "user@example.com"}, "name": {"S": "John Doe"}}'

# Get item from DynamoDB
awslocal dynamodb get-item \
  --table-name users \
  --key '{"userId": {"S": "user123"}}'

# Scan DynamoDB table
awslocal dynamodb scan --table-name users

# Query DynamoDB with GSI
awslocal dynamodb query \
  --table-name users \
  --index-name EmailIndex \
  --key-condition-expression "email = :email" \
  --expression-attribute-values '{":email": {"S": "user@example.com"}}'

# Delete item from DynamoDB
awslocal dynamodb delete-item \
  --table-name users \
  --key '{"userId": {"S": "user123"}}'
```

### Docker Commands for Debugging

```bash
# Check LocalStack logs
docker logs local-dev-localstack

# Execute commands in LocalStack container
docker exec -it local-dev-localstack bash

# Check LocalStack health
# Check DynamoDB tables
awslocal dynamodb list-tables

# View LocalStack dashboard
open http://localhost:8055
```

## 📊 Common Patterns

### Event-Driven Architecture with SQS/SNS

```yaml
# Configuration for event-driven system
overrides:
  localstack:
    sqs_queues:
      - name: "user-command-queue"
      - name: "user-event-queue"
      - name: "email-queue"
      - name: "analytics-queue"
      - name: "audit-queue"
    sns_topics:
      - name: "user-events"
        subscriptions:
          - protocol: "sqs"
            endpoint: "user-event-queue"
            raw_message_delivery: true
          - protocol: "sqs"
            endpoint: "analytics-queue"
            raw_message_delivery: true
          - protocol: "sqs"
            endpoint: "audit-queue"
            raw_message_delivery: true
      - name: "notifications"
        subscriptions:
          - protocol: "sqs"
            endpoint: "email-queue"
            raw_message_delivery: false
```

### Microservices Communication

```java
// Order Service publishes events
@Service
public class OrderService {
    
    @Autowired
    private NotificationPublisher publisher;
    
    public Order createOrder(CreateOrderRequest request) {
        Order order = new Order(request);
        order = orderRepository.save(order);
        
        // Publish event to SNS topic
        publisher.publishNotification("order-events", 
            OrderCreatedEvent.builder()
                .orderId(order.getId())
                .customerId(order.getCustomerId())
                .amount(order.getAmount())
                .build());
        
        return order;
    }
}

// Inventory Service consumes events
@Component
public class InventoryEventHandler {
    
    @SqsListener("inventory-queue")
    public void handleOrderCreated(OrderCreatedEvent event) {
        // Reserve inventory for the order
        inventoryService.reserveItems(event.getOrderId(), event.getItems());
    }
}

// Email Service consumes events
@Component
public class EmailEventHandler {
    
    @SqsListener("email-queue")
    public void handleOrderCreated(OrderCreatedEvent event) {
        // Send confirmation email
        emailService.sendOrderConfirmation(event.getCustomerId(), event.getOrderId());
    }
}
```

### Dead Letter Queue Processing

```java
@Component
@Slf4j
public class DeadLetterQueueHandler {
    
    // DLQ created automatically with dead_letter_queue: true
    @SqsListener("user-events-dlq")
    public void handleFailedUserEvents(UserEvent event, 
                                     @Header Map<String, Object> headers) {
        
        String approximateReceiveCount = (String) headers.get("ApproximateReceiveCount");
        log.error("Processing failed user event (receive count: {}): {}", 
                  approximateReceiveCount, event);
        
        // Alert operations team
        alertService.sendAlert("DLQ Processing", 
            "Failed to process user event after " + approximateReceiveCount + " attempts", 
            event);
        
        // Optionally try manual processing or store for later analysis
        failedEventRepository.save(FailedEvent.from(event, headers));
    }
    
    // DLQ created automatically with dead_letter_queue: true
    @SqsListener("orders-dlq")
    public void handleFailedOrders(OrderEvent event) {
        // Critical business process - might need immediate attention
        alertService.sendCriticalAlert("Order Processing Failed", event);
        
        // Try alternative processing path
        orderRecoveryService.processFailedOrder(event);
    }
}
```

## 🔧 Advanced Configuration

### Queue with Custom Attributes

```yaml
sqs_queues:
  - name: "priority-queue"
    visibility_timeout: 120
    message_retention_period: 604800 # 7 days
    receive_message_wait_time: 20 # Long polling
    delay_seconds: 30 # 30-second delay
    max_receive_count: 5
    dead_letter_queue: true # Creates "priority-queue-dlq"
```

### Topic with Message Filtering

```yaml
sns_topics:
  - name: "filtered-notifications"
    subscriptions:
      - protocol: "sqs"
        endpoint: "high-priority-queue"
        raw_message_delivery: true
        filter_policy: '{"priority": ["high", "critical"]}'
      - protocol: "sqs"
        endpoint: "low-priority-queue"
        raw_message_delivery: true
        filter_policy: '{"priority": ["low", "medium"]}'
```

### Multiple Environment Setup

```yaml
# Development environment
overrides:
  localstack:
    services: [sqs, sns]
    sqs_queues:
      - name: "dev-user-events"
        dead_letter_queue: true # Creates "dev-user-events-dlq"
      - name: "dev-notifications"
        dead_letter_queue: true # Creates "dev-notifications-dlq"
    sns_topics:
      - name: "dev-alerts"
        subscriptions:
          - protocol: "sqs"
            endpoint: "dev-user-events"
            raw_message_delivery: true
```

## 🐛 Troubleshooting

### Common Issues

**LocalStack Won't Start**:
```bash
# Check Docker daemon is running
docker info

# Check LocalStack logs
./scripts/manage.sh logs localstack

# Check if ports are available
lsof -i :4566
lsof -i :8055
```

**Queues/Topics Not Created**:
```bash
# Check initialization logs
docker logs local-dev-localstack-init

# Verify configuration file
cat .localstack-config.json

# Manually run initialization
docker exec -it local-dev-localstack bash
/etc/localstack/init/ready.d/init-aws-resources.sh
```

**Messages Not Being Delivered**:
```bash
# Check queue attributes
awslocal sqs get-queue-attributes \
  --queue-url http://localhost:4566/000000000000/your-queue \
  --attribute-names All

# Check subscription status
awslocal sns list-subscriptions-by-topic \
  --topic-arn arn:aws:sns:us-east-1:000000000000:your-topic
```

**Permission Denied Errors**:
- LocalStack uses test credentials (`test`/`test`)
- No IAM policies are enforced in LocalStack
- Check endpoint URLs are correct (`http://localhost:4566`)

**DynamoDB Table Not Found**:
```bash
# Check if tables were created
awslocal dynamodb list-tables

# Check table creation logs
docker logs local-dev-localstack-init

# Manually create table if needed
awslocal dynamodb create-table \
  --table-name test-table \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

### Monitoring and Debugging

```bash
# Real-time LocalStack logs
./scripts/manage.sh logs localstack --follow

# Check LocalStack health
curl http://localhost:4566/health | jq

# List all SQS queues with details
awslocal sqs list-queues | jq '.QueueUrls[]' | while read queue; do
  echo "Queue: $queue"
  awslocal sqs get-queue-attributes --queue-url "$queue" --attribute-names All
done

# Monitor queue depth
awslocal sqs get-queue-attributes \
  --queue-url http://localhost:4566/000000000000/user-events \
  --attribute-names ApproximateNumberOfMessages
```

## 📚 Best Practices

### Queue Design
- **Use descriptive names**: `user-events`, `order-processing`, `email-notifications`
- **Plan for failure**: Always configure dead letter queues for important processes
- **Set appropriate timeouts**: Match visibility timeout to processing time
- **Use long polling**: Set `receive_message_wait_time` > 0 to reduce empty receives

### Topic Design
- **Fan-out pattern**: Use SNS topics to broadcast events to multiple services
- **Raw message delivery**: Use `true` for SQS subscribers to avoid JSON wrapping
- **Message filtering**: Use filter policies to route messages efficiently
- **Idempotent processing**: Design consumers to handle duplicate messages

### Error Handling
- **Implement retries**: Use exponential backoff for transient failures
- **Monitor DLQs**: Set up alerts for messages in dead letter queues
- **Log thoroughly**: Include correlation IDs and context in all log messages
- **Test failure scenarios**: Simulate network issues and processing failures

### Development Tips
- **Start simple**: Begin with basic queues and topics, add complexity gradually
- **Use LocalStack dashboard**: Great for debugging and monitoring
- **Test locally**: Use framework's LocalStack setup for development
- **Document queue purposes**: Maintain clear documentation of what each queue handles

## 🎯 Production Considerations

When moving to production:

1. **Use AWS SQS/SNS**: Replace LocalStack with actual AWS services
2. **IAM Policies**: Implement proper access controls and permissions
3. **Monitoring**: Set up CloudWatch metrics and alarms
4. **Encryption**: Enable encryption at rest and in transit
5. **Cross-region**: Consider cross-region replication for high availability
6. **Cost optimization**: Monitor usage and optimize queue/topic configurations

## 🔗 Additional Resources

- [LocalStack Documentation](https://docs.localstack.cloud/)
- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [AWS SNS Documentation](https://docs.aws.amazon.com/sns/)
- [Spring Cloud AWS](https://docs.awspring.io/spring-cloud-aws/docs/current/reference/html/)
- [Event-Driven Architecture Patterns](https://aws.amazon.com/event-driven-architecture/)

Happy cloud development! ☁️🚀