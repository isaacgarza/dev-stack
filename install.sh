#!/bin/bash
set -e

# dev-stack installation script
REPO="isaacgarza/dev-stack"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="dev-stack"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          echo -e "${RED}Unsupported OS: $(uname -s)${NC}" >&2; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)          echo -e "${RED}Unsupported architecture: $(uname -m)${NC}" >&2; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install_dev_stack() {
    local platform version download_url binary_name
    
    platform=$(detect_platform)
    version=$(get_latest_version)
    
    if [ -z "$version" ]; then
        echo -e "${RED}Failed to get latest version${NC}" >&2
        exit 1
    fi
    
    echo -e "${BLUE}Installing dev-stack ${version} for ${platform}...${NC}"
    
    # Determine binary name based on platform
    case "$platform" in
        windows-*) binary_name="${BINARY_NAME}.exe" ;;
        *)         binary_name="${BINARY_NAME}" ;;
    esac
    
    download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}"
    
    # Create temp directory
    temp_dir=$(mktemp -d)
    temp_file="${temp_dir}/${binary_name}"
    
    # Download binary
    echo -e "${YELLOW}Downloading from ${download_url}...${NC}"
    if ! curl -L -o "$temp_file" "$download_url"; then
        echo -e "${RED}Failed to download binary${NC}" >&2
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Make executable
    chmod +x "$temp_file"
    
    # Install to system
    if [ -w "$INSTALL_DIR" ]; then
        mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo -e "${YELLOW}Installing to ${INSTALL_DIR} (requires sudo)...${NC}"
        sudo mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    echo -e "${GREEN}✅ dev-stack installed successfully!${NC}"
    echo -e "${BLUE}Run 'dev-stack --help' to get started${NC}"
}

# Main
main() {
    echo -e "${BLUE}🚀 dev-stack installer${NC}"
    echo
    
    # Check dependencies
    for cmd in curl; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            echo -e "${RED}Required command not found: $cmd${NC}" >&2
            exit 1
        fi
    done
    
    install_dev_stack
}

main "$@"
