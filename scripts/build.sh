#!/bin/bash

# Build the React app
cd frontend
npm run build

# Create the static/build directory if it doesn't exist
cd ..
mkdir -p static/build

# Copy the build files to static/build
cp -r frontend/build/* static/build/

# Build the Go binary
go build -o chatsync

echo "Build completed successfully!" 