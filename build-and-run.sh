#!/bin/bash

# Burma 2D App Config Server - Build & Run Script

set -e

echo "🏗️  Building Burma 2D App Config Server..."

cd "$(dirname "$0")"

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Build for current platform
echo "🔨 Building executable..."
go build -o appconfig-server main.go

echo "✅ Build complete!"
echo ""
echo "🚀 Starting server on port 8080..."
echo "📡 Main endpoint: http://localhost:8080/api/burma2d/config"
echo "🏥 Health check:  http://localhost:8080/health"
echo ""
echo "Press Ctrl+C to stop the server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Run the server
./appconfig-server
