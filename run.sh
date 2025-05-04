#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_message() {
    echo -e "${BLUE}[ChatSync]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[Success]${NC} $1"
}

print_error() {
    echo -e "${RED}[Error]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    print_message "Visit https://golang.org/doc/install for installation instructions."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.16"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    print_error "Go version $REQUIRED_VERSION or higher is required. Current version: $GO_VERSION"
    exit 1
fi

# Create necessary directories if they don't exist
mkdir -p static

# Check if static/index.html exists
if [ ! -f "static/index.html" ]; then
    print_error "static/index.html not found. Please ensure you have the correct project structure."
    exit 1
fi

# Install dependencies
print_message "Installing dependencies..."
go mod tidy
if [ $? -ne 0 ]; then
    print_error "Failed to install dependencies."
    exit 1
fi
print_success "Dependencies installed successfully."

# Build the project
print_message "Building the project..."
go build -o chatsync
if [ $? -ne 0 ]; then
    print_error "Failed to build the project."
    exit 1
fi
print_success "Project built successfully."

# Run the application
print_message "Starting ChatSync..."
print_message "The application will be available at http://localhost:8080"
print_message "Press Ctrl+C to stop the server."
./chatsync 