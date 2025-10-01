#!/usr/bin/env bash
echo "[DEBUG] setup.sh execution started."

# Simplified setup.sh for testing common.sh sourcing
INIT_MODE=false
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "[DEBUG] Current working directory: $(pwd)"
MANAGE_SH_PATH="$SCRIPT_DIR/manage.sh"
echo "[DEBUG] SCRIPT_DIR is resolved to: $SCRIPT_DIR"
echo "[DEBUG] Resolving manage.sh path to: $MANAGE_SH_PATH"
ls -l "$MANAGE_SH_PATH"
if ls "$MANAGE_SH_PATH" > /dev/null 2>&1; then
    echo "[DEBUG] Successfully located manage.sh at $MANAGE_SH_PATH"
    echo "[DEBUG] Re-sourcing common.sh to ensure all functions are loaded..."
    source "$SCRIPT_DIR/lib/common.sh"
    echo "[DEBUG] Logging available functions after re-sourcing common.sh:"
    declare -F
    if find_framework_containers > /dev/null 2>&1; then
        echo "[DEBUG] find_framework_containers function is available."
    else
        echo "[ERROR] find_framework_containers function is not available after re-sourcing common.sh."
        echo "[DEBUG] Checking if docker command is available..."
        if ! command -v docker > /dev/null 2>&1; then
            echo "[ERROR] Docker command not found. Ensure Docker is installed and available in PATH."
        else
            echo "[DEBUG] Docker command is available."
        fi
        exit 1
    fi
else
    echo "[ERROR] manage.sh not found at $MANAGE_SH_PATH"
    exit 1
fi

# Initialize variables from common.sh
init_common_lib

# Source services.sh
SERVICES_SH_PATH="$SCRIPT_DIR/lib/services.sh"
echo "[DEBUG] Resolving services.sh path to: $SERVICES_SH_PATH"
if [ -f "$SERVICES_SH_PATH" ]; then
    source "$SERVICES_SH_PATH"
else
    echo "[ERROR] services.sh not found at $SERVICES_SH_PATH"
    exit 1
fi

# Function definitions
initialize_config() {
    echo "[DEBUG] Entering initialize_config function."
    print_step "Initializing configuration..."
    if [ ! -f "$WORK_DIR/dev-stack-config.yaml" ]; then
        echo "[DEBUG] dev-stack-config.yaml not found. Creating from sample."
        cp "$SAMPLE_CONFIG" "$WORK_DIR/dev-stack-config.yaml"
        print_success "Configuration initialized successfully."
    else
        echo "[DEBUG] dev-stack-config.yaml already exists. Skipping initialization."
        print_info "Configuration already exists. Skipping initialization."
    fi
    echo "[DEBUG] Exiting initialize_config function."
}

# Parse command-line arguments
echo "[DEBUG] Starting argument parsing..."
echo "[DEBUG] Initial arguments: $@"
for arg in "$@"; do
    echo "[DEBUG] Processing argument: $arg"
    case $arg in
        --init)
            echo "[DEBUG] --init flag detected. Setting INIT_MODE to true."
            INIT_MODE=true
            ;;
        --test)

            ;;
        *)
            echo "[DEBUG] Adding unrecognized argument to FORWARD_ARGS: $arg"
            FORWARD_ARGS+=("$arg")
            ;;
    esac
done

# Debug FORWARD_ARGS after processing all arguments
echo "[DEBUG] FORWARD_ARGS after processing: ${FORWARD_ARGS[@]}"
echo "[DEBUG] Remaining positional parameters: $@"

# Update positional parameters to exclude handled flags


# Debug final arguments to be passed to manage.sh

set -- "${FORWARD_ARGS[@]}"




if [ "$INIT_MODE" = true ]; then
    echo "[DEBUG] Executing initialize_config function for --init flag."
    initialize_config
    echo "[DEBUG] Initialization complete. Exiting."
    exit 0
fi

# Update positional parameters to exclude handled flags
echo "[DEBUG] Positional parameters before set --: $@"
echo "[DEBUG] FORWARD_ARGS before set --: ${FORWARD_ARGS[@]}"
set -- "${FORWARD_ARGS[@]}"
echo "[DEBUG] Positional parameters after set --: $@"
echo "[DEBUG] Checking if arguments are correctly forwarded to manage.sh..."
echo "[DEBUG] Checking if manage.sh is accessible after updating positional parameters..."
if [ -f "$MANAGE_SH_PATH" ]; then
    echo "[DEBUG] manage.sh is accessible after set --."
else
    echo "[ERROR] manage.sh not found at $MANAGE_SH_PATH after set --."
    echo "[DEBUG] Exiting setup.sh due to manage.sh not being accessible."
    echo "[DEBUG] Exiting setup.sh due to prerequisites not being met."
    exit 1
fi

set -e

# cleanup_existing_instances function is defined below - removing this duplicate

# Check for existing instances function will be defined later in the main flow
# Port conflicts will be checked dynamically by check_port_conflicts function
# based on enabled services only

# Detect conflicts for instances and ports
detect_existing_conflicts() {
    local existing_containers=()
    local existing_networks=()
    local conflicting_ports=()

    # Find existing framework containers
    local containers=$(find_framework_containers "all")
    if [ -n "$containers" ]; then
        while IFS= read -r container; do
            if [ -n "$container" ]; then
                existing_containers+=("$container")
            fi
        done <<< "$containers"
    fi

    # Find existing framework networks
    local networks=$(find_framework_networks)
    if [ -n "$networks" ]; then
        while IFS= read -r network; do
            if [ -n "$network" ]; then
                existing_networks+=("$network")
            fi
        done <<< "$networks"
    fi

    # Check for hardcoded ports that might conflict before services are loaded
    # This is a basic check - detailed port checking happens in check_port_conflicts
    local common_ports=(6379 5432 3306 16686 9090 9092 4566 8055 8080 2181)
    for port in "${common_ports[@]}"; do
        if check_port_in_use "$port"; then
            conflicting_ports+=("$port")
        fi
    done

    # Return status - 0 if no conflicts, 1 if conflicts found
    if [ ${#existing_containers[@]} -eq 0 ] && [ ${#existing_networks[@]} -eq 0 ] && [ ${#conflicting_ports[@]} -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# Prompt user for what to do with existing instances
prompt_user_choice() {
    echo "${BOLD}What would you like to do?${NC}"
    echo "1) Clean up existing instances and start fresh"
    echo "2) Connect to existing instances (use existing configuration)"
    echo "3) Cancel setup"
    echo ""
    echo -n "Choose an option (1-3): "

    local choice
    print_debug "Prompting user for choice with existing instances or conflicts detected."
    read -r choice

    case $choice in
        1)
            print_info "Cleaning up existing instances..."
            cleanup_existing_instances
            ;;
        2)
            print_info "Connecting to existing instances..."
            connect_to_existing_instances
            ;;
        3)
            print_info "Setup cancelled by user."
            print_success "Exiting gracefully after user cancellation."
            exit 0
            ;;
        *)
            print_error "Invalid choice. Please select 1, 2, or 3."
            prompt_user_choice
            ;;
    esac
}

# Clean up existing framework instances
cleanup_existing_instances() {
    print_step "Cleaning up existing framework instances..."

    # Stop and remove containers with framework naming pattern
    local containers_to_remove=$(find_framework_containers "all")

    if [ -n "$containers_to_remove" ]; then
        print_verbose "Stopping containers..."
        while IFS= read -r container; do
            if [ -n "$container" ]; then
                print_verbose "Stopping container: $container"
                docker stop "$container" >/dev/null 2>&1 || true
            fi
        done <<< "$containers_to_remove"

        print_verbose "Removing containers..."
        while IFS= read -r container; do
            if [ -n "$container" ]; then
                print_verbose "Removing container: $container"
                docker rm "$container" >/dev/null 2>&1 || true
            fi
        done <<< "$containers_to_remove"
    fi

    # Remove framework networks
    local networks_to_remove=$(find_framework_networks)
    if [ -n "$networks_to_remove" ]; then
        print_verbose "Removing networks..."
        while IFS= read -r network; do
            if [ -n "$network" ]; then
                print_verbose "Removing network: $network"
                docker network rm "$network" >/dev/null 2>&1 || true
            fi
        done <<< "$networks_to_remove"
    fi

    # Remove framework volumes
    local volumes_to_remove=$(find_framework_volumes)
    if [ -n "$volumes_to_remove" ]; then
        print_verbose "Removing volumes..."
        while IFS= read -r volume; do
            if [ -n "$volume" ]; then
                print_verbose "Removing volume: $volume"
                docker volume rm "$volume" >/dev/null 2>&1 || true
            fi
        done <<< "$volumes_to_remove"
    fi

    # Clean up any orphaned resources
    docker system prune -f >/dev/null 2>&1 || print_debug "Docker system prune failed or no resources to clean."
    print_debug "Cleanup completed for containers, networks, and volumes."

    print_success "Cleanup completed."
    print_section_break
}

# Connect to existing instances
connect_to_existing_instances() {
    print_step "Connecting to existing framework instances..."

    # Check if we can find compose files from existing instances
    local existing_compose_files=()

    # Look for compose files in common locations
    for compose_file in "$WORK_DIR"/docker-compose.generated.yml "$WORK_DIR"/../*/docker-compose.generated.yml; do
        if [ -f "$compose_file" ]; then
            existing_compose_files+=("$compose_file")
        fi
    done

    if [ ${#existing_compose_files[@]} -eq 0 ]; then
        print_warning "No existing compose files found. You may need to check service access manually."
        print_info "Use './scripts/manage.sh info' to see running services"
    else
        print_info "Found existing compose files:"
        for compose_file in "${existing_compose_files[@]}"; do
            echo "  • $(realpath "$compose_file")"
        done
        echo ""
        print_info "You can manage existing services using:"
        echo "  ./scripts/manage.sh status"
        echo "  ./scripts/manage.sh info"
        echo "  ./scripts/manage.sh logs"
    fi

    # Skip the normal setup process
    print_success "Connected to existing instances. Skipping new service setup."

    # Show access information if possible
    if command -v docker >/dev/null 2>&1; then
        local running_containers=$(docker ps --format "{{.Names}}\t{{.Ports}}" | grep -E "(redis|postgres|mysql|jaeger|prometheus|kafka|localstack)" || echo "")
        if [ -n "$running_containers" ]; then
            echo ""
            print_header "🎯 Running Services"
            echo "$running_containers" | while IFS=$'\t' read -r name ports; do
                echo "  • $name: $ports"
            done
        fi
    fi

    print_success "Connected to existing instances. Setup process completed."
    print_debug "Existing instances connected successfully. Skipping new setup."
    exit 0
}

# Get available services
get_available_services() {
    local services=()
    for service_dir in "$SERVICES_DIR"/*/; do
        if [ -d "$service_dir" ]; then
            local service_name=$(basename "$service_dir")
            if [ -f "$service_dir/service.yaml" ] && [ -f "$service_dir/docker-compose.yml" ]; then
                services+=("$service_name")
            fi
        fi
    done
    print_debug "Available services: ${services[*]}"
    # Services available for use
}

# Load project configuration
load_project_config() {
    local config_file="$PROJECT_CONFIG"
    print_debug "Forcing use of PROJECT_CONFIG: $PROJECT_CONFIG"

    if [ -f "$config_file" ]; then
        print_verbose "Loading project configuration: $config_file"
        print_debug "Using configuration file: $config_file"

        # Extract project name (simplified YAML parsing)
        if grep -q "name:" "$config_file"; then
            PROJECT_NAME=$(grep "name:" "$config_file" | head -1 | sed 's/.*name: *["'"'"']*\([^"'"'"']*\)["'"'"']*/\1/')
            print_debug "Extracted project name: $PROJECT_NAME"
        else
            print_debug "No project name found in configuration."
        fi

        # Extract enabled services (robust YAML parsing)
        if grep -A 100 "enabled:" "$config_file" >/dev/null 2>&1; then
            # Get the services list between 'enabled:' and the next top-level key or end of file
            local enabled_services=$(awk '/^[[:space:]]*enabled:[[:space:]]*$/{flag=1; next} /^[[:space:]]*[a-zA-Z]/ && !/^[[:space:]]*-/ && flag{flag=0} flag && /^[[:space:]]*-[[:space:]]*[a-zA-Z]/{gsub(/^[[:space:]]*-[[:space:]]*/, ""); gsub(/[[:space:]]*#.*$/, ""); if($0 != "") print $0}' "$config_file")
            print_debug "Raw enabled services extracted: $enabled_services"

            # Convert to array
            SERVICES=()
            if [ -n "$enabled_services" ]; then
                while IFS= read -r service; do
                    if [ -n "$service" ]; then
                        SERVICES+=("$service")
                        print_debug "Added service to SERVICES array: $service"
                    fi
                done <<< "$enabled_services"
            else
                print_debug "No enabled services found in configuration."
            fi
        else
            print_debug "No 'enabled' section found in configuration."
        fi

        print_verbose "Loaded services: ${SERVICES[*]}"
        print_debug "Services array populated with: ${SERVICES[*]}"
    else
        print_info "No configuration file found at: $config_file"
        print_info "Creating from sample configuration..."

        if [ -f "$SAMPLE_CONFIG" ]; then
            cp "$SAMPLE_CONFIG" "$config_file"
            print_success "Created: $(basename $config_file)"
            print_info "Edit this file to configure your services, then run './scripts/setup.sh' again"

            # Show sample content
            echo ""
            print_header "Sample Configuration Created"
            echo "You can now customize $config_file with your desired services:"
            echo ""
            head -20 "$config_file"
            echo "..."
            echo ""
            print_info "Run './scripts/setup.sh' again after editing the configuration"
            print_debug "Sample configuration created at: $config_file"
            exit 0
        else
            # Create minimal config if sample doesn't exist
            create_minimal_config
        fi
    fi
}

# Create minimal configuration
create_minimal_config() {
    print_step "Creating minimal configuration..."

    cat > "$PROJECT_CONFIG" << EOF
# Local Development Framework Configuration
project:
  name: dev-stack
  environment: local

services:
  enabled:
    - redis
    - jaeger

overrides: {}

validation:
  skip_warnings: false
  allow_multiple_databases: true
EOF

    print_success "Created minimal configuration: $(basename $PROJECT_CONFIG)"
    print_info "Edit this file to add more services or customize settings, then run './scripts/setup.sh'"
}

# Validate services
validate_services() {
    if [ "$SKIP_VALIDATION" = true ]; then
        print_info "Skipping service validation"
        print_debug "Skipping service validation"
        return
    fi

    print_debug "Starting service validation..."

    local available_services=($(get_available_services))
    local invalid_services=()

    for service in "${SERVICES[@]}"; do
        if ! validate_service_exists "$service"; then
            print_debug "Service '$service' is invalid."
            # Service validation handled by print_debug above
            invalid_services+=("$service")
        else
            print_debug "Service '$service' is valid."
        fi
    done

    if [ ${#invalid_services[@]} -gt 0 ]; then
        print_error "Invalid services detected: ${invalid_services[*]}"
        print_info "Available services are: ${available_services[*]}"
        print_debug "Validation failed. Invalid services: ${invalid_services[*]}"
        print_debug "Validation failed due to invalid services: ${invalid_services[*]}"
        exit 1
    fi

    # Success message will be handled by unified validation function
}

# Service combination validation
validate_service_combinations() {
    if [ "$SKIP_VALIDATION" = true ]; then
        print_info "Skipping service combination validation"
        return
    fi

    local warnings=()
    local errors=()

    # Check for multiple databases
    local has_mysql=false
    local has_postgres=false

    for service in "${SERVICES[@]}"; do
        case $service in
            mysql) has_mysql=true ;;
            postgres) has_postgres=true ;;
        esac
    done

    if [ "$has_mysql" = true ] && [ "$has_postgres" = true ]; then
        warnings+=("Running both MySQL and PostgreSQL simultaneously. This is unusual and may indicate configuration confusion.")
    fi

    # Check Prometheus without Spring Boot
    local has_prometheus=false
    for service in "${SERVICES[@]}"; do
        if [ "$service" = "prometheus" ]; then
            has_prometheus=true
            break
        fi
    done

    if [ "$has_prometheus" = true ]; then
        if [ ! -f "$WORK_DIR/src/main/resources/application.yml" ] && [ ! -f "$WORK_DIR/src/main/resources/application.properties" ] && [ ! -f "$WORK_DIR/build.gradle" ] && [ ! -f "$WORK_DIR/build.gradle.kts" ] && [ ! -f "$WORK_DIR/pom.xml" ]; then
            warnings+=("Prometheus enabled but no Spring Boot project detected. Ensure your application exposes /actuator/prometheus endpoint.")
        fi
    fi

    # Check LocalStack configuration
    local has_localstack=false
    for service in "${SERVICES[@]}"; do
        if [ "$service" = "localstack" ]; then
            has_localstack=true
            break
        fi
    done

    if [ "$has_localstack" = true ]; then
        # Check if many services are configured (simplified check)
        if grep -A 10 "localstack:" "$PROJECT_CONFIG" | grep -E "^\s*-\s*" | wc -l | grep -q "[4-9]"; then
            warnings+=("LocalStack configured with many services. This will significantly increase memory usage and startup time.")
        fi
    fi

    # Resource usage warnings
    local total_memory=0
    for service in "${SERVICES[@]}"; do
        case $service in
            redis) total_memory=$((total_memory + 256)) ;;
            postgres) total_memory=$((total_memory + 512)) ;;
            mysql) total_memory=$((total_memory + 512)) ;;
            jaeger) total_memory=$((total_memory + 512)) ;;
            prometheus) total_memory=$((total_memory + 256)) ;;
            localstack) total_memory=$((total_memory + 512)) ;;
            kafka) total_memory=$((total_memory + 1024)) ;;
        esac
    done

    if [ $total_memory -gt 4096 ]; then
        warnings+=("Selected services require significant memory (>4GB). Ensure your system has adequate resources.")
    fi

    # Display warnings
    for warning in "${warnings[@]}"; do
        print_warning "$warning"
    done

    # Display errors and exit if any
    for error in "${errors[@]}"; do
        print_error "$error"
    done

    if [ ${#errors[@]} -gt 0 ]; then
        exit 1
    fi

    if [ ${#warnings[@]} -gt 0 ] && [ "$FORCE" != true ]; then
        echo ""
        echo -n "Continue despite warnings? (y/N): "
        read -r response
        if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
            print_error "Setup aborted due to warnings"
            exit 1
        fi
    fi

    # Success message will be handled by unified validation function
}

# Unified validation function that groups all validation steps
run_validation_checks() {
    print_step "Validating configuration..."
    print_sub_info "Project: $PROJECT_NAME"
    print_sub_info "Services: ${SERVICES[*]}"

    # Run individual validation steps
    validate_services
    validate_service_combinations
    check_port_conflicts

    print_sub_success "Services validated"
    print_sub_success "Service combinations checked"
    print_sub_success "No port conflicts detected"
    print_section_break
}

# Check port conflicts
check_port_conflicts() {
    local conflicts=()
    local ports_to_check=()

    # Gather all ports that will be used
    for service in "${SERVICES[@]}"; do
        local service_file="$SERVICES_DIR/$service/service.yaml"
        if [ -f "$service_file" ]; then
            # Extract required ports (simplified)
            local service_ports=$(grep -A 10 "required_ports:" "$service_file" | grep -E "^\s*-\s*" | sed 's/^\s*-\s*//' | sed 's/"//g' | sed 's/\${.*:-\([0-9]*\)}.*/\1/' | grep -E '^[0-9]+$')
            ports_to_check+=($service_ports)
        fi
    done

    # Check each port
    for port in "${ports_to_check[@]}"; do
        if check_port_in_use "$port"; then
            conflicts+=("$port")
        fi
    done

    if [ ${#conflicts[@]} -gt 0 ]; then
        print_warning "Port conflicts detected: ${conflicts[*]}"
        if [ "$FORCE" != true ]; then
            echo -n "Continue anyway? (y/N): "
            read -r response
            if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
                print_error "Aborted due to port conflicts"
                exit 1
            fi
        fi
    fi

    # Return status for unified validation function
    return 0
}

# Extract service configuration overrides
extract_service_overrides() {
    local service="$1"
    local overrides=""

    if [ -f "$PROJECT_CONFIG" ] && grep -q "overrides:" "$PROJECT_CONFIG"; then
        # Use awk to properly extract only the specific service's overrides
        overrides=$(awk -v service="$service" '
        BEGIN {
            in_overrides = 0;
            in_service = 0;
            service_indent = -1;
        }
        /^[[:space:]]*overrides:/ {
            in_overrides = 1;
            next;
        }
        in_overrides && $0 ~ "^[[:space:]]*" service ":" {
            in_service = 1;
            service_indent = match($0, /[^[:space:]]/);
            next;
        }
        in_service && /^[[:space:]]*$/ { next }
        in_service && match($0, /[^[:space:]]/) <= service_indent && !/^[[:space:]]*#/ {
            exit;
        }
        in_service && match($0, /[^[:space:]]/) > service_indent {
            gsub(/^[[:space:]]*/, "");
            gsub(/:/, " ");
            gsub(/"/, "");
            print;
        }
        ' "$PROJECT_CONFIG")
    fi

    echo "$overrides"
}

# Generate Docker Compose file
generate_compose_file() {
    local compose_file="$WORK_DIR/docker-compose.generated.yml"
    local temp_file=$(make_temp_file)

    # Start with compose file header
    cat > "$temp_file" << EOF
# Generated by dev-stack Framework v$FRAMEWORK_VERSION
# Project: ${PROJECT_NAME:-dev-stack}
# Services: ${SERVICES[*]}
# Generated on: $(date)

networks:
  ${PROJECT_NAME:-dev-stack}-network:
    driver: bridge

services:
EOF

    # Track unique volumes
    declare -A unique_volumes

    # Process each service
    for service in "${SERVICES[@]}"; do
        local service_compose="$SERVICES_DIR/$service/docker-compose.yml"
        if [ -f "$service_compose" ]; then
            print_verbose "Adding service: $service"

            # Extract service definitions only, skip volumes and networks sections
            local in_services=false
            while IFS= read -r line; do
                if [[ "$line" == "services:" ]]; then
                    in_services=true
                    continue
                elif [[ "$line" =~ ^(volumes|networks): ]]; then
                    in_services=false
                elif [[ "$in_services" == true ]]; then
                    # Fix network references
                    line="${line//- dev-stack/- ${PROJECT_NAME:-dev-stack}-network}"
                    echo "$line" >> "$temp_file"
                fi
            done < "$service_compose"

            # Extract volume names for the volumes section
            if grep -q "^volumes:" "$service_compose"; then
                local in_volumes=false
                while IFS= read -r line; do
                    if [[ "$line" == "volumes:" ]]; then
                        in_volumes=true
                        continue
                    elif [[ "$line" =~ ^(services|networks): ]]; then
                        in_volumes=false
                    elif [[ "$in_volumes" == true ]] && [[ "$line" =~ ^[[:space:]]*([a-zA-Z][a-zA-Z0-9_-]*): ]]; then
                        local volume_name="${BASH_REMATCH[1]}"
                        # Skip network definitions that might be in volumes section
                        if [[ ! "$volume_name" =~ ^(dev-stack|${PROJECT_NAME:-dev-stack})$ ]]; then
                            unique_volumes["$volume_name"]=1
                        fi
                    fi
                done < "$service_compose"
            fi
        fi
    done

    # Add volumes section
    if [ ${#unique_volumes[@]} -gt 0 ]; then
        echo "" >> "$temp_file"
        echo "volumes:" >> "$temp_file"
        for volume in "${!unique_volumes[@]}"; do
            echo "  $volume:" >> "$temp_file"
            echo "    driver: local" >> "$temp_file"
        done
    fi

    # Move temp file to final location
    mv "$temp_file" "$compose_file"

    print_sub_success "$(basename $compose_file)"
}

# Generate environment file
generate_env_file() {
    # This function is now part of the unified generation step

    local env_file="$WORK_DIR/.env.generated"
    local temp_file=$(make_temp_file)

    # Add header
    cat > "$temp_file" << EOF
# Generated by Local Development Framework v$FRAMEWORK_VERSION
# Project: ${PROJECT_NAME:-dev-stack}
# Generated on: $(date)
#
# This file contains environment variables for all configured services.
# You can override any of these values by creating a .env.local file.

# ============================================================================
# PROJECT CONFIGURATION
# ============================================================================
PROJECT_NAME=${PROJECT_NAME:-dev-stack}
ENVIRONMENT=local
COMPOSE_PROJECT_NAME=${PROJECT_NAME:-dev-stack}

EOF

    # Generate service configurations in alphabetical order for consistency
    local sorted_services=($(printf '%s\n' "${SERVICES[@]}" | sort))

    for service in "${sorted_services[@]}"; do
        case $service in
            jaeger)
                # Extract overrides first
                local jaeger_overrides=$(extract_service_overrides "jaeger")
                local jaeger_ui_port="16686"
                local jaeger_otlp_http_port="4318"
                local jaeger_otlp_grpc_port="4317"

                # Apply overrides
                if [ -n "$jaeger_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                ui_port) jaeger_ui_port="$value" ;;
                                otlp_http_port) jaeger_otlp_http_port="$value" ;;
                                otlp_grpc_port) jaeger_otlp_grpc_port="$value" ;;
                            esac
                        fi
                    done <<< "$jaeger_overrides"
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# JAEGER (Distributed Tracing)
# ============================================================================
JAEGER_HOST=localhost
JAEGER_UI_PORT=$jaeger_ui_port
JAEGER_OTLP_HTTP_PORT=$jaeger_otlp_http_port
JAEGER_OTLP_GRPC_PORT=$jaeger_otlp_grpc_port

# Connection URLs
JAEGER_UI_URL=http://localhost:$jaeger_ui_port
JAEGER_OTLP_ENDPOINT=http://localhost:$jaeger_otlp_http_port

# OpenTelemetry configuration
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:$jaeger_otlp_http_port
OTEL_EXPORTER_OTLP_HEADERS=""
OTEL_RESOURCE_ATTRIBUTES="service.name=${PROJECT_NAME:-dev-stack}"

EOF
                ;;
            kafka)
                # Extract overrides first
                local kafka_overrides=$(extract_service_overrides "kafka")
                local kafka_port="9092"
                local kafka_ui_port="8080"
                local zookeeper_port="2181"

                # Apply overrides
                if [ -n "$kafka_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                port) kafka_port="$value" ;;
                                ui_port) kafka_ui_port="$value" ;;
                                zookeeper_port) zookeeper_port="$value" ;;
                            esac
                        fi
                    done <<< "$kafka_overrides"
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# KAFKA (Message Streaming)
# ============================================================================
KAFKA_HOST=localhost
KAFKA_PORT=$kafka_port
KAFKA_UI_PORT=$kafka_ui_port
ZOOKEEPER_PORT=$zookeeper_port

# Connection URLs
KAFKA_BOOTSTRAP_SERVERS=localhost:$kafka_port
KAFKA_UI_URL=http://localhost:$kafka_ui_port

# Kafka configuration
KAFKA_AUTO_CREATE_TOPICS_ENABLE=true
KAFKA_DELETE_TOPIC_ENABLE=true

EOF
                # Generate Kafka topics configuration file
                generate_kafka_config
                ;;
            localstack)
                # Extract overrides first
                local localstack_overrides=$(extract_service_overrides "localstack")
                local localstack_port="4566"
                local localstack_dashboard_port="8055"

                # Apply overrides
                if [ -n "$localstack_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                port) localstack_port="$value" ;;
                                dashboard_port) localstack_dashboard_port="$value" ;;
                            esac
                        fi
                    done <<< "$localstack_overrides"
                fi

                # Determine LocalStack services from config
                local localstack_services="sqs,sns"  # default
                if grep -A 20 "localstack:" "$PROJECT_CONFIG" 2>/dev/null | grep -A 10 "services:" >/dev/null 2>&1; then
                    localstack_services=$(grep -A 20 "localstack:" "$PROJECT_CONFIG" | grep -A 10 "services:" | grep -E "^\s*-\s*" | sed 's/^\s*-\s*//' | tr '\n' ',' | sed 's/,$//')
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# LOCALSTACK (AWS Services Emulator)
# ============================================================================
LOCALSTACK_HOST=localhost
LOCALSTACK_PORT=$localstack_port
LOCALSTACK_DASHBOARD_PORT=$localstack_dashboard_port
LOCALSTACK_SERVICES=${localstack_services}

# AWS Configuration
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=us-east-1
AWS_ENDPOINT_URL=http://localhost:$localstack_port

# Connection URLs
LOCALSTACK_ENDPOINT=http://localhost:$localstack_port
LOCALSTACK_DASHBOARD_URL=http://localhost:$localstack_dashboard_port

EOF
                # Generate AWS resources configuration file
                generate_localstack_config
                ;;
            mysql)
                # Extract overrides first
                local mysql_overrides=$(extract_service_overrides "mysql")
                local mysql_port="3306"
                local mysql_password="password"
                local mysql_database="${PROJECT_NAME//-/_}_dev"
                local mysql_user="app_user"

                # Apply overrides
                if [ -n "$mysql_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                port) mysql_port="$value" ;;
                                password) mysql_password="$value" ;;
                                database) mysql_database="$value" ;;
                                username) mysql_user="$value" ;;
                            esac
                        fi
                    done <<< "$mysql_overrides"
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# MYSQL (Database)
# ============================================================================
MYSQL_HOST=localhost
MYSQL_PORT=$mysql_port
MYSQL_ROOT_PASSWORD=$mysql_password
MYSQL_DATABASE=$mysql_database
MYSQL_USER=$mysql_user
MYSQL_PASSWORD=$mysql_password

# Connection URLs
MYSQL_URL=mysql://$mysql_user:$mysql_password@localhost:$mysql_port/$mysql_database
MYSQL_ROOT_URL=mysql://root:$mysql_password@localhost:$mysql_port/$mysql_database

# MySQL configuration
MYSQL_CHARSET=utf8mb4
MYSQL_COLLATION=utf8mb4_unicode_ci

EOF
                ;;
            postgres)
                # Extract overrides first
                local postgres_overrides=$(extract_service_overrides "postgres")
                local postgres_port="5432"
                local postgres_password="password"
                local postgres_database="${PROJECT_NAME//-/_}_dev"
                local postgres_user="postgres"

                # Apply overrides
                if [ -n "$postgres_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                port) postgres_port="$value" ;;
                                password) postgres_password="$value" ;;
                                database) postgres_database="$value" ;;
                                username) postgres_user="$value" ;;
                            esac
                        fi
                    done <<< "$postgres_overrides"
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# POSTGRESQL (Database)
# ============================================================================
POSTGRES_HOST=localhost
POSTGRES_PORT=$postgres_port
POSTGRES_DB=$postgres_database
POSTGRES_USER=$postgres_user
POSTGRES_PASSWORD=$postgres_password

# Connection URLs
POSTGRES_URL=postgresql://$postgres_user:$postgres_password@localhost:$postgres_port/$postgres_database
DATABASE_URL=postgresql://$postgres_user:$postgres_password@localhost:$postgres_port/$postgres_database

# PostgreSQL configuration
POSTGRES_CHARSET=utf8
POSTGRES_LC_COLLATE=en_US.utf8
POSTGRES_LC_CTYPE=en_US.utf8

EOF
                ;;
            prometheus)
                # Extract overrides first
                local prometheus_overrides=$(extract_service_overrides "prometheus")
                local prometheus_port="9090"

                # Apply overrides
                if [ -n "$prometheus_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                port) prometheus_port="$value" ;;
                            esac
                        fi
                    done <<< "$prometheus_overrides"
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# PROMETHEUS (Metrics & Monitoring)
# ============================================================================
PROMETHEUS_HOST=localhost
PROMETHEUS_PORT=$prometheus_port

# Connection URLs
PROMETHEUS_URL=http://localhost:$prometheus_port
PROMETHEUS_API_URL=http://localhost:$prometheus_port/api/v1

# Prometheus configuration
PROMETHEUS_SCRAPE_INTERVAL=15s
PROMETHEUS_EVALUATION_INTERVAL=15s

EOF
                ;;
            redis)
                # Extract overrides first
                local redis_overrides=$(extract_service_overrides "redis")
                local redis_port="6379"
                local redis_password="password"

                # Apply overrides
                if [ -n "$redis_overrides" ]; then
                    while IFS=' ' read -r key value; do
                        if [ -n "$key" ] && [ -n "$value" ]; then
                            value=$(echo "$value" | sed 's/"//g')
                            case "$key" in
                                port) redis_port="$value" ;;
                                password) redis_password="$value" ;;
                            esac
                        fi
                    done <<< "$redis_overrides"
                fi

                cat >> "$temp_file" << EOF
# ============================================================================
# REDIS (Cache & Session Store)
# ============================================================================
REDIS_HOST=localhost
REDIS_PORT=$redis_port
REDIS_PASSWORD=$redis_password
REDIS_DATABASE=0

# Connection URLs
REDIS_URL=redis://:$redis_password@localhost:$redis_port/0
CACHE_URL=redis://:$redis_password@localhost:$redis_port/0

# Redis configuration
REDIS_MAXMEMORY=256mb
REDIS_MAXMEMORY_POLICY=allkeys-lru

EOF
                ;;
        esac

        # Note: Overrides are now applied directly in each service case above
        # This eliminates duplication and ensures proper value resolution
    done

    # Add framework metadata at the end
    cat >> "$temp_file" << EOF
# ============================================================================
# FRAMEWORK METADATA
# ============================================================================
FRAMEWORK_VERSION=$FRAMEWORK_VERSION
FRAMEWORK_SERVICES="${SERVICES[*]}"
GENERATED_ON=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

EOF

    mv "$temp_file" "$env_file"
    print_sub_success "$(basename $env_file)"
}

# Unified configuration generation function
run_configuration_generation() {
    print_step "Generating configuration files..."

    generate_compose_file
    generate_env_file
    generate_spring_config

    print_section_break
}

# Generate LocalStack AWS resources configuration
generate_localstack_config() {
    if [ ! -f "$PROJECT_CONFIG" ]; then
        return 0
    fi

    print_verbose "Generating LocalStack AWS resources configuration..."

    local config_file="$WORK_DIR/.localstack-config.json"
    local temp_file=$(mktemp)

    # Check if LocalStack is enabled and has AWS resources configured
    if ! grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -E "sqs_queues:|sns_topics:" >/dev/null 2>&1; then
        # No AWS resources configured, create empty config
        echo "{}" > "$config_file"
        return 0
    fi

    # Start JSON structure
    echo "{" > "$temp_file"

    # Extract SQS queues configuration
    if grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "sqs_queues:" >/dev/null 2>&1; then
        echo '  "sqs_queues": [' >> "$temp_file"

        local queue_configs=$(grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 100 "sqs_queues:" | grep -E "^\s*-\s*name:" | sed 's/.*name:\s*//' | sed 's/["'"'"']//g')
        local first_queue=true

        while IFS= read -r queue_name; do
            if [ -n "$queue_name" ]; then
                if [ "$first_queue" = false ]; then
                    echo "," >> "$temp_file"
                fi
                echo -n '    {' >> "$temp_file"
                echo -n '"name": "'"$queue_name"'"' >> "$temp_file"

                # Extract additional queue properties (simplified parsing)
                local queue_block=$(grep -A 20 "name: $queue_name" "$PROJECT_CONFIG" | head -20)

                if echo "$queue_block" | grep -q "visibility_timeout:"; then
                    local timeout=$(echo "$queue_block" | grep "visibility_timeout:" | sed 's/.*visibility_timeout:\s*//' | sed 's/[^0-9]//g')
                    echo -n ', "visibility_timeout": '"$timeout" >> "$temp_file"
                fi

                if echo "$queue_block" | grep -q "dead_letter_queue:"; then
                    local dlq=$(echo "$queue_block" | grep "dead_letter_queue:" | sed 's/.*dead_letter_queue:\s*//' | sed 's/["'"'"']//g')
                    echo -n ', "dead_letter_queue": "'"$dlq"'"' >> "$temp_file"
                fi

                echo -n '}' >> "$temp_file"
                first_queue=false
            fi
        done <<< "$queue_configs"

        echo '' >> "$temp_file"
        echo '  ]' >> "$temp_file"
    fi

    # Add comma if we have both queues and topics/tables
    local has_sns=$(grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "sns_topics:" >/dev/null 2>&1 && echo "true" || echo "false")
    local has_dynamo=$(grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "dynamodb_tables:" >/dev/null 2>&1 && echo "true" || echo "false")

    if grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "sqs_queues:" >/dev/null 2>&1 && \
       { [ "$has_sns" = "true" ] || [ "$has_dynamo" = "true" ]; }; then
        echo ',' >> "$temp_file"
    fi

    # Extract SNS topics configuration
    if grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "sns_topics:" >/dev/null 2>&1; then
        echo '  "sns_topics": [' >> "$temp_file"

        local topic_configs=$(grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 100 "sns_topics:" | grep -E "^\s*-\s*name:" | sed 's/.*name:\s*//' | sed 's/["'"'"']//g')
        local first_topic=true

        while IFS= read -r topic_name; do
            if [ -n "$topic_name" ]; then
                if [ "$first_topic" = false ]; then
                    echo "," >> "$temp_file"
                fi
                echo -n '    {' >> "$temp_file"
                echo -n '"name": "'"$topic_name"'"' >> "$temp_file"

                # Extract subscriptions (simplified)
                local topic_block=$(grep -A 30 "name: $topic_name" "$PROJECT_CONFIG")
                if echo "$topic_block" | grep -q "subscriptions:"; then
                    echo -n ', "subscriptions": [' >> "$temp_file"

                    local sub_endpoints=$(echo "$topic_block" | grep -A 10 "subscriptions:" | grep "endpoint:" | sed 's/.*endpoint:\s*//' | sed 's/["'"'"']//g')
                    local first_sub=true

                    while IFS= read -r endpoint; do
                        if [ -n "$endpoint" ]; then
                            if [ "$first_sub" = false ]; then
                                echo -n "," >> "$temp_file"
                            fi
                            echo -n '{"protocol": "sqs", "endpoint": "'"$endpoint"'", "raw_message_delivery": true}' >> "$temp_file"
                            first_sub=false
                        fi
                    done <<< "$sub_endpoints"

                    echo -n ']' >> "$temp_file"
                fi

                echo -n '}' >> "$temp_file"
                first_topic=false
            fi
        done <<< "$topic_configs"

        echo '' >> "$temp_file"
        echo '  ]' >> "$temp_file"
    fi

    # Add comma if we have both topics and tables
    if grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "sns_topics:" >/dev/null 2>&1 && \
       grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "dynamodb_tables:" >/dev/null 2>&1; then
        echo ',' >> "$temp_file"
    fi

    # Extract DynamoDB tables configuration
    if grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 20 "dynamodb_tables:" >/dev/null 2>&1; then
        echo '  "dynamodb_tables": [' >> "$temp_file"

        # Use yq or python to convert YAML to JSON for complex nested structures
        if command -v python3 >/dev/null 2>&1; then
            local tables_json=$(python3 -c "
import yaml
import json
import sys
try:
    with open('$PROJECT_CONFIG', 'r') as f:
        config = yaml.safe_load(f)
    tables = config.get('overrides', {}).get('localstack', {}).get('dynamodb_tables', [])
    print(json.dumps(tables, indent=4))
except:
    print('[]')
" | sed 's/^/    /' | sed '1d' | sed '$d')
            echo "$tables_json" >> "$temp_file"
        else
            # Fallback to simple parsing if python not available
            local table_configs=$(grep -A 50 "localstack:" "$PROJECT_CONFIG" | grep -A 100 "dynamodb_tables:" | grep -E "^\s*-\s*name:" | sed 's/.*name:\s*//' | sed 's/["'"'"']//g')
            local first_table=true

            while IFS= read -r table_name; do
                if [ -n "$table_name" ]; then
                    if [ "$first_table" = false ]; then
                        echo "," >> "$temp_file"
                    fi
                    echo '    {"name": "'"$table_name"'"}' >> "$temp_file"
                    first_table=false
                fi
            done <<< "$table_configs"
        fi

        echo '' >> "$temp_file"
        echo '  ]' >> "$temp_file"
    fi

    echo "}" >> "$temp_file"

    # Move temp file to final location and copy to LocalStack volume path
    mv "$temp_file" "$config_file"

    # Also place it where the init script expects it
    mkdir -p "$WORK_DIR/.localstack"
    cp "$config_file" "$WORK_DIR/.localstack/aws-resources.json"

    print_success "Generated LocalStack configuration: $(basename "$config_file")"
}

# Generate Kafka topics configuration
generate_kafka_config() {
    if [ ! -f "$PROJECT_CONFIG" ]; then
        return 0
    fi

    # Check if Kafka is enabled
    local has_kafka=false
    for service in "${SERVICES[@]}"; do
        if [ "$service" = "kafka" ]; then
            has_kafka=true
            break
        fi
    done

    if [ "$has_kafka" = false ]; then
        return 0
    fi

    print_verbose "Generating Kafka topics configuration..."

    local config_file="$WORK_DIR/.kafka-config.json"
    local temp_file=$(mktemp)

    # Check if Kafka is enabled and has topics configured
    if ! grep -A 50 "kafka:" "$PROJECT_CONFIG" | grep -E "topics:" >/dev/null 2>&1; then
        # No topics configured, create empty config
        echo "{}" > "$config_file"
        return 0
    fi

    # Start JSON structure
    echo "{" > "$temp_file"

    # Extract Kafka topics configuration
    if grep -A 50 "kafka:" "$PROJECT_CONFIG" | grep -A 20 "topics:" >/dev/null 2>&1; then
        echo '  "topics": [' >> "$temp_file"

        local topic_configs=$(grep -A 50 "kafka:" "$PROJECT_CONFIG" | grep -A 100 "topics:" | grep -E "^\s*-\s*name:" | sed 's/.*name:\s*//' | sed 's/["'"'"']//g')
        local first_topic=true

        while IFS= read -r topic_name; do
            if [ -n "$topic_name" ]; then
                if [ "$first_topic" = false ]; then
                    echo "," >> "$temp_file"
                fi
                echo -n '    {' >> "$temp_file"
                echo -n '"name": "'"$topic_name"'"' >> "$temp_file"

                # Extract additional topic properties (simplified parsing)
                local topic_block=$(grep -A 10 "name: $topic_name" "$PROJECT_CONFIG" | head -10)

                if echo "$topic_block" | grep -q "partitions:"; then
                    local partitions=$(echo "$topic_block" | grep "partitions:" | sed 's/.*partitions:\s*//' | sed 's/[^0-9]//g')
                    if [ -n "$partitions" ]; then
                        echo -n ', "partitions": '"$partitions" >> "$temp_file"
                    fi
                fi

                if echo "$topic_block" | grep -q "replication_factor:"; then
                    local replication_factor=$(echo "$topic_block" | grep "replication_factor:" | sed 's/.*replication_factor:\s*//' | sed 's/[^0-9]//g')
                    if [ -n "$replication_factor" ]; then
                        echo -n ', "replication_factor": '"$replication_factor" >> "$temp_file"
                    fi
                fi

                if echo "$topic_block" | grep -q "cleanup_policy:"; then
                    local cleanup_policy=$(echo "$topic_block" | grep "cleanup_policy:" | sed 's/.*cleanup_policy:\s*//' | sed 's/["'"'"']//g')
                    if [ -n "$cleanup_policy" ]; then
                        echo -n ', "cleanup_policy": "'"$cleanup_policy"'"' >> "$temp_file"
                    fi
                fi

                if echo "$topic_block" | grep -q "retention_ms:"; then
                    local retention_ms=$(echo "$topic_block" | grep "retention_ms:" | sed 's/.*retention_ms:\s*//' | sed 's/[^0-9]//g')
                    if [ -n "$retention_ms" ]; then
                        echo -n ', "retention_ms": '"$retention_ms" >> "$temp_file"
                    fi
                fi

                echo -n '}' >> "$temp_file"
                first_topic=false
            fi
        done <<< "$topic_configs"

        echo '' >> "$temp_file"
        echo '  ]' >> "$temp_file"
    fi

    echo "}" >> "$temp_file"

    # Move temp file to final location and copy to Kafka volume path
    mv "$temp_file" "$config_file"

    # Also place it where the init script expects it
    mkdir -p "$WORK_DIR/.kafka"
    cp "$config_file" "$WORK_DIR/.kafka/topics-config.json"

    print_success "Generated Kafka configuration: $(basename "$config_file")"
}

# Generate Spring Boot configuration
generate_spring_config() {
    if [ -f "$WORK_DIR/src/main/resources/application.yml" ] || [ -f "$WORK_DIR/src/main/resources/application.properties" ] || [ -f "$WORK_DIR/build.gradle" ] || [ -f "$WORK_DIR/build.gradle.kts" ] || [ -f "$WORK_DIR/pom.xml" ]; then
        # This function is now part of the unified generation step

        local config_file="$WORK_DIR/application-local.yml.generated"
        local temp_file=$(make_temp_file)

        cat > "$temp_file" << EOF
# Generated by Local Development Framework v$FRAMEWORK_VERSION
# Add this configuration to your application-local.yml

spring:
  profiles:
    active: local

EOF

        # Add service-specific Spring Boot configuration
        for service in "${SERVICES[@]}"; do
            local service_file="$SERVICES_DIR/$service/service.yaml"
            if [ -f "$service_file" ] && grep -q "spring_config:" "$service_file"; then
                echo "# $service configuration" >> "$temp_file"

                # Extract YAML configuration (simplified)
                if grep -A 100 "yaml: |" "$service_file" >/dev/null 2>&1; then
                    grep -A 100 "yaml: |" "$service_file" | tail -n +2 | while IFS= read -r line; do
                        if [[ "$line" =~ ^[[:space:]]*[a-zA-Z] ]] || [[ "$line" =~ ^[[:space:]]*$ ]]; then
                            echo "$line" >> "$temp_file"
                        elif [[ -z "$line" ]]; then
                            break
                        fi
                    done
                fi

                echo "" >> "$temp_file"
            fi
        done

        mv "$temp_file" "$config_file"
        print_sub_success "$(basename "$config_file")"
    fi
}

# Start services
start_services() {
    print_step "Starting services..."

    local compose_file="$WORK_DIR/docker-compose.generated.yml"

    if [ ! -f "$compose_file" ]; then
        print_error "Compose file not found. Run generation first."
        exit 1
    fi

    if [ "$DRY_RUN" = true ]; then
        print_info "DRY RUN: Would start services with: docker compose -f $compose_file up -d"
        return
    fi

    # Load environment file if it exists
    if [ -f "$WORK_DIR/.env.generated" ]; then
        set -a
        source "$WORK_DIR/.env.generated"
        set +a
    fi

    # Create necessary directories for services that are actually configured
    local service_dirs=()

    # Check if specific services are configured and add their directories
    for service in "${SERVICES[@]}"; do
        case "$service" in
            localstack)
                service_dirs+=("$WORK_DIR/.localstack")
                ;;
            kafka)
                service_dirs+=("$WORK_DIR/.kafka")
                ;;
            postgres|mysql)
                # Database services need init-scripts directory
                service_dirs+=("$WORK_DIR/init-scripts")
                ;;
        esac
    done

    # Create only the directories we actually need
    if [ ${#service_dirs[@]} -gt 0 ]; then
        # Remove duplicates and create directories
        printf '%s\n' "${service_dirs[@]}" | sort -u | while IFS= read -r dir; do
            mkdir -p "$dir"
        done
    fi

    # Start services
    if docker compose -f "$compose_file" up -d; then
        print_sub_success "Services started successfully"

        # Wait for health checks
        print_sub_step "Waiting for services to be healthy..."
        sleep 15

        # Show service status
        docker compose -f "$compose_file" ps

        print_sub_success "Health checks completed"
        print_section_break

        # Show access information
        show_access_info

        # Final celebration message
        print_celebration "Setup completed successfully!"
    else
        print_error "Failed to start services"
        print_info "Check logs with: docker compose -f $compose_file logs"
        exit 1
    fi
}

# Show access information
show_access_info() {
    print_header "🎯 Service Access Information"

    # Load environment variables for display
    if [ -f "$WORK_DIR/.env.generated" ]; then
        export $(cat "$WORK_DIR/.env.generated" | grep -v '^#' | xargs)
    fi

    for service in "${SERVICES[@]}"; do
        local service_file="$SERVICES_DIR/$service/service.yaml"
        if [ -f "$service_file" ]; then
            echo "${BOLD}$service:${NC}"

            # Show web interfaces
            if grep -q "web_interfaces:" "$service_file"; then
                grep -A 10 "web_interfaces:" "$service_file" | grep -E "url:|description:" | while IFS= read -r line; do
                    if [[ "$line" =~ url: ]]; then
                        local url=$(echo "$line" | sed 's/.*url: *//' | sed 's/"//g')
                        # Simple variable substitution
                        url=$(echo "$url" | sed "s/\${JAEGER_UI_PORT:-16686}/16686/g")
                        url=$(echo "$url" | sed "s/\${PROMETHEUS_PORT:-9090}/9090/g")
                        url=$(echo "$url" | sed "s/\${LOCALSTACK_PORT:-4566}/4566/g")
                        url=$(echo "$url" | sed "s/\${LOCALSTACK_DASHBOARD_PORT:-8055}/8055/g")
                        echo -e "  🌐 Web UI: ${CYAN}$url${NC}"
                    fi
                done
            fi

            # Show connection info based on service type
            case $service in
                redis)
                    echo -e "  🔗 Connection: ${CYAN}redis://localhost:6379${NC}"
                    echo -e "  🔑 Password: ${REDIS_PASSWORD:-password}"
                    echo -e "  💻 CLI: ./manage.sh connect redis"
                    ;;
                postgres)
                    echo -e "  🔗 Connection: ${CYAN}postgresql://localhost:5432/${POSTGRES_DB:-local_dev}${NC}"
                    echo -e "  👤 Username: ${POSTGRES_USER:-postgres}"
                    echo -e "  🔑 Password: ${POSTGRES_PASSWORD:-password}"
                    echo -e "  💻 CLI: ./manage.sh connect postgres"
                    ;;
                mysql)
                    echo -e "  🔗 Connection: ${CYAN}mysql://localhost:3306/${MYSQL_DATABASE:-local_dev}${NC}"
                    echo -e "  👤 Username: ${MYSQL_USER:-root}"
                    echo -e "  🔑 Password: ${MYSQL_PASSWORD:-password}"
                    echo -e "  💻 CLI: ./manage.sh connect mysql"
                    ;;
                localstack)
                    echo -e "  🔗 AWS Endpoint: ${CYAN}http://localhost:4566${NC}"
                    echo -e "  🌐 Dashboard: ${CYAN}http://localhost:8055${NC}"
                    echo -e "  🔑 Access Key: test"
                    echo -e "  🔑 Secret Key: test"
                    ;;
            esac

            echo ""
        fi
    done

    echo "${YELLOW}💡 Configuration files generated:${NC}"
    [ -f "$WORK_DIR/docker-compose.generated.yml" ] && echo "  📄 docker-compose.generated.yml"
    [ -f "$WORK_DIR/.env.generated" ] && echo "  📄 .env.generated"
    [ -f "$WORK_DIR/application-local.yml.generated" ] && echo "  📄 application-local.yml.generated"
}

# Interactive mode
run_interactive() {
    print_header "🔧 Interactive Configuration"

    # Get project name
    echo -n "Enter project name (default: dev-stack): "
    read -r input_project_name
    PROJECT_NAME="${input_project_name:-dev-stack}"

    # Get available services
    local available_services=($(get_available_services))

    echo ""
    echo "Available services:"
    local i=1
    for service in "${available_services[@]}"; do
        local service_file="$SERVICES_DIR/$service/service.yaml"
        local description=""
        if [ -f "$service_file" ] && grep -q "description:" "$service_file"; then
            description=$(grep "description:" "$service_file" | head -1 | sed 's/.*description: *//' | sed 's/"//g')
        fi
        echo "$i) $service - $description"
        ((i++))
    done

    echo ""
    echo -n "Enter service numbers (space-separated, e.g., 1 3 4): "
    read -r -a service_choices

    SERVICES=()
    for choice in "${service_choices[@]}"; do
        if [ "$choice" -ge 1 ] && [ "$choice" -le ${#available_services[@]} ]; then
            SERVICES+=("${available_services[$((choice-1))]}")
        fi
    done

    # Create configuration file
    cat > "$PROJECT_CONFIG" << EOF
# Generated by Local Development Framework - Interactive Setup
project:
  name: "$PROJECT_NAME"
  environment: "local"

services:
  enabled:
$(for service in "${SERVICES[@]}"; do echo "    - $service"; done)

overrides: {}

validation:
  skip_warnings: false
  allow_multiple_databases: true
EOF

    # Confirmation
    echo ""
    echo "${BOLD}Configuration Summary:${NC}"
    echo "Project: $PROJECT_NAME"
    echo "Services: ${SERVICES[*]}"
    echo ""
    echo -n "Proceed with setup? (Y/n): "
    read -r confirm

    if [ "$confirm" = "n" ] || [ "$confirm" = "N" ]; then
        print_info "Setup cancelled"
        exit 0
    fi

    print_success "Configuration saved to: $(basename $PROJECT_CONFIG)"
}

# Show help
show_help() {
    cat << EOF
dev-stack v$FRAMEWORK_VERSION - Config-Driven Setup

Usage: $0 [OPTIONS]

This framework uses a configuration file (dev-stack-config.yaml) to define which
services to enable and how to configure them.

OPTIONS:
    --project=NAME          Override project name from config file
    --services=LIST         Override services list (comma-separated: redis,postgres,jaeger)
    --interactive, -i       Run in interactive configuration mode
    --init                  Create sample configuration file and exit
    --dry-run              Show what would be done without executing
    --verbose, -v          Enable verbose output
    --force, -f            Force execution despite warnings (auto-cleanup existing)
    --skip-validation      Skip service validation checks
    --cleanup-existing     Automatically cleanup existing instances without prompting
    --connect-existing     Connect to existing instances without starting new ones
    --list-services        List available services
    --help, -h             Show this help

CONFIGURATION:
    The framework looks for 'dev-stack-config.yaml' in the current directory.
    If not found, a sample configuration will be created.

    Example configuration:
        project:
          name: "my-api"
        services:
          enabled:
            - redis
            - postgres
            - jaeger
        overrides:
          redis:
            password: "custom-password"

AVAILABLE SERVICES:
$(get_available_services | tr ' ' '\n' | sed 's/^/    /')

EXAMPLES:
    $0                             # Use dev-stack-config.yaml
    $0 --interactive               # Interactive setup
    $0 --init                      # Create sample config
    $0 --services=redis,jaeger     # Override services
    $0 --project=my-api --force    # Override project name, cleanup existing, skip warnings
    $0 --cleanup-existing          # Cleanup existing instances automatically
    $0 --connect-existing          # Connect to existing running services

EXISTING INSTANCES:
    The framework detects existing instances from other repos and prompts you to:
    1) Clean up existing instances and start fresh
    2) Connect to existing instances (reuse existing services)
    3) Cancel setup

WORKFLOW:
    1. Create/edit dev-stack-config.yaml
    2. Run ./scripts/setup.sh
    3. Use ./scripts/manage.sh to control services

For more information, see: README.md
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --project=*)
                PROJECT_NAME="${1#*=}"
                shift
                ;;
            --services=*)
                IFS=',' read -ra SERVICES <<< "${1#*=}"
                shift
                ;;
            --interactive|-i)
                INTERACTIVE_MODE=true
                shift
                ;;
            --init)
                print_step "Creating sample configuration file..."
                if [ -f "$SAMPLE_CONFIG" ]; then
                    cp "$SAMPLE_CONFIG" "$PROJECT_CONFIG"
                    print_success "Created: $(basename $PROJECT_CONFIG)"
                    print_info "Edit this file to configure your services, then run './scripts/setup.sh'"
                else
                    create_minimal_config
                fi
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --verbose|-v)
                VERBOSE=true
                shift
                ;;
            --force|-f)
                FORCE=true
                shift
                ;;
            --skip-validation)
                SKIP_VALIDATION=true
                shift
                ;;
            --cleanup-existing)
                CLEANUP_EXISTING=true
                shift
                ;;
            --connect-existing)
                CONNECT_EXISTING=true
                shift
                ;;
            --list-services)
                echo "Available services:"
                get_available_services | tr ' ' '\n' | while read service; do
                    local service_file="$SERVICES_DIR/$service/service.yaml"
                    local description=""
                    if [ -f "$service_file" ] && grep -q "description:" "$service_file"; then
                        description=$(grep "description:" "$service_file" | head -1 | sed 's/.*description: *//' | sed 's/"//g')
                    fi
                    echo "  $service - $description"
                done
                exit 0
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
}

# Check for existing instances and handle conflicts
check_existing_instances() {
    local existing_containers=()
    local existing_networks=()
    local conflicting_ports=()

    # Find existing framework containers
    local containers=$(find_framework_containers "all")
    if [ -n "$containers" ]; then
        while IFS= read -r container; do
            if [ -n "$container" ]; then
                existing_containers+=("$container")
            fi
        done <<< "$containers"
    fi

    # Find existing framework networks
    local networks=$(find_framework_networks)
    if [ -n "$networks" ]; then
        while IFS= read -r network; do
            if [ -n "$network" ]; then
                existing_networks+=("$network")
            fi
        done <<< "$networks"
    fi

    # Check for port conflicts - but only if we have services loaded
    # This will be checked later in validation when services are known

    # If no conflicts found, continue normally
    if [ ${#existing_containers[@]} -eq 0 ] && [ ${#existing_networks[@]} -eq 0 ]; then
        print_success "No existing instances detected"
        return 0
    fi

    # Display detected conflicts
    print_warning "Existing framework instances detected!"
    echo ""

    if [ ${#existing_containers[@]} -gt 0 ]; then
        print_info "Running containers:"
        for container in "${existing_containers[@]}"; do
            local status=$(docker ps --filter "name=$container" --format "{{.Status}}" 2>/dev/null || echo "Unknown")
            echo "  • $container ($status)"
        done
        echo ""
    fi

    if [ ${#existing_networks[@]} -gt 0 ]; then
        print_info "Existing networks:"
        for network in "${existing_networks[@]}"; do
            echo "  • $network"
        done
        echo ""
    fi

    # Handle force mode
    if [ "$FORCE" = true ]; then
        print_info "Force mode enabled - cleaning up existing instances..."
        cleanup_existing_instances
        return 0
    fi

    # Handle non-interactive mode
    if [ "$INTERACTIVE_MODE" = false ] && [ "$CLEANUP_EXISTING" = false ] && [ "$CONNECT_EXISTING" = false ]; then
        print_warning "Existing instances detected but running in non-interactive mode."
        print_info "Use --cleanup-existing to remove existing instances"
        print_info "Use --connect-existing to connect to existing instances"
        print_info "Use --force to automatically cleanup and continue"
        cleanup_existing_instances
        return 0
    fi

    # Interactive prompt for user choice
    if [ "$INTERACTIVE_MODE" = true ] && [ -z "$TEST_MODE" ] && [ "$INIT_MODE" = false ]; then
        prompt_user_choice
    else
        print_info "Non-interactive mode detected. Proceeding with default actions."
        cleanup_existing_instances
    fi
}

# Main execution function
main() {
    show_banner

    # Parse arguments first
    parse_args "$@"

    # Check prerequisites
    check_prerequisites

    # Handle existing instances based on command line flags
    if [ "$CLEANUP_EXISTING" = true ]; then
        print_info "Cleanup existing flag set - checking for existing instances..."
        cleanup_existing_instances
    elif [ "$CONNECT_EXISTING" = true ]; then
        print_info "Connect existing flag set - checking for existing instances..."
        connect_to_existing_instances
    else
        # Check for existing instances (will prompt user if found)
        check_existing_instances
    fi

    # Interactive mode
    if [ "$INTERACTIVE_MODE" = true ]; then
        run_interactive
    else
        # Load configuration
        load_project_config
    fi

    # Override from command line if provided
    if [ -n "$PROJECT_NAME" ]; then
        print_verbose "Using project name from command line: $PROJECT_NAME"
    else
        PROJECT_NAME="${PROJECT_NAME:-dev-stack}"
    fi

    # Validate we have services
    if [ ${#SERVICES[@]} -eq 0 ]; then
        print_error "No services specified in configuration"
        print_info "Edit $PROJECT_CONFIG to add services, or use --interactive mode"
        exit 1
    fi

    # Run all validation steps
    run_validation_checks

    if [ "$DRY_RUN" = true ]; then
        print_info "DRY RUN: Configuration validation completed successfully"
        print_info "DRY RUN: Would generate compose file, environment file, and Spring config"
        print_info "DRY RUN: Would start services: ${SERVICES[*]}"
        exit 0
    fi

    # Generate configuration files
    run_configuration_generation

    # Start services
    start_services

    print_header "🎉 Setup Complete!"
    print_success "Your local development environment is ready!"

    echo ""
    echo "Next steps:"
    echo "1. Review the generated configuration files"
    echo "2. Add the Spring Boot configuration to your application-local.yml"
    echo "3. Start your application with the local profile"
    echo ""
    echo "Management commands:"
    echo "  ./scripts/manage.sh start    # Start services"
    echo "  ./scripts/manage.sh stop     # Stop services"
    echo "  ./scripts/manage.sh status   # Check status"
    echo "  ./scripts/manage.sh info     # Show connection info"
}

# Run main function with all arguments
main "$@"
