#!/usr/bin/env bash

# Local Development Framework - Management Script
# This script provides ongoing management operations for the development environment

set -e

# Determine script directory and source common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

# Show status of all services
show_status() {
    print_header "📊 Service Status"

    load_environment
    local cmd=$(get_compose_cmd)

    if $cmd ps | grep -q "Up"; then
        print_success "Development environment is running"

        local running_count=$($cmd ps --services | wc -l | tr -d ' ')
        local healthy_count=$($cmd ps --format "{{.Status}}" | grep -c "healthy" || echo "0")
        local starting_count=$($cmd ps --format "{{.Status}}" | grep -c "Up" || echo "0")
        starting_count=$((starting_count - healthy_count))

        print_sub_info "$running_count services active"
        if [ "$healthy_count" -gt 0 ] || [ "$starting_count" -gt 0 ]; then
            print_sub_info "$healthy_count healthy, $starting_count starting"
        fi
        print_section_break

        $cmd ps
        print_section_break

        print_step "Health Status"
        # Check service health
        $cmd ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" | tail -n +2 | while IFS=$'\t' read -r name status ports; do
            if [[ "$status" == *"healthy"* ]]; then
                print_sub_success "$name: Healthy"
            elif [[ "$status" == *"Up"* ]]; then
                print_sub_info "$name: Starting/No health check"
            else
                echo -e "   ${RED}❌${NC} $name: $status"
            fi
        done

    else
        print_warning "Development environment is not running"
        print_info "Use './manage.sh start' to start services"
    fi
}

# Check for existing instances before starting
check_existing_instances() {
    print_step "Checking for conflicts..."

    local existing_containers=($(find_framework_containers "running"))
    local conflicting_ports=()

    # Check for port conflicts with common framework ports
    local ports_to_check=(6379 5432 3306 16686 9090 9092 4566 8055 8080 2181)
    for port in "${ports_to_check[@]}"; do
        if check_port_in_use "$port"; then
            conflicting_ports+=("$port")
        fi
    done

    # If no conflicts found, continue normally
    if [ ${#existing_containers[@]} -eq 0 ] && [ ${#conflicting_ports[@]} -eq 0 ]; then
        print_success "No conflicts detected"
        print_section_break
        return 0
    fi

    # Display detected conflicts
    print_warning "Existing framework instances detected!"
    print_section_break

    if [ ${#existing_containers[@]} -gt 0 ]; then
        print_info "Running containers:"
        for container in "${existing_containers[@]}"; do
            local status=$(docker ps --filter "name=$container" --format "{{.Status}}" 2>/dev/null || echo "Unknown")
            print_sub_info "$container ($status)"
        done
        print_section_break
    fi

    if [ ${#conflicting_ports[@]} -gt 0 ]; then
        print_info "Ports in use:"
        for port in "${conflicting_ports[@]}"; do
            local process=$(get_port_process "$port")
            print_sub_info "Port $port (used by $process)"
        done
        print_section_break
    fi

    # Prompt user for action
    echo -e "${BOLD}What would you like to do?${NC}"
    echo "1) Stop existing instances and start fresh"
    echo "2) Cancel start operation"
    echo ""
    echo -n "Choose an option (1-2): "

    local choice
    read -r choice

    case $choice in
        1)
            print_step "Stopping existing instances..."
            stop_existing_instances
            print_success "Existing instances stopped"
            print_section_break
            ;;
        2)
            print_info "Start operation cancelled"
            exit 0
            ;;
        *)
            print_error "Invalid choice. Please select 1 or 2."
            check_existing_instances
            ;;
    esac
}

# Stop existing framework instances
stop_existing_instances() {
    local containers_to_stop=($(find_framework_containers "running"))

    if [ ${#containers_to_stop[@]} -gt 0 ]; then
        for container in "${containers_to_stop[@]}"; do
            print_verbose "Stopping container: $container"
            docker stop "$container" >/dev/null 2>&1 || true
            docker rm "$container" >/dev/null 2>&1 || true
        done
    fi
}

# Start services
start_services() {
    print_header "🚀 Starting Services"

    check_environment

    # Check for existing instances before starting
    check_existing_instances

    load_environment
    local cmd=$(get_compose_cmd)

    print_step "Starting development environment..."

    if $cmd up -d; then
        print_sub_success "Services started successfully"
        print_sub_step "Waiting for services to be ready..."
        sleep 10
        print_sub_success "Health checks completed"
        print_section_break

        show_status
        show_access_info

        print_celebration "Development environment ready!"
    else
        print_error "Failed to start services"
        exit 1
    fi
}

# Stop services
stop_services() {
    print_header "⏹️  Stopping Services"

    check_environment
    local cmd=$(get_compose_cmd)

    print_step "Stopping development environment..."

    if $cmd down; then
        print_success "Services stopped successfully"
    else
        print_error "Failed to stop services"
        exit 1
    fi
}

# Restart services
restart_services() {
    print_header "🔄 Restarting Services"

    stop_services
    echo ""
    start_services
}

# Show logs
show_logs() {
    local service="$1"
    local follow_flag=""

    check_environment
    local cmd=$(get_compose_cmd)

    if [ "$2" = "--follow" ] || [ "$2" = "-f" ]; then
        follow_flag="-f"
        print_header "📋 Following Logs${service:+ for $service}"
        print_info "Press Ctrl+C to stop following logs"
    else
        print_header "📋 Recent Logs${service:+ for $service}"
    fi

    if [ -n "$service" ]; then
        $cmd logs $follow_flag --tail=50 "$service"
    else
        $cmd logs $follow_flag --tail=50
    fi
}

# Execute command in service container
exec_service() {
    local service="$1"
    shift
    local command="$@"

    check_environment
    local cmd=$(get_compose_cmd)

    print_step "Executing in $service: $command"

    $cmd exec "$service" $command
}

# Connect to service (interactive shell)
connect_service() {
    local service="$1"

    check_environment
    local cmd=$(get_compose_cmd)

    print_header "🔗 Connecting to $service"

    case $service in
        redis)
            print_info "Connecting to Redis CLI..."
            $cmd exec redis redis-cli -a "${REDIS_PASSWORD:-password}"
            ;;
        postgres)
            print_info "Connecting to PostgreSQL..."
            $cmd exec postgres psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-local_dev}"
            ;;
        mysql)
            print_info "Connecting to MySQL..."
            $cmd exec mysql mysql -u "${MYSQL_USER:-root}" -p"${MYSQL_PASSWORD:-password}" "${MYSQL_DATABASE:-local_dev}"
            ;;
        kafka)
            print_info "Connecting to Kafka console consumer..."
            $cmd exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic test --from-beginning
            ;;
        localstack)
            print_info "Connecting to LocalStack AWS CLI..."
            $cmd exec localstack bash -c "aws --endpoint-url=http://localhost:4566 --region=us-east-1 $@"
            ;;
        *)
            print_info "Opening shell in $service..."
            $cmd exec "$service" /bin/sh
            ;;
    esac
}

# Backup service data
backup_service() {
    local service="$1"
    local backup_dir="./backups"
    local timestamp=$(date +%Y%m%d_%H%M%S)

    check_environment
    local cmd=$(get_compose_cmd)

    ensure_directory "$backup_dir"

    print_header "💾 Backing up $service"

    case $service in
        postgres)
            local backup_file="$backup_dir/postgres_backup_$timestamp.sql"
            print_step "Creating database backup..."
            $cmd exec postgres pg_dump -U "${POSTGRES_USER:-postgres}" "${POSTGRES_DB:-local_dev}" > "$backup_file"
            print_sub_success "PostgreSQL backup completed"
            print_sub_info "Saved to: $backup_file"
            ;;
        mysql)
            local backup_file="$backup_dir/mysql_backup_$timestamp.sql"
            print_step "Creating database backup..."
            $cmd exec mysql mysqldump -u "${MYSQL_USER:-root}" -p"${MYSQL_PASSWORD:-password}" "${MYSQL_DATABASE:-local_dev}" > "$backup_file"
            print_sub_success "MySQL backup completed"
            print_sub_info "Saved to: $backup_file"
            ;;
        redis)
            local backup_file="$backup_dir/redis_backup_$timestamp.rdb"
            print_step "Creating cache backup..."
            $cmd exec redis redis-cli -a "${REDIS_PASSWORD:-password}" --rdb "$backup_file"
            print_sub_success "Redis backup completed"
            print_sub_info "Saved to: $backup_file"
            ;;
        kafka)
            print_warning "Kafka backup requires external tools. Consider using Kafka MirrorMaker or Confluent Replicator for production backups."
            print_info "Topics can be listed with: ./manage.sh exec kafka kafka-topics --list --bootstrap-server localhost:9092"
            ;;
        localstack)
            print_warning "LocalStack data backup should be done by exporting AWS resources configuration."
            print_info "SQS queues: aws --endpoint-url=http://localhost:4566 sqs list-queues"
            print_info "SNS topics: aws --endpoint-url=http://localhost:4566 sns list-topics"
            ;;
        *)
            print_error "Backup not supported for service: $service"
            exit 1
            ;;
    esac

    print_section_break
}

# Restore service data
restore_service() {
    local service="$1"
    local backup_file="$2"

    if [ -z "$backup_file" ]; then
        print_error "Backup file not specified"
        exit 1
    fi

    if [ ! -f "$backup_file" ]; then
        print_error "Backup file not found: $backup_file"
        exit 1
    fi

    check_environment
    local cmd=$(get_compose_cmd)

    print_header "📥 Restoring $service from $backup_file"
    print_warning "This will overwrite existing data!"

    echo -n "Continue? (y/N): "
    read -r confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "Restore cancelled"
        exit 0
    fi

    case $service in
        postgres)
            print_step "Restoring PostgreSQL backup..."
            $cmd exec -T postgres psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-local_dev}" < "$backup_file"
            print_success "PostgreSQL restore completed"
            ;;
        mysql)
            print_step "Restoring MySQL backup..."
            $cmd exec -T mysql mysql -u "${MYSQL_USER:-root}" -p"${MYSQL_PASSWORD:-password}" "${MYSQL_DATABASE:-local_dev}" < "$backup_file"
            print_success "MySQL restore completed"
            ;;
        *)
            print_error "Restore not supported for service: $service"
            exit 1
            ;;
    esac
}

# Clean up everything
cleanup() {
    print_header "🧹 Cleaning Up Development Environment"

    check_environment
    local cmd=$(get_compose_cmd)

    print_warning "This will remove all containers, volumes, and data!"
    echo -n "Are you sure? (y/N): "
    read -r confirm

    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "Cleanup cancelled"
        exit 0
    fi

    print_step "Cleaning up environment..."

    print_sub_step "Stopping and removing containers..."
    $cmd down -v --remove-orphans

    print_sub_step "Removing generated files..."
    rm -f "$COMPOSE_FILE" "$ENV_FILE" "$WORK_DIR/application-local.yml.generated"

    print_sub_step "Pruning unused Docker resources..."
    docker system prune -f

    print_success "Cleanup completed"
    print_section_break
}

# Show access information
show_access_info() {
    print_header "🎯 Service Access Information"

    load_environment
    local cmd=$(get_compose_cmd)

    # Get running services
    local running_services=$($cmd ps --services | xargs)

    for service in $running_services; do
        echo -e "${BOLD}$service:${NC}"

        case $service in
            redis)
                echo -e "  🔗 Connection: redis://localhost:${REDIS_PORT:-6379}"
                echo -e "  🔑 Password: ${REDIS_PASSWORD:-password}"
                echo -e "  💻 CLI: ./manage.sh connect redis"
                ;;
            postgres)
                echo -e "  🔗 Connection: postgresql://localhost:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-local_dev}"
                echo -e "  👤 Username: ${POSTGRES_USER:-postgres}"
                echo -e "  🔑 Password: ${POSTGRES_PASSWORD:-password}"
                echo -e "  💻 CLI: ./manage.sh connect postgres"
                ;;
            mysql)
                echo -e "  🔗 Connection: mysql://localhost:${MYSQL_PORT:-3306}/${MYSQL_DATABASE:-local_dev}"
                echo -e "  👤 Username: ${MYSQL_USER:-root}"
                echo -e "  🔑 Password: ${MYSQL_PASSWORD:-password}"
                echo -e "  💻 CLI: ./manage.sh connect mysql"
                ;;
            jaeger)
                echo -e "  🌐 Web UI: http://localhost:${JAEGER_UI_PORT:-16686}"
                echo -e "  📡 OTLP HTTP: http://localhost:${JAEGER_OTLP_HTTP_PORT:-4318}/v1/traces"
                echo -e "  📡 OTLP gRPC: localhost:${JAEGER_OTLP_GRPC_PORT:-4317}"
                ;;
            prometheus)
                echo -e "  🌐 Web UI: http://localhost:${PROMETHEUS_PORT:-9090}"
                echo -e "  📊 Metrics: http://localhost:${PROMETHEUS_PORT:-9090}/metrics"
                ;;
            kafka)
                echo -e "  🔗 Brokers: localhost:${KAFKA_PORT:-9092}"
                echo -e "  🌐 UI: http://localhost:${KAFKA_UI_PORT:-8080}"
                echo -e "  💻 CLI: ./manage.sh connect kafka"
                ;;
            localstack)
                echo -e "  🔗 AWS Endpoint: http://localhost:${LOCALSTACK_PORT:-4566}"
                echo -e "  🌐 Dashboard: http://localhost:${LOCALSTACK_DASHBOARD_PORT:-8055}"
                echo -e "  🔑 Access Key: test"
                echo -e "  🔑 Secret Key: test"
                ;;
        esac

        echo ""
    done

    print_info "Configuration files:"
    [ -f "$COMPOSE_FILE" ] && echo -e "  📄 $(basename "$COMPOSE_FILE")"
    [ -f "$ENV_FILE" ] && echo -e "  📄 $(basename "$ENV_FILE")"
    [ -f "$WORK_DIR/application-local.yml.generated" ] && echo -e "  📄 application-local.yml.generated"
}

# Update services (pull latest images)
update_services() {
    print_header "🔄 Updating Services"

    check_environment
    local cmd=$(get_compose_cmd)

    print_step "Pulling latest images..."
    $cmd pull

    print_step "Recreating containers with new images..."
    $cmd up -d --force-recreate

    print_success "Services updated successfully"
    show_status
}

# Scale services
scale_service() {
    local service="$1"
    local replicas="$2"

    if [ -z "$service" ] || [ -z "$replicas" ]; then
        print_error "Usage: ./manage.sh scale <service> <replicas>"
        exit 1
    fi

    check_environment
    local cmd=$(get_compose_cmd)

    print_step "Scaling $service to $replicas replicas..."
    $cmd up -d --scale "$service=$replicas"

    print_success "Service scaled successfully"
    show_status
}

# Monitor services (real-time stats)
monitor_services() {
    print_header "📈 Service Monitoring"
    print_info "Press Ctrl+C to stop monitoring"

    check_environment

    # Use docker stats to show real-time resource usage
    docker stats $(docker compose -f "$COMPOSE_FILE" ps -q) --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
}

# List all framework instances across the system
list_all_instances() {
    print_header "🔍 All Framework Instances"

    local all_containers=($(find_framework_containers "all"))

    if [ ${#all_containers[@]} -gt 0 ]; then
        print_info "Framework containers found:"
        print_section_break
        printf "%-30s %-20s %s\n" "CONTAINER NAME" "STATUS" "PORTS"
        echo "$(printf '%.0s-' $(seq 1 80))"
        for container in "${all_containers[@]}"; do
            local status=$(docker ps -a --filter "name=$container" --format "{{.Status}}" 2>/dev/null || echo "Unknown")
            local ports=$(docker ps -a --filter "name=$container" --format "{{.Ports}}" 2>/dev/null || echo "")
            printf "%-30s %-20s %s\n" "$container" "$status" "$ports"
        done
    else
        print_info "No framework containers found"
    fi

    print_section_break
    local framework_networks=($(find_framework_networks))
    if [ ${#framework_networks[@]} -gt 0 ]; then
        print_info "Framework networks found:"
        for network in "${framework_networks[@]}"; do
            print_sub_info "$network"
        done
    fi

    print_section_break
    local framework_volumes=($(find_framework_volumes))
    if [ ${#framework_volumes[@]} -gt 0 ]; then
        print_info "Framework volumes found:"
        for volume in "${framework_volumes[@]}"; do
            local size=$(docker system df -v 2>/dev/null | grep "$volume" | awk '{print $3}' || echo "Unknown")
            print_sub_info "$volume ($size)"
        done
    fi
}

# Cleanup all framework instances across the system
cleanup_all_instances() {
    print_header "🧹 Cleaning Up All Framework Instances"

    print_warning "This will remove ALL framework containers, networks, and volumes from this system!"
    print_warning "This includes instances from other projects using this framework!"
    echo -n "Are you sure you want to continue? (type 'yes' to confirm): "
    read -r confirm

    if [ "$confirm" != "yes" ]; then
        print_info "Cleanup cancelled"
        return 0
    fi

    print_step "Cleaning up all framework instances..."

    print_sub_step "Stopping framework containers..."
    local all_containers=($(find_framework_containers "running"))
    if [ ${#all_containers[@]} -gt 0 ]; then
        for container in "${all_containers[@]}"; do
            print_verbose "Stopping container: $container"
            docker stop "$container" >/dev/null 2>&1 || true
        done
    fi

    print_sub_step "Removing framework containers..."
    local all_containers_to_remove=($(find_framework_containers "all"))
    if [ ${#all_containers_to_remove[@]} -gt 0 ]; then
        for container in "${all_containers_to_remove[@]}"; do
            print_verbose "Removing container: $container"
            docker rm "$container" >/dev/null 2>&1 || true
        done
    fi

    print_sub_step "Removing framework networks..."
    local all_networks=($(find_framework_networks))
    if [ ${#all_networks[@]} -gt 0 ]; then
        for network in "${all_networks[@]}"; do
            print_verbose "Removing network: $network"
            docker network rm "$network" >/dev/null 2>&1 || true
        done
    fi

    print_sub_step "Removing framework volumes..."
    local all_volumes=($(find_framework_volumes))
    if [ ${#all_volumes[@]} -gt 0 ]; then
        for volume in "${all_volumes[@]}"; do
            print_verbose "Removing volume: $volume"
            docker volume rm "$volume" >/dev/null 2>&1 || true
        done
    fi

    print_sub_step "Cleaning up Docker system..."
    docker system prune -f >/dev/null 2>&1 || true

    print_success "All framework instances cleaned up successfully"
    print_section_break
}

# Show help
show_help() {
    cat << EOF
Local Development Framework - Management Script

Usage: $0 <command> [options]

COMMANDS:
    status                  Show status of all services
    start                   Start all services
    stop                    Stop all services
    restart                 Restart all services
    logs [service] [-f]     Show logs (optionally follow)
    connect <service>       Connect to service (interactive)
    exec <service> <cmd>    Execute command in service container

    backup <service>        Backup service data
    restore <service> <file> Restore service data from backup

    update                  Update services to latest images
    scale <service> <n>     Scale service to n replicas
    monitor                 Monitor resource usage

    info                    Show service access information
    cleanup                 Remove all containers and data

    list-all                List all framework instances on this system
    cleanup-all             Remove all framework instances on this system

    help                    Show this help

EXAMPLES:
    $0 start                # Start all services
    $0 logs jaeger -f       # Follow Jaeger logs
    $0 connect redis        # Connect to Redis CLI
    $0 connect kafka        # Connect to Kafka consumer
    $0 backup postgres      # Backup PostgreSQL database
    $0 exec kafka kafka-topics --list --bootstrap-server localhost:9092  # List Kafka topics
    $0 scale redis 2        # Scale Redis to 2 replicas
    $0 list-all             # List all framework instances
    $0 cleanup-all          # Remove all framework instances (all projects)

SERVICES:
    Services available depend on your setup configuration.
    Common services: redis, postgres, mysql, jaeger, prometheus, kafka, localstack

KAFKA COMMANDS:
    List topics:     $0 exec kafka kafka-topics --list --bootstrap-server localhost:9092
    Create topic:    $0 exec kafka kafka-topics --create --bootstrap-server localhost:9092 --topic my-topic --partitions 3 --replication-factor 1
    Describe topic:  $0 exec kafka kafka-topics --describe --bootstrap-server localhost:9092 --topic my-topic
    Consume:         $0 exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic my-topic --from-beginning
    Produce:         $0 exec kafka kafka-console-producer --bootstrap-server localhost:9092 --topic my-topic

LOCALSTACK COMMANDS:
    List SQS queues: $0 exec localstack aws --endpoint-url=http://localhost:4566 sqs list-queues
    List SNS topics: $0 exec localstack aws --endpoint-url=http://localhost:4566 sns list-topics
    Send SQS msg:    $0 exec localstack aws --endpoint-url=http://localhost:4566 sqs send-message --queue-url QUEUE_URL --message-body "Hello"
    Publish SNS:     $0 exec localstack aws --endpoint-url=http://localhost:4566 sns publish --topic-arn TOPIC_ARN --message "Hello"
    List DynamoDB:   $0 exec localstack aws --endpoint-url=http://localhost:4566 dynamodb list-tables
    Describe table:  $0 exec localstack aws --endpoint-url=http://localhost:4566 dynamodb describe-table --table-name TABLE_NAME
    Put DDB item:    $0 exec localstack aws --endpoint-url=http://localhost:4566 dynamodb put-item --table-name TABLE --item '{"id":{"S":"123"}}'
    Get DDB item:    $0 exec localstack aws --endpoint-url=http://localhost:4566 dynamodb get-item --table-name TABLE --key '{"id":{"S":"123"}}'
    Dashboard:       open http://localhost:8055

EOF
}

# Main execution
main() {
    case "${1:-}" in
        "status")
            show_status
            ;;
        "start")
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            show_logs "$2" "$3"
            ;;
        "connect")
            if [ -z "$2" ]; then
                print_error "Service name required"
                echo "Usage: $0 connect <service>"
                exit 1
            fi
            connect_service "$2"
            ;;
        "exec")
            if [ -z "$2" ]; then
                print_error "Service name required"
                echo "Usage: $0 exec <service> <command>"
                exit 1
            fi
            exec_service "$2" "${@:3}"
            ;;
        "backup")
            if [ -z "$2" ]; then
                print_error "Service name required"
                echo "Usage: $0 backup <service>"
                exit 1
            fi
            backup_service "$2"
            ;;
        "restore")
            if [ -z "$2" ] || [ -z "$3" ]; then
                print_error "Service name and backup file required"
                echo "Usage: $0 restore <service> <backup_file>"
                exit 1
            fi
            restore_service "$2" "$3"
            ;;
        "update")
            update_services
            ;;
        "scale")
            scale_service "$2" "$3"
            ;;
        "monitor")
            monitor_services
            ;;
        "info")
            show_access_info
            ;;
        "cleanup")
            cleanup
            ;;
        "list-all")
            list_all_instances
            ;;
        "cleanup-all")
            cleanup_all_instances
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use './scripts/manage.sh help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
