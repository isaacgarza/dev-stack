---
title: "Integration Guide"
description: "Integrate dev-stack with applications, IDEs, and CI/CD pipelines"
lead: "Connect dev-stack with Spring Boot, testing frameworks, and development tools"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 40
toc: true
---

# Integration Guide

This guide covers integrating the Local Development Framework with Spring Boot applications, IDEs, testing frameworks, and CI/CD pipelines.

## 📋 Overview

### Quick Integration Patterns

- **Spring Boot:** Use generated configuration for datasource, Redis, Kafka, and AWS services.
- **CI/CD:** Use framework services in GitHub Actions, GitLab CI, or Docker Compose for integration tests.
- **IDE Integration:** Connect to framework databases, Redis, Kafka, and LocalStack from IntelliJ or VS Code.
- **Testcontainers:** Use framework images or connect to running services for integration tests.

The framework automatically generates configuration files and provides seamless integration with popular development tools and frameworks. This guide shows how to make the most of these integrations.

## 🍃 Spring Boot Integration

### Automatic Configuration Generation

The framework automatically generates `application-local.yml.generated` when Spring Boot projects are detected:

```yaml
# Generated application-local.yml
spring:
  profiles:
    active: local
  datasource:
    url: jdbc:postgresql://localhost:5432/my_app_dev
    username: app_user
    password: dev-password
    driver-class-name: org.postgresql.Driver
  data:
    redis:
      host: localhost
      port: 6379
      password: dev-password
      timeout: 2000ms
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true
    properties:
      hibernate:
        format_sql: true

management:
  tracing:
    enabled: true
    sampling:
      probability: 1.0
  otlp:
    tracing:
      endpoint: http://localhost:4318/v1/traces
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus

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

spring:
  kafka:
    bootstrap-servers: localhost:9092
    consumer:
      group-id: ${spring.application.name:my-app}
      auto-offset-reset: earliest
    producer:
      key-serializer: org.apache.kafka.common.serialization.StringSerializer
      value-serializer: org.apache.kafka.common.serialization.StringSerializer
```

### Using Generated Configuration

Copy the generated configuration to your application:

```bash
# Copy generated config to your application config
cp application-local.yml.generated src/main/resources/application-local.yml

# Or reference it directly
ln -s ../application-local.yml.generated src/main/resources/application-local.yml
```

### Required Dependencies

Add these dependencies to your `build.gradle` based on enabled services:

```gradle
dependencies {
    // Core Spring Boot
    implementation 'org.springframework.boot:spring-boot-starter-web'
    implementation 'org.springframework.boot:spring-boot-starter-actuator'

    // Database (choose one)
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa'
    runtimeOnly 'org.postgresql:postgresql'        // For PostgreSQL
    // runtimeOnly 'mysql:mysql-connector-java'   // For MySQL

    // Redis
    implementation 'org.springframework.boot:spring-boot-starter-data-redis'

    // Observability
    implementation 'io.micrometer:micrometer-tracing-bridge-otel'
    implementation 'io.opentelemetry:opentelemetry-exporter-otlp'
    implementation 'io.micrometer:micrometer-registry-prometheus'

    // AWS Services (LocalStack)
    implementation 'org.springframework.cloud:spring-cloud-starter-aws'
    implementation 'org.springframework.cloud:spring-cloud-starter-aws-messaging'
    implementation 'com.amazonaws:aws-java-sdk-dynamodb'

    // Kafka
    implementation 'org.springframework.kafka:spring-kafka'

    // Testing
    testImplementation 'org.springframework.boot:spring-boot-starter-test'
    testImplementation 'org.testcontainers:junit-jupiter'
    testImplementation 'org.testcontainers:postgresql'
    testImplementation 'org.testcontainers:kafka'
    testImplementation 'org.testcontainers:localstack'
}
```

### Application Configuration Examples

#### Database Configuration

```java
@Configuration
@EnableJpaRepositories
public class DatabaseConfig {

    @Bean
    @Primary
    @ConfigurationProperties("spring.datasource")
    public DataSource dataSource() {
        return DataSourceBuilder.create().build();
    }

    @Bean
    public JdbcTemplate jdbcTemplate(DataSource dataSource) {
        return new JdbcTemplate(dataSource);
    }
}
```

#### Redis Configuration

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
        template.setHashKeySerializer(new StringRedisSerializer());
        template.setHashValueSerializer(new GenericJackson2JsonRedisSerializer());
        template.afterPropertiesSet();
        return template;
    }

    @Bean
    public CacheManager cacheManager(RedisConnectionFactory factory) {
        RedisCacheConfiguration config = RedisCacheConfiguration.defaultCacheConfig()
            .entryTtl(Duration.ofHours(1))
            .serializeKeysWith(RedisSerializationContext.SerializationPair
                .fromSerializer(new StringRedisSerializer()))
            .serializeValuesWith(RedisSerializationContext.SerializationPair
                .fromSerializer(new GenericJackson2JsonRedisSerializer()))
            .disableCachingNullValues();

        return RedisCacheManager.builder(factory)
            .cacheDefaults(config)
            .transactionAware()
            .build();
    }
}
```

#### AWS Services Configuration

```java
@Configuration
public class AwsConfig {

    @Bean
    @Primary
    public AmazonSQS amazonSQS() {
        return AmazonSQSClientBuilder.standard()
            .withEndpointConfiguration(new AwsClientBuilder.EndpointConfiguration(
                "http://localhost:4566", "us-east-1"))
            .withCredentials(new AWSStaticCredentialsProvider(
                new BasicAWSCredentials("test", "test")))
            .build();
    }

    @Bean
    @Primary
    public AmazonSNS amazonSNS() {
        return AmazonSNSClientBuilder.standard()
            .withEndpointConfiguration(new AwsClientBuilder.EndpointConfiguration(
                "http://localhost:4566", "us-east-1"))
            .withCredentials(new AWSStaticCredentialsProvider(
                new BasicAWSCredentials("test", "test")))
            .build();
    }

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

#### Kafka Configuration

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
        props.put(ConsumerConfig.AUTO_OFFSET_RESET_CONFIG, "earliest");
        return new DefaultKafkaConsumerFactory<>(props);
    }

    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, Object> kafkaListenerContainerFactory() {
        ConcurrentKafkaListenerContainerFactory<String, Object> factory =
            new ConcurrentKafkaListenerContainerFactory<>();
        factory.setConsumerFactory(consumerFactory());
        return factory;
    }
}
```

### Service Integration Examples

#### Event Processing Service

```java
@Service
@Slf4j
public class EventService {

    @Autowired
    private KafkaTemplate<String, Object> kafkaTemplate;

    @Autowired
    private RedisTemplate<String, Object> redisTemplate;

    @Autowired
    private UserRepository userRepository;

    public void publishUserEvent(String userId, String eventType, Object eventData) {
        UserEvent event = new UserEvent(userId, eventType, eventData, Instant.now());
        kafkaTemplate.send("user-events", userId, event);
        log.info("Published user event: {} for user: {}", eventType, userId);
    }

    @KafkaListener(topics = "user-events")
    public void handleUserEvent(UserEvent event) {
        log.info("Processing user event: {}", event);

        // Update cache
        String cacheKey = "user:" + event.getUserId();
        redisTemplate.opsForValue().set(cacheKey, event, Duration.ofHours(1));

        // Update database
        User user = userRepository.findById(event.getUserId()).orElse(null);
        if (user != null) {
            user.setLastActivity(event.getTimestamp());
            userRepository.save(user);
        }
    }

    @Cacheable(value = "users", key = "#userId")
    public User getUserWithCache(String userId) {
        return userRepository.findById(userId).orElse(null);
    }
}
```

## 🧪 Testing Integration

### Integration Testing with Framework Services

#### Test Configuration

```yaml
# application-test.yml
spring:
  profiles:
    active: test
  datasource:
    url: jdbc:postgresql://localhost:5432/my_app_test
    username: test_user
    password: test-password
  data:
    redis:
      host: localhost
      port: 6379
      database: 1 # Use different Redis database for tests
  kafka:
    bootstrap-servers: localhost:9092
    consumer:
      group-id: test-group

cloud:
  aws:
    credentials:
      access-key: test
      secret-key: test
    sqs:
      endpoint: http://localhost:4566
```

#### Framework Integration Tests

```java
@SpringBootTest
@TestPropertySource(properties = {
    "spring.datasource.url=jdbc:postgresql://localhost:5432/my_app_test",
    "spring.data.redis.database=1",
    "spring.kafka.consumer.group-id=test-group"
})
class FrameworkIntegrationTest {

    @Autowired
    private UserService userService;

    @Autowired
    private RedisTemplate<String, Object> redisTemplate;

    @Autowired
    private KafkaTemplate<String, Object> kafkaTemplate;

    @Test
    void testDatabaseConnection() {
        User user = new User("test@example.com", "Test User");
        User saved = userService.save(user);
        assertThat(saved.getId()).isNotNull();
    }

    @Test
    void testRedisCache() {
        String key = "test:key";
        String value = "test-value";

        redisTemplate.opsForValue().set(key, value);
        String retrieved = (String) redisTemplate.opsForValue().get(key);

        assertThat(retrieved).isEqualTo(value);
    }

    @Test
    void testKafkaMessaging() throws InterruptedException {
        CountDownLatch latch = new CountDownLatch(1);

        @KafkaListener(topics = "test-topic")
        void handleTestMessage(String message) {
            assertThat(message).isEqualTo("test-message");
            latch.countDown();
        }

        kafkaTemplate.send("test-topic", "test-message");
        assertThat(latch.await(10, TimeUnit.SECONDS)).isTrue();
    }
}
```

#### Testcontainers Alternative

```java
@SpringBootTest
@Testcontainers
class ContainerizedIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:15-alpine")
            .withDatabaseName("test_db")
            .withUsername("test_user")
            .withPassword("test_password");

    @Container
    static GenericContainer<?> redis = new GenericContainer<>("redis:7-alpine")
            .withExposedPorts(6379);

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        registry.add("spring.data.redis.host", redis::getHost);
        registry.add("spring.data.redis.port", redis::getFirstMappedPort);
    }

    @Test
    void testWithContainers() {
        // Test logic using containerized services
    }
}
```

## 🏭 IDE Integration

### IntelliJ IDEA Setup

#### Database Integration

1. **Database Tool Window**: Connect to framework databases
   - Host: localhost
   - Port: 5432 (PostgreSQL) / 3306 (MySQL)
   - Database: Your configured database name
   - Username/Password: From framework configuration

2. **Redis Plugin**: Install Redis plugin for cache inspection
   - Host: localhost
   - Port: 6379
   - Password: From framework configuration

#### Run Configurations

```xml
<!-- IntelliJ Run Configuration -->
<configuration name="Application (Local)" type="SpringBootApplicationConfigurationType">
  <option name="SPRING_BOOT_MAIN_CLASS" value="com.example.Application" />
  <option name="ACTIVE_PROFILES" value="local" />
  <option name="VM_PARAMETERS" value="-Dspring.profiles.active=local" />
  <option name="PROGRAM_PARAMETERS" value="" />
  <option name="ALTERNATIVE_JRE_PATH_ENABLED" value="false" />
  <option name="ALTERNATIVE_JRE_PATH" />
  <option name="SHORTEN_COMMAND_LINE" value="NONE" />
  <option name="FRAME_DEACTIVATION_UPDATE_POLICY" value="UpdateClassesAndResources" />
</configuration>
```

#### Test Configuration

```xml
<configuration name="Integration Tests" type="JUnit" factoryName="JUnit">
  <option name="VM_PARAMETERS" value="-Dspring.profiles.active=test -Dtestcontainers.reuse.enable=true" />
  <option name="PROGRAM_PARAMETERS" value="" />
  <option name="ALTERNATIVE_JRE_PATH_ENABLED" value="false" />
  <option name="ALTERNATIVE_JRE_PATH" />
  <option name="PACKAGE_NAME" value="" />
  <option name="MAIN_CLASS_NAME" value="" />
  <option name="METHOD_NAME" value="" />
  <option name="TEST_OBJECT" value="package" />
  <option name="PARAMETERS" value="" />
</configuration>
```

### VS Code Setup

#### Settings Configuration

```json
{
  "spring-boot.ls.problem.application-properties.enabled": true,
  "java.test.config": {
    "vmArgs": [
      "-Dspring.profiles.active=test",
      "-Dspring.datasource.url=jdbc:postgresql://localhost:5432/my_app_test"
    ]
  },
  "java.debug.settings.onBuildFailureProceed": true,
  "java.compile.nullAnalysis.mode": "automatic"
}
```

#### Launch Configuration

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "java",
      "name": "Application (Local)",
      "request": "launch",
      "mainClass": "com.example.Application",
      "projectName": "my-app",
      "args": "--spring.profiles.active=local",
      "vmArgs": "-Dspring.profiles.active=local"
    },
    {
      "type": "java",
      "name": "Integration Tests",
      "request": "launch",
      "mainClass": "com.example.IntegrationTest",
      "projectName": "my-app",
      "vmArgs": "-Dspring.profiles.active=test"
    }
  ]
}
```

## 🚀 CI/CD Integration

### GitHub Actions

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up JDK 17
        uses: actions/setup-java@v3
        with:
          java-version: "17"
          distribution: "temurin"

      - name: Start Framework Services
        run: |
          dev-stack up

      - name: Wait for Services
        run: |
          timeout 60 bash -c 'until dev-stack status; do sleep 2; done'

      - name: Run Tests
        run: ./gradlew test integrationTest

      - name: Cleanup
        run: dev-stack cleanup
```

### Docker Compose for CI

```yaml
# docker-compose.ci.yml - Lightweight for CI
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: test
      POSTGRES_DB: ci_test
    ports:
      - "5432:5432"
    tmpfs:
      - /var/lib/postgresql/data # Use tmpfs for speed

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    tmpfs:
      - /data # Use tmpfs for speed
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test

integration-tests:
  stage: test
  image: openjdk:17-jdk
  services:
    - postgres:15-alpine
    - redis:7-alpine
  variables:
    POSTGRES_DB: ci_test
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: test
    SPRING_DATASOURCE_URL: jdbc:postgresql://postgres:5432/ci_test
    SPRING_DATA_REDIS_HOST: redis
  script:
    - ./gradlew test integrationTest
  artifacts:
    reports:
      junit: build/test-results/test/TEST-*.xml
```

## 🔄 Advanced Integration Patterns

### Health Check Integration

```java
@Component
public class FrameworkHealthIndicator implements HealthIndicator {

    @Autowired
    private RedisTemplate<String, Object> redisTemplate;

    @Autowired
    private DataSource dataSource;

    @Override
    public Health health() {
        HealthBuilder builder = Health.up();

        // Check Redis
        try {
            redisTemplate.opsForValue().get("health-check");
            builder.withDetail("redis", "UP");
        } catch (Exception e) {
            builder.down().withDetail("redis", "DOWN: " + e.getMessage());
        }

        // Check Database
        try (Connection connection = dataSource.getConnection()) {
            if (connection.isValid(5)) {
                builder.withDetail("database", "UP");
            } else {
                builder.down().withDetail("database", "DOWN: Invalid connection");
            }
        } catch (Exception e) {
            builder.down().withDetail("database", "DOWN: " + e.getMessage());
        }

        return builder.build();
    }
}
```

### Metrics Integration

```java
@Component
public class FrameworkMetrics {

    private final MeterRegistry meterRegistry;
    private final RedisTemplate<String, Object> redisTemplate;

    public FrameworkMetrics(MeterRegistry meterRegistry, RedisTemplate<String, Object> redisTemplate) {
        this.meterRegistry = meterRegistry;
        this.redisTemplate = redisTemplate;

        // Register custom metrics
        Gauge.builder("redis.connections")
            .description("Number of Redis connections")
            .register(meterRegistry, this, FrameworkMetrics::getRedisConnections);
    }

    private double getRedisConnections(FrameworkMetrics metrics) {
        try {
            Properties info = redisTemplate.getConnectionFactory()
                .getConnection()
                .info("clients");
            return Double.parseDouble(info.getProperty("connected_clients", "0"));
        } catch (Exception e) {
            return 0;
        }
    }
}
```

## 🧭 Next Steps

## 📚 See Also

- [README](../README.md)
- [Configuration Guide](configuration.md)
- [Services Guide](services.md)
- [Usage Guide](usage.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Contributing Guide](contributing.md)

- **[Configuration Guide](configuration.md)** - Advanced configuration options
- **[Services Guide](services.md)** - Detailed service information
- **[Usage Guide](usage.md)** - Daily commands and workflows
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
