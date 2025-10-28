/*
Copyright 2024 BlackRock, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/tinymultiverse/tinyapp/gateway/internal"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	StateTokenName   = "oidc_state"
	SessionTokenName = "oidc_session"
	LoginPath        = "/auth/login"
	CallbackPath     = "/auth/callback"
	LogoutPath       = "/auth/logout"
)

type OIDCAuth struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	scopes       []string
	envVars      internal.EnvVars
}

type UserInfo struct {
	Sub               string   `json:"sub"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	PreferredUsername string   `json:"preferred_username"`
	Roles             []string `json:"roles"`  // User roles for authorization
	Scope             string   `json:"scope"`  // OAuth scopes
	Groups            []string `json:"groups"` // User groups (alternative to roles)
}

type SessionData struct {
	UserInfo  UserInfo  `json:"user_info"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewOIDCAuth(envVars internal.EnvVars) (*OIDCAuth, error) {
	if !envVars.OIDCEnabled {
		return nil, nil
	}

	if envVars.OIDCIssuerURL == "" || envVars.OIDCClientID == "" ||
		envVars.OIDCClientSecret == "" || envVars.OIDCRedirectURL == "" {
		return nil, fmt.Errorf("OIDC_ISSUER_URL, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL must be set when OIDC_ENABLED is true")
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, envVars.OIDCIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	scopes := strings.Split(envVars.OIDCScopes, ",")
	for i, scope := range scopes {
		scopes[i] = strings.TrimSpace(scope)
	}

	oauth2Config := oauth2.Config{
		ClientID:     envVars.OIDCClientID,
		ClientSecret: envVars.OIDCClientSecret,
		RedirectURL:  envVars.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: envVars.OIDCClientID})

	return &OIDCAuth{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		scopes:       scopes,
		envVars:      envVars,
	}, nil
}

func (o *OIDCAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth endpoints
		switch r.URL.Path {
		case LoginPath:
			o.handleLogin(w, r)
			return
		case CallbackPath:
			o.handleCallback(w, r)
			return
		case LogoutPath:
			o.handleLogout(w, r)
			return
		}

		// Check if user is authenticated
		if !o.isAuthenticated(r) {
			// Redirect to login
			state := generateRandomString(32)
			http.SetCookie(w, &http.Cookie{
				Name:     StateTokenName,
				Value:    state,
				Path:     "/",
				HttpOnly: true,
				Secure:   r.TLS != nil,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   300, // 5 minutes
			})

			authURL := o.oauth2Config.AuthCodeURL(state)
			zap.S().Debugw("redirecting to OIDC provider", "url", authURL)
			http.Redirect(w, r, authURL, http.StatusFound)
			return
		}

		// User is authenticated, proceed to next handler
		next.ServeHTTP(w, r)
	})
}

func (o *OIDCAuth) handleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateRandomString(32)
	http.SetCookie(w, &http.Cookie{
		Name:     StateTokenName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
	})

	authURL := o.oauth2Config.AuthCodeURL(state)
	zap.S().Infow("initiating OIDC login", "redirect_url", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (o *OIDCAuth) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter
	stateCookie, err := r.Cookie(StateTokenName)
	if err != nil {
		zap.S().Errorw("state cookie not found", "error", err)
		http.Error(w, "Authentication failed: state not found", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != stateCookie.Value {
		zap.S().Error("state parameter mismatch")
		http.Error(w, "Authentication failed: state mismatch", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     StateTokenName,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Exchange authorization code for tokens
	oauth2Token, err := o.oauth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		zap.S().Errorw("failed to exchange code for token", "error", err)
		http.Error(w, "Authentication failed: token exchange failed", http.StatusInternalServerError)
		return
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		zap.S().Error("no id_token field in oauth2 token")
		http.Error(w, "Authentication failed: no ID token", http.StatusInternalServerError)
		return
	}

	// Verify ID token
	idToken, err := o.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		zap.S().Errorw("failed to verify ID token", "error", err)
		http.Error(w, "Authentication failed: token verification failed", http.StatusInternalServerError)
		return
	}

	// Extract user info from ID token
	var userInfo UserInfo
	if err := idToken.Claims(&userInfo); err != nil {
		zap.S().Errorw("failed to extract user info from ID token", "error", err)
		http.Error(w, "Authentication failed: failed to extract user info", http.StatusInternalServerError)
		return
	}

	// Extract additional claims for authorization
	var allClaims map[string]interface{}
	if err := idToken.Claims(&allClaims); err != nil {
		zap.S().Warnw("failed to extract additional claims", "error", err)
	} else {
		zap.S().Debugw("all JWT claims", "claims", allClaims)
		
		// Extract roles from custom claim
		roleClaim := o.envVars.AuthzRoleClaim
		zap.S().Debugw("looking for role claim", "claim_name", roleClaim)
		if roleValue, exists := allClaims[roleClaim]; exists {
			zap.S().Debugw("found role claim", "claim_name", roleClaim, "value", roleValue, "type", fmt.Sprintf("%T", roleValue))
			if roles, ok := o.extractStringSlice(roleValue); ok {
				userInfo.Roles = roles
				zap.S().Debugw("extracted roles", "roles", roles)
			} else {
				zap.S().Warnw("failed to extract roles from claim", "claim_name", roleClaim, "value", roleValue)
			}
		} else {
			zap.S().Warnw("role claim not found in token", "claim_name", roleClaim, "available_claims", func() []string {
				keys := make([]string, 0, len(allClaims))
				for k := range allClaims {
					keys = append(keys, k)
				}
				return keys
			}())
		}

		// Extract groups (alternative to roles)
		if groupClaim, exists := allClaims["groups"]; exists {
			if groups, ok := o.extractStringSlice(groupClaim); ok {
				userInfo.Groups = groups
			}
		}

		// Extract scope claim for OAuth scopes
		if scopeClaim, exists := allClaims[o.envVars.AuthzScopeClaim]; exists {
			if scope, ok := scopeClaim.(string); ok {
				userInfo.Scope = scope
			}
		}
	}

	zap.S().Infow("user authenticated successfully",
		"user", userInfo.Sub,
		"email", userInfo.Email,
		"roles", userInfo.Roles,
		"groups", userInfo.Groups,
		"scope", userInfo.Scope)

	// Create session
	sessionData := SessionData{
		UserInfo:  userInfo,
		ExpiresAt: idToken.Expiry,
	}

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		zap.S().Errorw("failed to marshal session data", "error", err)
		http.Error(w, "Authentication failed: session creation failed", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionTokenName,
		Value:    base64.StdEncoding.EncodeToString(sessionJSON),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  idToken.Expiry,
	})

	zap.S().Infow("user authenticated successfully", "user", userInfo.Sub, "email", userInfo.Email)

	// Redirect to root or intended page
	redirectURL := "/"
	if returnTo := r.URL.Query().Get("return_to"); returnTo != "" {
		redirectURL = returnTo
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (o *OIDCAuth) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionTokenName,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	zap.S().Info("user logged out")

	// Redirect to root or OIDC provider logout if supported
	http.Redirect(w, r, "/", http.StatusFound)
}

func (o *OIDCAuth) isAuthenticated(r *http.Request) bool {
	sessionCookie, err := r.Cookie(SessionTokenName)
	if err != nil {
		return false
	}

	sessionJSON, err := base64.StdEncoding.DecodeString(sessionCookie.Value)
	if err != nil {
		zap.S().Debugw("failed to decode session cookie", "error", err)
		return false
	}

	var sessionData SessionData
	if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
		zap.S().Debugw("failed to unmarshal session data", "error", err)
		return false
	}

	// Check if session has expired
	if time.Now().After(sessionData.ExpiresAt) {
		zap.S().Debug("session has expired")
		return false
	}

	return true
}

func (o *OIDCAuth) GetUserInfo(r *http.Request) (*UserInfo, error) {
	sessionCookie, err := r.Cookie(SessionTokenName)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	sessionJSON, err := base64.StdEncoding.DecodeString(sessionCookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &sessionData.UserInfo, nil
}

// CheckAuthorization verifies if the user has the required roles and scopes
func (o *OIDCAuth) CheckAuthorization(r *http.Request) error {
	if !o.envVars.AuthzEnabled {
		return nil // Authorization disabled, allow access
	}

	userInfo, err := o.GetUserInfo(r)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	// Check if user has admin role (bypasses all other checks)
	if o.envVars.AuthzAdminRoles != "" {
		adminRoles := o.parseCommaSeparated(o.envVars.AuthzAdminRoles)
		if o.hasAnyRole(userInfo, adminRoles) {
			zap.S().Debugw("user has admin role, bypassing authorization checks", "user", userInfo.Sub)
			return nil
		}
	}

	// Check required roles
	if o.envVars.AuthzRequiredRoles != "" {
		requiredRoles := o.parseCommaSeparated(o.envVars.AuthzRequiredRoles)
		if !o.hasAnyRole(userInfo, requiredRoles) {
			zap.S().Warnw("user lacks required roles", "user", userInfo.Sub, "required", requiredRoles, "user_roles", userInfo.Roles)
			return fmt.Errorf("access denied: user lacks required roles %v", requiredRoles)
		}
	}

	// Check required OAuth scopes
	if o.envVars.AuthzRequiredScopes != "" {
		requiredScopes := o.parseCommaSeparated(o.envVars.AuthzRequiredScopes)
		userScopes := o.parseCommaSeparated(userInfo.Scope)
		if !o.hasAnyScope(userScopes, requiredScopes) {
			zap.S().Warnw("user lacks required scopes", "user", userInfo.Sub, "required", requiredScopes, "user_scopes", userScopes)
			return fmt.Errorf("access denied: user lacks required scopes %v", requiredScopes)
		}
	}

	zap.S().Debugw("authorization check passed", "user", userInfo.Sub)
	return nil
}

// hasAnyRole checks if the user has any of the required roles
func (o *OIDCAuth) hasAnyRole(userInfo *UserInfo, requiredRoles []string) bool {
	for _, required := range requiredRoles {
		for _, userRole := range userInfo.Roles {
			if strings.TrimSpace(userRole) == strings.TrimSpace(required) {
				return true
			}
		}
		// Also check groups as roles (some OIDC providers use groups instead of roles)
		for _, userGroup := range userInfo.Groups {
			if strings.TrimSpace(userGroup) == strings.TrimSpace(required) {
				return true
			}
		}
	}
	return false
}

// hasAnyScope checks if the user has any of the required OAuth scopes
func (o *OIDCAuth) hasAnyScope(userScopes []string, requiredScopes []string) bool {
	for _, required := range requiredScopes {
		for _, userScope := range userScopes {
			if strings.TrimSpace(userScope) == strings.TrimSpace(required) {
				return true
			}
		}
	}
	return false
}

// parseCommaSeparated splits a comma-separated string into a slice of trimmed strings
func (o *OIDCAuth) parseCommaSeparated(input string) []string {
	if input == "" {
		return []string{}
	}
	parts := strings.Split(input, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
}

// extractStringSlice converts various claim formats to a string slice
func (o *OIDCAuth) extractStringSlice(claim interface{}) ([]string, bool) {
	switch v := claim.(type) {
	case []string:
		return v, true
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result, len(result) > 0
	case string:
		// Handle comma-separated string
		return o.parseCommaSeparated(v), true
	default:
		return nil, false
	}
}

// AuthorizationMiddleware adds authorization checks to the authentication middleware
func (o *OIDCAuth) AuthorizationMiddleware(next http.Handler) http.Handler {
	return o.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization after authentication
		if err := o.CheckAuthorization(r); err != nil {
			zap.S().Warnw("authorization failed", "error", err, "path", r.URL.Path)
			http.Error(w, "Access Forbidden: "+err.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
