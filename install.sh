#!/bin/bash

# bakdb Installation Script
# Installs Database Backup Manager

set -e

APP_NAME="bakdb"
REPO_URL="https://github.com/mtai0524/tui_backup_db"
INSTALL_PATH="/usr/local/bin"
CONFIG_PATH="${HOME}/.bakdb"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

check_requirements() {
    print_header "Checking Requirements"

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        echo "Please install Go from: https://golang.org/doc/install"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go $GO_VERSION found"

    # Check if git is installed
    if ! command -v git &> /dev/null; then
        print_error "Git is not installed"
        exit 1
    fi
    print_success "Git found"
}

setup_build_environment() {
    print_header "Setting Up Build Environment"

    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    print_info "Using temporary directory: $TEMP_DIR"

    # Clone repository
    print_info "Cloning repository..."
    git clone "$REPO_URL" "$TEMP_DIR/$APP_NAME" 2>/dev/null || {
        print_error "Failed to clone repository"
        exit 1
    }

    cd "$TEMP_DIR/$APP_NAME"
    print_success "Repository cloned"
}

build_application() {
    print_header "Building Application"

    if [ ! -f "Makefile" ]; then
        print_error "Makefile not found"
        exit 1
    fi

    print_info "Building $APP_NAME..."
    make build || {
        print_error "Build failed"
        exit 1
    }

    print_success "Build successful"
}

install_application() {
    print_header "Installing Application"

    # Check write permission
    if [ ! -w "$INSTALL_PATH" ]; then
        print_warning "No write permission to $INSTALL_PATH"
        print_info "Attempting with sudo..."
        sudo cp "build/$APP_NAME" "$INSTALL_PATH/$APP_NAME"
        sudo chmod +x "$INSTALL_PATH/$APP_NAME"
    else
        cp "build/$APP_NAME" "$INSTALL_PATH/$APP_NAME"
        chmod +x "$INSTALL_PATH/$APP_NAME"
    fi

    print_success "Installed to $INSTALL_PATH/$APP_NAME"
}

setup_config_directory() {
    print_header "Setting Up Configuration"

    if [ ! -d "$CONFIG_PATH" ]; then
        mkdir -p "$CONFIG_PATH"
        print_success "Created config directory: $CONFIG_PATH"
    fi

    # Copy .env.example if it exists
    if [ -f ".env.example" ]; then
        if [ ! -f "$CONFIG_PATH/.env" ]; then
            cp ".env.example" "$CONFIG_PATH/.env.example"
            print_info "Copied .env.example to $CONFIG_PATH/"
            print_info "Create $CONFIG_PATH/.env to configure defaults"
        fi
    fi
}

verify_installation() {
    print_header "Verifying Installation"

    if ! command -v $APP_NAME &> /dev/null; then
        print_error "Installation verification failed"
        exit 1
    fi

    VERSION=$($APP_NAME --version 2>/dev/null || echo "unknown")
    print_success "$APP_NAME installed successfully"
    print_info "Version: $VERSION"
}

cleanup() {
    print_header "Cleaning Up"

    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
        print_success "Cleaned up temporary files"
    fi
}

show_next_steps() {
    print_header "Installation Complete! 🎉"

    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Configure your defaults (optional):"
    echo "   vim $CONFIG_PATH/.env"
    echo ""
    echo "2. Run bakdb:"
    echo "   bakdb"
    echo ""
    echo "3. First use:"
    echo "   - Select database type (MySQL, PostgreSQL, SQL Server)"
    echo "   - Enter connection details"
    echo "   - Start backup"
    echo "   - Optionally send via email"
    echo ""
    echo "Documentation:"
    echo "   $REPO_URL/blob/main/README.md"
    echo ""
}

main() {
    print_header "bakdb - Database Backup Manager Installer"

    check_requirements
    setup_build_environment
    build_application
    install_application
    setup_config_directory
    verify_installation
    cleanup
    show_next_steps
}

# Trap errors
trap 'print_error "Installation failed"; exit 1' ERR

# Run installation
main
