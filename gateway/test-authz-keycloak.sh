#!/bin/bash

# TinyApp Gateway OIDC Authorization Test Script (Keycloak)
# Tests role-based access control with Keycloak OIDC provider

set -e

echo "🚀 TinyApp Gateway OIDC Authorization Test (Keycloak)"
echo "====================================================="

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

echo "📦 Starting Keycloak and backend services..."

# Start Keycloak and mock backend
docker-compose -f docker-compose.test.yml up -d

echo "⏳ Waiting for Keycloak to start (this may take a minute)..."
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

# Load Keycloak authorization test environment variables (based on .env.test but with authz enabled)
echo "⚙️  Loading Keycloak authorization test configuration..."
export HTTP_PORT=8005
export PRIMARY_TARGET_PORT=5001
export URL_SUB_PATH=/
export METRICS_ENABLED=true
export METRICS_PORT=9090
export METRICS_PATH=/metrics
export OIDC_ENABLED=true
export OIDC_ISSUER_URL=http://localhost:9000/realms/tinyapp
export OIDC_CLIENT_ID=tinyapp-gateway
export OIDC_CLIENT_SECRET=gateway-secret-123
export OIDC_REDIRECT_URL=http://localhost:8005/auth/callback
export OIDC_SCOPES=openid,profile,email
export AUTHZ_ENABLED=true
export AUTHZ_REQUIRED_ROLES=user,member
export AUTHZ_REQUIRED_SCOPES=read
export AUTHZ_ROLE_CLAIM=roles
export AUTHZ_SCOPE_CLAIM=scopes
export AUTHZ_ADMIN_ROLES=admin
export LOG_LEVEL=debug

echo "🚀 Starting TinyApp Gateway with Authorization..."
go run ../cmd/main.go gateway &
GATEWAY_PID=$!

# Wait a moment for the gateway to start
sleep 3

echo ""
echo "🎉 Keycloak OAuth Delegation Test Environment Ready!"
echo "===================================================="
echo ""
echo "📝 OAuth Delegation Test Instructions:"
echo "1. Open your browser and go to: http://localhost:8005"
echo "2. You'll be redirected to Keycloak login page"
echo "3. Use these test accounts:"
echo "   - testuser / password123 (scopes: read, write, admin, delete)"
echo "   - alice / alice123 (scopes: read, write)"  
echo "   - bob / bob123 (scopes: read)"
echo ""
echo "🔧 OAuth Delegation Endpoints to Test:"
echo "   - http://localhost:8005/api/user-profile (user profile from Keycloak)"
echo "   - http://localhost:8005/api/user-data (delegation demo)"
echo "   - http://localhost:8005/api/proxy/http://localhost:9000/realms/tinyapp/protocol/openid-connect/userinfo"
echo ""
echo "🔍 Your backend will receive these headers:"
echo "   - X-User-ID: <user-id>"
echo "   - X-User-Email: <user-email>"
echo "   - X-User-Roles: <user-roles>"
echo "   - X-User-Scopes: <user-scopes>"
echo "   - Authorization: Bearer <user-access-token>"
echo "   - testuser / password123 (roles: user, member, admin) ✅ Should work"
echo "   - alice / alice123 (roles: user, member) ✅ Should work"  
echo "   - bob / bob123 (roles: user) ❌ Should work (has required 'user' role)"
echo "4. Check metrics at: http://localhost:9090/metrics"
echo ""
echo "🔐 Authorization Configuration:"
echo "   - Required Roles: user, member (ANY of these)"
echo "   - Required Scopes: read"
echo "   - Admin Roles: admin (bypass all checks)"
echo ""
echo "🔗 URLs:"
echo "   - Gateway: http://localhost:8005"
echo "   - Keycloak: http://localhost:9000"
echo "   - Backend: http://localhost:5001"
echo "   - Metrics: http://localhost:9090/metrics"
echo ""
echo "🧪 To test authorization failures:"
echo "   1. Stop this script (Ctrl+C)"
echo "   2. Change AUTHZ_REQUIRED_ROLES=admin in this script"
echo "   3. Restart and try with alice or bob (should fail)"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for user to stop
wait $GATEWAY_PID