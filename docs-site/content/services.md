# Available Services

This documentation is automatically generated from service definitions.


## Cache Services


### redis

**Description:** Redis in-memory data store for caching and session storage



**Configuration Options:**

- port

- password

- memory_limit

- persistence

- config



**Examples:**

- redis-cli -h localhost -p 6379 ping

- spring.data.redis.host=localhost



**Usage Notes:** Use Redis for caching, session storage, and pub/sub. Set a password for production-like security.

**Links:**

- https://redis.io/documentation

- https://spring.io/projects/spring-data-redis



---



## Cloud Services


### localstack-core

**Description:** LocalStack core AWS service emulator



**Configuration Options:**

- port

- dashboard_port

- memory_limit

- persistence



**Examples:**

- aws --endpoint-url=http://localhost:4566 s3 ls



**Usage Notes:** Core LocalStack service providing AWS API emulation for local development

**Links:**

- https://docs.localstack.cloud/



---


### localstack-dynamodb

**Description:** LocalStack DynamoDB NoSQL database emulation

**Dependencies:** localstack-core

**Configuration Options:**

- default_tables



**Examples:**

- aws --endpoint-url=http://localhost:4566 dynamodb list-tables



**Usage Notes:** DynamoDB NoSQL database emulation for local development

**Links:**

- https://docs.localstack.cloud/user-guide/aws/dynamodb/



---


### localstack-s3

**Description:** LocalStack S3 (Simple Storage Service) emulation

**Dependencies:** localstack-core

**Configuration Options:**

- default_buckets



**Examples:**

- aws --endpoint-url=http://localhost:4566 s3 ls



**Usage Notes:** S3 object storage emulation for local development and testing

**Links:**

- https://docs.localstack.cloud/user-guide/aws/s3/



---


### localstack-sns

**Description:** LocalStack SNS (Simple Notification Service) emulation

**Dependencies:** localstack-core

**Configuration Options:**

- default_topics



**Examples:**

- aws --endpoint-url=http://localhost:4566 sns list-topics



**Usage Notes:** SNS pub/sub notification service emulation. Can integrate with SQS for subscriptions.

**Links:**

- https://docs.localstack.cloud/user-guide/aws/sns/



---


### localstack-sqs

**Description:** LocalStack SQS (Simple Queue Service) emulation

**Dependencies:** localstack-core

**Configuration Options:**

- default_queues



**Examples:**

- aws --endpoint-url=http://localhost:4566 sqs list-queues



**Usage Notes:** SQS message queue emulation for local development

**Links:**

- https://docs.localstack.cloud/user-guide/aws/sqs/



---



## Database Services


### mysql

**Description:** MySQL relational database for persistent data storage



**Configuration Options:**

- port

- database

- username

- password

- memory_limit



**Examples:**

- mysql -h localhost -u root -p -e "SELECT version();"

- spring.datasource.url=jdbc:mysql://localhost:3306/my_app_dev



**Usage Notes:** MySQL database for relational data storage. Alternative to PostgreSQL.

**Links:**

- https://dev.mysql.com/doc/



---


### postgres

**Description:** PostgreSQL relational database for persistent data storage



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

- psql -h localhost -U postgres -c "SELECT version();"

- spring.datasource.url=jdbc:postgresql://localhost:5432/my_app_dev



**Usage Notes:** Ideal for structured data and transactional workloads. Use overrides to set custom database/user.

**Links:**

- https://www.postgresql.org/docs/

- https://spring.io/projects/spring-data-jpa



---



## Messaging Services


### kafka-broker

**Description:** Apache Kafka broker for event streaming and messaging

**Dependencies:** zookeeper

**Configuration Options:**

- port

- memory_limit

- auto_create_topics

- num_partitions

- replication_factor



**Examples:**

- kafka-broker-api-versions --bootstrap-server localhost:9092



**Usage Notes:** Kafka broker handles message storage and delivery. Requires Zookeeper for coordination.

**Links:**

- https://kafka.apache.org/documentation/

- https://docs.spring.io/spring-kafka/docs/current/reference/html/



---


### kafka-topics

**Description:** Kafka topic initialization and management service

**Dependencies:** kafka-broker

**Configuration Options:**

- default_topics_enabled



**Examples:**

- kafka-topics --create --bootstrap-server localhost:9092 --topic test --partitions 3 --replication-factor 1



**Usage Notes:** Initializes default Kafka topics for development. Runs once on startup.

**Links:**

- https://kafka.apache.org/documentation/#quickstart_createtopic



---


### kafka-ui

**Description:** Web UI for Kafka cluster management and topic browsing

**Dependencies:** kafka-broker

**Configuration Options:**

- port

- memory_limit



**Examples:**

- curl -f http://localhost:8080/actuator/health



**Usage Notes:** Web interface for managing Kafka topics, viewing messages, and monitoring cluster health

**Links:**

- https://github.com/provectus/kafka-ui



---


### zookeeper

**Description:** Apache Zookeeper coordination service for distributed systems



**Configuration Options:**

- port

- memory_limit



**Examples:**

- echo 'ruok' | nc localhost 2181



**Usage Notes:** Zookeeper provides coordination services for Kafka and other distributed systems

**Links:**

- https://zookeeper.apache.org/doc/current/



---



## Observability Services


### jaeger

**Description:** Jaeger distributed tracing system for monitoring and troubleshooting microservices



**Configuration Options:**

- ui_port

- otlp_http_port

- otlp_grpc_port

- memory_limit



**Examples:**

- curl -f http://localhost:16686/



**Usage Notes:** Distributed tracing for microservices. Supports OpenTelemetry and Zipkin protocols.

**Links:**

- https://www.jaegertracing.io/docs/

- https://opentelemetry.io/docs/



---


### prometheus

**Description:** Prometheus metrics collection and monitoring system



**Configuration Options:**

- port

- memory_limit

- retention_time



**Examples:**

- curl http://localhost:9090/metrics



**Usage Notes:** Metrics collection and monitoring. Scrapes application /metrics endpoints.

**Links:**

- https://prometheus.io/docs/



---




## Service Categories


- **Cache**: 1 service

- **Cloud**: 5 services

- **Database**: 2 services

- **Messaging**: 4 services

- **Observability**: 2 services


*Documentation generated automatically from service definitions*
