#!/usr/bin/env bash

# Local Development Framework - Common Library
# Shared functions, colors, and utilities for setup.sh and manage.sh

# Prevent multiple sourcing
if [[ "${BASH_SOURCE[0]}" != "${0}" ]] && [[ "${_COMMON_LIB_LOADED:-}" == "true" ]]; then
    return 0
fi
_COMMON_LIB_LOADED=true

# Framework constants
FRAMEWORK_VERSION="1.0.0"

# Determine script directory and framework paths
if [[ -n "${SCRIPT_DIR:-}" ]]; then
    # Use existing SCRIPT_DIR if already set
    FRAMEWORK_DIR="$(dirname "$SCRIPT_DIR")"
else
    # Derive from this library's location
    LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    SCRIPT_DIR="$(dirname "$LIB_DIR")"
    FRAMEWORK_DIR="$(dirname "$SCRIPT_DIR")"
fi

# Common directories
SERVICES_DIR="$FRAMEWORK_DIR/services"
CONFIG_DIR="$FRAMEWORK_DIR/config"
WORK_DIR="$PWD"

# Common configuration files
FRAMEWORK_CONFIG="$CONFIG_DIR/framework.yaml"
PROJECT_CONFIG="$WORK_DIR/local-dev-config.yaml"
SAMPLE_CONFIG="$FRAMEWORK_DIR/local-dev-config.sample.yaml"
COMPOSE_FILE="$WORK_DIR/docker-compose.generated.yml"
ENV_FILE="$WORK_DIR/.env.generated"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Basic logging functions
print_header() {
    echo -e "${BOLD}${BLUE}$1${NC}"
    echo "$(printf '=%.0s' $(seq 1 ${#1}))"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_step() {
    echo -e "${CYAN}📋 $1${NC}"
}

print_verbose() {
    if [ "${VERBOSE:-false}" = true ]; then
        echo -e "${PURPLE}🔍 $1${NC}"
    fi
}

# Enhanced logging functions for grouped output
print_sub_info() {
    echo -e "   ${BLUE}ℹ️  $1${NC}"
}

print_sub_success() {
    echo -e "   ${GREEN}✅ $1${NC}"
}

print_sub_step() {
    echo -e "   ${CYAN}• $1${NC}"
}

print_celebration() {
    echo -e "${GREEN}🎉 $1${NC}"
}

print_section_break() {
    echo ""
}

# Show framework banner (corrected version from setup.sh)
show_banner() {
    local version_text="Version $FRAMEWORK_VERSION"
    local version_padding=$(( (62 - ${#version_text}) / 2 ))
    local version_right_padding=$(( 62 - ${#version_text} - version_padding ))

    echo -e "${BOLD}${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║              🚀 Local Development Framework                   ║"
    printf "║%*s%s%*s║\n" $version_padding "" "$version_text" $version_right_padding ""
    echo "║                   Config-Driven Setup                        ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Environment and configuration functions
load_environment() {
    if [ -f "$ENV_FILE" ]; then
        # Export variables, filtering out comments and empty lines
        set -a
        source "$ENV_FILE"
        set +a
    fi
}

# Docker Compose helper
get_compose_cmd() {
    echo "docker compose -f $COMPOSE_FILE"
}

# Check if environment is set up
check_environment() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_error "No local development environment found."
        print_info "Run './scripts/setup.sh' first to create an environment."
        exit 1
    fi
}

# Check prerequisites (Docker and Docker Compose)
check_prerequisites() {
    print_step "Checking prerequisites..."

    # Check Docker
    if ! command -v docker >/dev/null 2>&1; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Check Docker Compose
    if ! docker compose version >/dev/null 2>&1 && ! command -v docker-compose >/dev/null 2>&1; then
        print_error "Docker Compose is not installed"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker daemon is not running"
        exit 1
    fi

    print_success "Prerequisites check passed"
}

# Get available services from the framework
get_available_services() {
    if [ -d "$SERVICES_DIR" ]; then
        find "$SERVICES_DIR" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort
    fi
}

# Parse YAML configuration (simple key extraction)
parse_yaml_key() {
    local file="$1"
    local key="$2"

    if [ -f "$file" ]; then
        # Simple YAML parsing - finds key and extracts value
        grep "^$key:" "$file" 2>/dev/null | sed "s/^$key:[ ]*\(.*\)/\1/" | sed 's/[",]//g' | sed 's/^ *//' | head -1
    fi
}

# Parse YAML array (simple implementation)
parse_yaml_array() {
    local file="$1"
    local key="$2"

    if [ -f "$file" ]; then
        # Extract array items under the key
        awk "
        BEGIN { in_section=0; in_array=0 }
        /^$key:/ { in_section=1; in_array=1; next }
        in_array && /^[[:space:]]*-[[:space:]]*/ {
            gsub(/^[[:space:]]*-[[:space:]]*/, \"\")
            gsub(/[\"',]/, \"\")
            if (NF > 0) print \$0
            next
        }
        in_array && /^[[:alpha:]]/ { in_array=0; in_section=0 }
        " "$file" 2>/dev/null
    fi
}

# Extract service names from project config
extract_project_services() {
    local config_file="${1:-$PROJECT_CONFIG}"

    if [ -f "$config_file" ]; then
        # Try multiple YAML patterns for service extraction
        {
            # Pattern 1: services.enabled array
            parse_yaml_array "$config_file" "enabled" | grep -v '^[[:space:]]*$'

            # Pattern 2: services direct array
            awk '/^services:/,/^[[:alpha:]]/ {
                if (/^[[:space:]]*-[[:space:]]*[a-zA-Z]/) {
                    gsub(/^[[:space:]]*-[[:space:]]*/, "")
                    gsub(/[",]/, "")
                    print $1
                }
            }' "$config_file" 2>/dev/null

            # Pattern 3: Simple list extraction
            grep -A 20 "services:" "$config_file" 2>/dev/null | \
                awk '/^[[:space:]]*-/ { gsub(/^[[:space:]]*-[[:space:]]*/, ""); gsub(/[",]/, ""); if (NF > 0) print $1 }'
        } | sort -u | grep -E '^[a-zA-Z][a-zA-Z0-9_-]*$' | head -20
    fi
}

# Extract project name from config
extract_project_name() {
    local config_file="${1:-$PROJECT_CONFIG}"
    local project_name

    if [ -f "$config_file" ]; then
        # Try different YAML patterns
        project_name=$(parse_yaml_key "$config_file" "name" | grep -v '^[[:space:]]*$' | head -1)

        if [ -z "$project_name" ]; then
            project_name=$(awk '/^project:/,/^[[:alpha:]]/ {
                if (/^[[:space:]]*name:/) {
                    gsub(/^[[:space:]]*name:[[:space:]]*/, "")
                    gsub(/[",]/, "")
                    print $0
                    exit
                }
            }' "$config_file" 2>/dev/null)
        fi
    fi

    # Default fallback
    echo "${project_name:-local-dev}"
}

# Check for port conflicts
check_port_in_use() {
    local port="$1"
    if [ -n "$port" ] && [ "$port" != "null" ]; then
        lsof -i ":$port" >/dev/null 2>&1
    else
        return 1
    fi
}

# Get process using a port
get_port_process() {
    local port="$1"
    lsof -i ":$port" 2>/dev/null | tail -n +2 | awk '{print $1}' | head -1 || echo "Unknown"
}

# Framework container detection patterns
is_framework_container() {
    local container_name="$1"
    [[ "$container_name" =~ -redis$|-postgres$|-mysql$|-jaeger$|-prometheus$|-kafka$|-localstack$|-zookeeper$|-kafka-ui$|-kafka-init$|-localstack-init$ ]]
}

is_framework_network() {
    local network_name="$1"
    [[ "$network_name" =~ .*-network$ ]] && [ "$network_name" != "bridge" ] && [ "$network_name" != "host" ] && [ "$network_name" != "none" ]
}

is_framework_volume() {
    local volume_name="$1"
    [[ "$volume_name" =~ redis-data|postgres-data|mysql-data|kafka-data|zookeeper-data|zookeeper-logs|prometheus-data|localstack-data ]]
}

# Find all framework containers
find_framework_containers() {
    local status_filter="${1:-all}" # all, running, stopped
    local filter_flag=""

    case "$status_filter" in
        "running") filter_flag="" ;;
        "stopped") filter_flag="-a" ;;
        "all") filter_flag="-a" ;;
    esac

    docker ps $filter_flag --format "{{.Names}}" 2>/dev/null | while read -r container; do
        if [ -n "$container" ] && is_framework_container "$container"; then
            echo "$container"
        fi
    done
}

# Find all framework networks
find_framework_networks() {
    docker network ls --format "{{.Name}}" 2>/dev/null | while read -r network; do
        if [ -n "$network" ] && is_framework_network "$network"; then
            echo "$network"
        fi
    done
}

# Find all framework volumes
find_framework_volumes() {
    docker volume ls --format "{{.Name}}" 2>/dev/null | while read -r volume; do
        if [ -n "$volume" ] && is_framework_volume "$volume"; then
            echo "$volume"
        fi
    done
}

# Validation helper: check if service exists in framework
validate_service_exists() {
    local service="$1"
    local available_services=($(get_available_services))

    for available in "${available_services[@]}"; do
        if [ "$service" = "$available" ]; then
            return 0
        fi
    done
    return 1
}

# Create directory if it doesn't exist
ensure_directory() {
    local dir="$1"
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
    fi
}

# Safe file operations
backup_file() {
    local file="$1"
    local backup_suffix="${2:-.backup.$(date +%Y%m%d_%H%M%S)}"

    if [ -f "$file" ]; then
        cp "$file" "${file}${backup_suffix}"
        return $?
    fi
    return 1
}

# Cleanup temp files on exit
cleanup_temp_files() {
    if [ -n "${TEMP_FILES:-}" ]; then
        for temp_file in $TEMP_FILES; do
            rm -f "$temp_file" 2>/dev/null || true
        done
        unset TEMP_FILES
    fi
}

# Register temp file for cleanup
register_temp_file() {
    local temp_file="$1"
    TEMP_FILES="${TEMP_FILES:-} $temp_file"
}

# Create safe temp file
make_temp_file() {
    local temp_file=$(mktemp)
    register_temp_file "$temp_file"
    echo "$temp_file"
}

# Set up cleanup trap
setup_cleanup_trap() {
    trap 'cleanup_temp_files' EXIT INT TERM
}

# Version comparison (simple)
version_compare() {
    local version1="$1"
    local version2="$2"

    if [ "$version1" = "$version2" ]; then
        echo "0"
    elif printf '%s\n%s\n' "$version1" "$version2" | sort -V | head -1 | grep -q "^$version1$"; then
        echo "-1"
    else
        echo "1"
    fi
}

# Initialize common library
init_common_lib() {
    setup_cleanup_trap

    # Set default verbose mode if not specified
    VERBOSE="${VERBOSE:-false}"

    # Ensure required directories exist
    ensure_directory "$WORK_DIR"
}

# Auto-initialize when sourced
init_common_lib
