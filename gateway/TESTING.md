# OIDC Testing Guide

## Quick Test Options

### Option 1: Super Simple Test (Recommended for First Try)

The easiest way to test with a lightweight mock provider:

```bash
cd gateway
./test-simple.sh
```

This uses a simple mock OIDC provider that requires no configuration and automatically logs you in.

### Option 2: Full Keycloak Test (More Realistic)

For a more complete OIDC experience with real login forms:

```bash
cd gateway
./test-oidc.sh
```

This sets up a full Keycloak server with proper login forms and user management.

### 2. Manual Test Setup

If you prefer to set things up manually:

```bash
# 1. Start test services
cd gateway
docker-compose -f docker-compose.test.yml up -d

# 2. Load test environment
export $(grep -v '^#' .env.test | xargs)

# 3. Build and run gateway
go build -o gateway-test ./cmd/
./gateway-test
```

### 3. Test Credentials

Use these test accounts:
- **Username**: `testuser`, **Password**: `password123`
- **Username**: `alice`, **Password**: `alice123`  
- **Username**: `bob`, **Password**: `bob123`

### 4. Test Flow

1. Visit http://localhost:8005
2. You'll be redirected to http://localhost:9000 (OIDC provider)
3. Log in with test credentials
4. You'll be redirected back and see the test page
5. Check metrics at http://localhost:9090/metrics for `username_counter`

## Testing with Real OIDC Providers

### Google (Free)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project and enable Google+ API
3. Create OAuth 2.0 credentials:
   - Application type: Web application
   - Authorized redirect URI: `http://localhost:8005/auth/callback`
4. Use these environment variables:

```bash
export OIDC_ENABLED=true
export OIDC_ISSUER_URL=https://accounts.google.com
export OIDC_CLIENT_ID=your-client-id.apps.googleusercontent.com
export OIDC_CLIENT_SECRET=your-client-secret
export OIDC_REDIRECT_URL=http://localhost:8005/auth/callback
```

### Auth0 (Free Tier)

1. Sign up at [Auth0](https://auth0.com/)
2. Create a Regular Web Application
3. Configure:
   - Allowed Callback URLs: `http://localhost:8005/auth/callback`
   - Allowed Logout URLs: `http://localhost:8005`
4. Use these environment variables:

```bash
export OIDC_ENABLED=true
export OIDC_ISSUER_URL=https://your-domain.auth0.com/
export OIDC_CLIENT_ID=your-client-id
export OIDC_CLIENT_SECRET=your-client-secret
export OIDC_REDIRECT_URL=http://localhost:8005/auth/callback
```

## Testing Without OIDC

To test that the gateway still works without OIDC:

```bash
export OIDC_ENABLED=false
# or just unset OIDC_ENABLED
go run ./cmd/start.go
```

Visit http://localhost:8005 - you should access the backend directly without authentication.

## Troubleshooting

### Common Issues

1. **Port conflicts (5000, 9000, 8005, 9090)**
   ```bash
   # Check if ports are in use
   lsof -i :5000  # or :5001, :8005, :9000, :9090
   
   # Kill processes using the ports if needed
   sudo kill -9 $(lsof -t -i:5000)
   
   # On macOS, port 5000 might be used by AirPlay
   # You can disable it in System Preferences > Sharing > AirPlay Receiver
   ```

2. **Connection refused to OIDC provider**
   ```bash
   # Check if provider is running
   curl http://localhost:9000/.well-known/openid_configuration
   ```

3. **State mismatch errors**
   - Clear browser cookies
   - Make sure redirect URL matches exactly

3. **Token verification failed**
   - Check client ID/secret are correct
   - Verify issuer URL is accessible

### Debug Mode

Enable detailed logging:

```bash
export LOG_LEVEL=debug
```

### Check Metrics

Visit http://localhost:9090/metrics and look for:
```
username_counter{username="testuser"} 1
username_counter{username="alice"} 2
```

### Manual API Testing

Test endpoints directly:

```bash
# Should redirect to OIDC provider
curl -v http://localhost:8005/

# Login endpoint
curl -v http://localhost:8005/auth/login

# Health check
curl -v http://localhost:8005/auth/callback
```

## Cleanup

Stop all test services:

```bash
docker-compose -f docker-compose.test.yml down -v
```