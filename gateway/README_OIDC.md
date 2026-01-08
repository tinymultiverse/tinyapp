# OIDC Authentication for TinyApp Gateway

This document explains how to configure and use OIDC (OpenID Connect) authentication with the TinyApp Gateway.

## Overview

The TinyApp Gateway now supports OIDC authentication, allowing you to secure access to your applications using any OIDC-compliant identity provider such as:

- Google
- Microsoft Azure AD
- Auth0
- Keycloak
- Okta
- And many others

## Configuration

### Environment Variables

To enable OIDC authentication, set the following environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OIDC_ENABLED` | No | `false` | Enable/disable OIDC authentication |
| `OIDC_ISSUER_URL` | Yes* | - | The issuer URL of your OIDC provider (e.g., `https://accounts.google.com`) |
| `OIDC_CLIENT_ID` | Yes* | - | The client ID registered with your OIDC provider |
| `OIDC_CLIENT_SECRET` | Yes* | - | The client secret for your OIDC application |
| `OIDC_REDIRECT_URL` | Yes* | - | The callback URL (e.g., `http://localhost:8005/auth/callback`) |
| `OIDC_SCOPES` | No | `openid,profile,email` | Comma-separated list of OIDC scopes to request |

*Required when `OIDC_ENABLED=true`

### Example Configuration

```bash
# Enable OIDC authentication
export OIDC_ENABLED=true

# Google OIDC configuration example
export OIDC_ISSUER_URL=https://accounts.google.com
export OIDC_CLIENT_ID=your-client-id.apps.googleusercontent.com
export OIDC_CLIENT_SECRET=your-client-secret
export OIDC_REDIRECT_URL=http://localhost:8005/auth/callback
export OIDC_SCOPES=openid,profile,email

# Azure AD OIDC configuration example
export OIDC_ISSUER_URL=https://login.microsoftonline.com/your-tenant-id/v2.0
export OIDC_CLIENT_ID=your-application-id
export OIDC_CLIENT_SECRET=your-client-secret
export OIDC_REDIRECT_URL=http://localhost:8005/auth/callback
```

## Setting Up OIDC Providers

### Google

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth 2.0 Client IDs"
5. Set the application type to "Web application"
6. Add your redirect URI: `http://your-gateway-host:port/auth/callback`
7. Note the Client ID and Client Secret

### Microsoft Azure AD

1. Go to the [Azure Portal](https://portal.azure.com/)
2. Navigate to "Azure Active Directory" > "App registrations"
3. Click "New registration"
4. Set the redirect URI to `http://your-gateway-host:port/auth/callback`
5. Go to "Certificates & secrets" and create a new client secret
6. Note the Application (client) ID, Directory (tenant) ID, and the client secret value

### Auth0

1. Go to your [Auth0 Dashboard](https://manage.auth0.com/)
2. Navigate to "Applications" and create a new Regular Web Application
3. Configure the Allowed Callback URLs: `http://your-gateway-host:port/auth/callback`
4. Note the Domain, Client ID, and Client Secret

## Authentication Flow

When OIDC is enabled, the gateway implements the following authentication flow:

1. **Unauthenticated Request**: When a user accesses the gateway without authentication, they are redirected to the OIDC provider's login page
2. **Login**: The user logs in with their credentials at the OIDC provider
3. **Callback**: The OIDC provider redirects back to the gateway with an authorization code
4. **Token Exchange**: The gateway exchanges the authorization code for an ID token
5. **Session Creation**: The gateway creates a secure session cookie containing user information
6. **Access Granted**: Subsequent requests use the session cookie for authentication

## Authentication Endpoints

The gateway provides the following authentication endpoints:

- `/auth/login` - Initiates the OIDC login flow
- `/auth/callback` - Handles the OIDC callback (configured in your OIDC provider)
- `/auth/logout` - Logs out the user and clears the session

## User Information

When OIDC authentication is enabled, the gateway will:

- Track actual usernames in metrics instead of using "any-user"
- Use the following priority for username identification:
  1. `preferred_username` claim
  2. `email` claim
  3. `sub` (subject) claim

## Security Considerations

1. **HTTPS in Production**: Always use HTTPS in production environments to protect authentication cookies and tokens
2. **Secure Cookies**: Session cookies are marked as `HttpOnly` and `Secure` (when HTTPS is used)
3. **State Parameter**: The implementation uses a secure state parameter to prevent CSRF attacks
4. **Token Validation**: ID tokens are properly verified using the OIDC provider's public keys
5. **Session Expiration**: Sessions automatically expire based on the ID token expiration time

## Troubleshooting

### Common Issues

1. **"Authentication failed: state mismatch"**
   - Ensure cookies are enabled in your browser
   - Check that the redirect URL exactly matches what's configured in your OIDC provider

2. **"failed to create OIDC provider"**
   - Verify the `OIDC_ISSUER_URL` is correct and accessible
   - Check network connectivity to the OIDC provider

3. **"Authentication failed: token exchange failed"**
   - Verify the `OIDC_CLIENT_ID` and `OIDC_CLIENT_SECRET` are correct
   - Ensure the redirect URL is properly configured in your OIDC provider

### Debug Mode

Enable debug logging to troubleshoot authentication issues:

```bash
export LOG_LEVEL=debug
```

This will provide detailed logs about the authentication process.

## Disabling OIDC Authentication

To disable OIDC authentication and return to the previous behavior:

```bash
export OIDC_ENABLED=false
```

Or simply remove/unset the `OIDC_ENABLED` environment variable.