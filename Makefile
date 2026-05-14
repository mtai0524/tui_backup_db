.PHONY: build install clean help dev release

APP_NAME := bakdb
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LD_FLAGS := -ldflags "-X 'main.version=$(VERSION)' -X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'"

# Output directories
BUILD_DIR := ./build
RELEASE_DIR := ./releases

# Platform variables
LINUX_AMD64 := $(BUILD_DIR)/$(APP_NAME)-linux-amd64
LINUX_ARM64 := $(BUILD_DIR)/$(APP_NAME)-linux-arm64
MACOS_AMD64 := $(BUILD_DIR)/$(APP_NAME)-macos-amd64
MACOS_ARM64 := $(BUILD_DIR)/$(APP_NAME)-macos-arm64
WINDOWS_AMD64 := $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe

help:
	@echo "bakdb - Database Backup Manager"
	@echo ""
	@echo "Available commands:"
	@echo "  make build        - Build for current platform"
	@echo "  make dev          - Build and run in development mode"
	@echo "  make release      - Build for all platforms"
	@echo "  make install      - Build and install to /usr/local/bin"
	@echo "  make uninstall    - Remove from /usr/local/bin"
	@echo "  make clean        - Clean build artifacts"
	@echo ""
	@echo "Release targets:"
	@echo "  make linux-amd64  - Build for Linux x86_64"
	@echo "  make linux-arm64  - Build for Linux ARM64"
	@echo "  make macos-amd64  - Build for macOS Intel"
	@echo "  make macos-arm64  - Build for macOS Apple Silicon"
	@echo "  make windows      - Build for Windows x86_64"
	@echo ""

build: clean
	@echo "🔨 Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

dev: build
	@echo "▶️  Running in development mode..."
	./$(BUILD_DIR)/$(APP_NAME)

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(RELEASE_DIR)
	go clean

install: build
	@echo "📦 Installing $(APP_NAME) to /usr/local/bin..."
	@if [ ! -w /usr/local/bin ]; then \
		echo "❌ No write permission to /usr/local/bin"; \
		echo "   Try: sudo make install"; \
		exit 1; \
	fi
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	chmod +x /usr/local/bin/$(APP_NAME)
	@echo "✅ Installed to /usr/local/bin/$(APP_NAME)"
	@echo "   Run: bakdb"

uninstall:
	@echo "🗑️  Removing $(APP_NAME) from /usr/local/bin..."
	@if [ -f /usr/local/bin/$(APP_NAME) ]; then \
		if [ ! -w /usr/local/bin ]; then \
			echo "❌ No write permission to /usr/local/bin"; \
			echo "   Try: sudo make uninstall"; \
			exit 1; \
		fi; \
		rm /usr/local/bin/$(APP_NAME); \
		echo "✅ Uninstalled"; \
	else \
		echo "⚠️  $(APP_NAME) not found in /usr/local/bin"; \
	fi

# Release builds for all platforms
release: clean
	@echo "🎁 Building releases..."
	@mkdir -p $(RELEASE_DIR)
	@$(MAKE) linux-amd64
	@$(MAKE) linux-arm64
	@$(MAKE) macos-amd64
	@$(MAKE) macos-arm64
	@$(MAKE) windows
	@echo "✅ All releases built in $(RELEASE_DIR)/"
	@ls -lh $(RELEASE_DIR)/

linux-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o $(LINUX_AMD64)
	@echo "✅ Linux AMD64: $(LINUX_AMD64)"

linux-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LD_FLAGS) -o $(LINUX_ARM64)
	@echo "✅ Linux ARM64: $(LINUX_ARM64)"

macos-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LD_FLAGS) -o $(MACOS_AMD64)
	@echo "✅ macOS Intel: $(MACOS_AMD64)"

macos-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LD_FLAGS) -o $(MACOS_ARM64)
	@echo "✅ macOS Apple Silicon: $(MACOS_ARM64)"

windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o $(WINDOWS_AMD64)
	@echo "✅ Windows x86_64: $(WINDOWS_AMD64)"
