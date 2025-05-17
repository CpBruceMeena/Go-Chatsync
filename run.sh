#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting ChatSync setup...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed${NC}"
    exit 1
fi

# Check if Go dependencies need to be installed
if [ ! -d "vendor" ]; then
    echo -e "${YELLOW}Installing Go dependencies...${NC}"
    go mod tidy
else
    echo -e "${GREEN}Go dependencies already installed${NC}"
fi

# Check if frontend dependencies need to be installed
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    cd frontend
    npm install
    cd ..
else
    echo -e "${GREEN}Frontend dependencies already installed${NC}"
fi

# Check if frontend build exists and is up to date
if [ ! -d "frontend/build" ] || [ "$(find frontend/src -newer frontend/build)" ]; then
    echo -e "${YELLOW}Building React app...${NC}"
    cd frontend
    npm run build
    cd ..
else
    echo -e "${GREEN}Frontend build is up to date${NC}"
fi

# Create static/build directory if it doesn't exist
mkdir -p static/build

# Copy build files to static directory
echo -e "${YELLOW}Copying build files to static directory...${NC}"
cp -r frontend/build/* static/build/

# Build the Go binary
echo -e "${YELLOW}Building Go binary...${NC}"
go build -o chatsync

# Check if build was successful
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
    echo -e "${YELLOW}Starting server...${NC}"
    echo -e "${GREEN}Server is running at http://localhost:8080${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
    ./chatsync
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi 