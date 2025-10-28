#!/bin/bash

# TinyApp Gateway OIDC Test Script
# This script helps you test the OIDC authentication functionality

set -e

echo "🚀 TinyApp Gateway OIDC Test Setup"
echo "=================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
    if [ ! -z "$GATEWAY_PID" ]; then
        kill $GATEWAY_PID 2>/dev/null || true
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

echo "📦 Starting test services..."

# Start the test OIDC provider and mock backend
docker-compose -f docker-compose.test.yml up -d

echo "⏳ Waiting for services to be ready..."
sleep 5

# Wait for OIDC provider to be ready
echo "🔍 Waiting for Keycloak to start (this may take a minute)..."
for i in {1..60}; do
    if curl -s http://localhost:9000/realms/tinyapp/.well-known/openid_configuration > /dev/null 2>&1; then
        break
    fi
    echo "  Waiting for Keycloak... ($i/60)"
    sleep 2
done

# Verify Keycloak is actually responding
if ! curl -s http://localhost:9000/realms/tinyapp/.well-known/openid_configuration > /dev/null 2>&1; then
    echo "❌ Failed to connect to Keycloak after 2 minutes"
    echo "Check if the container is running: docker ps"
    exit 1
fi

echo "✅ Keycloak OIDC provider is ready at http://localhost:9000"
echo "✅ Mock backend is ready at http://localhost:5001"

# Build the gateway
# echo "🔨 Building TinyApp Gateway..."
# go build -o gateway-test ./cmd/

# Load test environment variables
echo "⚙️  Loading test configuration..."
export $(grep -v '^#' .env.test | xargs)

echo "🚀 Starting TinyApp Gateway..."
go run ../cmd/main.go gateway &
GATEWAY_PID=$!

# Wait a moment for the gateway to start
sleep 2

echo ""
echo "🎉 Test Environment Ready!"
echo "========================="
echo ""
echo "📝 Test Instructions:"
echo "1. Open your browser and go to: http://localhost:8005"
echo "2. You should be redirected to the OIDC login page"
echo "3. Use one of these test accounts to log in:"
echo "   - Username: testuser, Password: password123"
echo "   - Username: alice, Password: alice123" 
echo "   - Username: bob, Password: bob123"
echo "4. After login, you should see the test backend page"
echo "5. Check metrics at: http://localhost:9090/metrics"
echo ""
echo "🔗 Useful URLs:"
echo "   - Gateway: http://localhost:8005"
echo "   - Keycloak Admin: http://localhost:9000 (admin/admin)"
echo "   - Backend: http://localhost:5001" 
echo "   - Metrics: http://localhost:9090/metrics"
echo "   - Login: http://localhost:8005/auth/login"
echo "   - Logout: http://localhost:8005/auth/logout"
echo ""
echo "📊 To view Keycloak logs:"
echo "   docker logs gateway_keycloak_1"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for user to stop
wait $GATEWAY_PID