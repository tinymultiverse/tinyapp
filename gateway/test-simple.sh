#!/bin/bash

# Simple TinyApp Gateway OIDC Test Script
# Uses a lightweight mock OIDC provider for easy testing

set -e

echo "🚀 TinyApp Gateway OIDC Simple Test"
echo "===================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    docker-compose -f docker-compose.simple.yml down -v 2>/dev/null || true
    if [ ! -z "$GATEWAY_PID" ]; then
        kill $GATEWAY_PID 2>/dev/null || true
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

echo "📦 Starting lightweight test services..."

# Start the simple mock OIDC provider and backend
docker-compose -f docker-compose.simple.yml up -d

echo "⏳ Waiting for services to be ready..."
sleep 8

# Wait for mock OIDC provider to be ready
echo "🔍 Checking mock OIDC provider..."
for i in {1..30}; do
    if curl -s http://localhost:9000/default/.well-known/openid_configuration > /dev/null 2>&1; then
        break
    fi
    echo "  Waiting for OIDC provider... ($i/30)"
    sleep 1
done

# Verify the provider is actually responding
if ! curl -s http://localhost:9000/default/.well-known/openid_configuration > /dev/null 2>&1; then
    echo "❌ Failed to connect to OIDC provider after 30 seconds"
    echo "Check if the container is running: docker ps"
    exit 1
fi

echo "✅ Mock OIDC provider is ready at http://localhost:9000"
echo "✅ Mock backend is ready at http://localhost:5001"

# Build the gateway
# echo "🔨 Building TinyApp Gateway..."
# go build -o gateway-test ./cmd/

# Load simple test environment variables
echo "⚙️  Loading simple test configuration..."
export $(grep -v '^#' .env.simple | xargs)

echo "🚀 Starting TinyApp Gateway..."
# ./gateway-test &
go run ../cmd/main.go gateway &
GATEWAY_PID=$!

# Wait a moment for the gateway to start
sleep 3

echo ""
echo "🎉 Simple Test Environment Ready!"
echo "================================="
echo ""
echo "📝 Test Instructions:"
echo "1. Open your browser and go to: http://localhost:8005"
echo "2. You'll be redirected to a simple login page"
echo "3. Click 'Sign in' with any username (it's a mock provider)"
echo "4. You'll be redirected back to see the test page"
echo "5. Check metrics at: http://localhost:9090/metrics"
echo ""
echo "🔗 URLs:"
echo "   - Gateway: http://localhost:8005"
echo "   - Mock OIDC: http://localhost:9000"
echo "   - Backend: http://localhost:5001"
echo "   - Metrics: http://localhost:9090/metrics"
echo ""
echo "📊 The mock provider will create a test user with:"
echo "   - Username: testuser"
echo "   - Email: test@example.com"
echo "   - Name: Test User"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for user to stop
wait $GATEWAY_PID