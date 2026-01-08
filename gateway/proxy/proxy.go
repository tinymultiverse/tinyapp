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

package proxy

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/tinymultiverse/tinyapp/gateway/auth"
	"github.com/tinymultiverse/tinyapp/gateway/internal"
	"github.com/tinymultiverse/tinyapp/gateway/util/metrics"
	globalutil "github.com/tinymultiverse/tinyapp/util"
	"go.uber.org/zap"
)

type proxyServerConfig struct {
	Proxy                  *httputil.ReverseProxy
	SecondaryProxy         *httputil.ReverseProxy
	SecondaryTargetPattern string
	URLSubPath             string
	OIDCAuth               *auth.OIDCAuth
}

func NewProxyServerConfig(envVars internal.EnvVars) (*proxyServerConfig, error) {
	targetURL, err := url.Parse("http://127.0.0.1:" + envVars.PrimaryTargetPort)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	var secondaryProxy *httputil.ReverseProxy
	if envVars.SecondaryTargetPattern != "" && envVars.SecondaryTargetPort != "" {
		secondaryTargetUrl := &url.URL{
			Scheme: "http",
			Host:   "127.0.0.1:" + envVars.SecondaryTargetPort,
		}
		secondaryProxy = httputil.NewSingleHostReverseProxy(secondaryTargetUrl)
	}

	// Initialize OIDC authentication if enabled
	oidcAuth, err := auth.NewOIDCAuth(envVars)
	if err != nil {
		return nil, err
	}

	return &proxyServerConfig{
		Proxy:                  proxy,
		SecondaryProxy:         secondaryProxy,
		SecondaryTargetPattern: envVars.SecondaryTargetPattern,
		URLSubPath:             envVars.URLSubPath,
		OIDCAuth:               oidcAuth,
	}, nil
}

func (p *proxyServerConfig) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	zap.S().Debugw("got a request", "host", req.Host, "method", req.Method, "requestURL", req.URL.String())

	// Handle OAuth delegation routes
	if p.OIDCAuth != nil && strings.HasPrefix(req.URL.Path, "/api/") {
		p.handleDelegatedRequest(res, req)
		return
	}

	if p.SecondaryProxy != nil && strings.Contains(req.URL.Path, p.SecondaryTargetPattern) {
		zap.S().Debugw("routing to secondary proxy", "path", req.URL.Path)
		p.enhanceRequestWithUserContext(req)
		p.SecondaryProxy.ServeHTTP(res, req)
		return
	}

	// Only increment user count if the request URL is app homepage
	if path.Clean(req.URL.Path) == path.Clean(p.URLSubPath) {
		zap.S().Info("Incrementing user count")

		// Get actual username from OIDC if authentication is enabled
		username := globalutil.AnyUserName
		if p.OIDCAuth != nil {
			if userInfo, err := p.OIDCAuth.GetUserInfo(req); err == nil {
				// Use preferred username if available, otherwise use email or sub
				if userInfo.PreferredUsername != "" {
					username = userInfo.PreferredUsername
				} else if userInfo.Email != "" {
					username = userInfo.Email
				} else {
					username = userInfo.Sub
				}
			}
		}

		metrics.UsernameCounter.WithLabelValues(username).Inc()
	}

	// Enhance primary proxy requests with user context
	p.enhanceRequestWithUserContext(req)
	p.Proxy.ServeHTTP(res, req)
}

// handleDelegatedRequest handles API requests that need OAuth delegation
func (p *proxyServerConfig) handleDelegatedRequest(res http.ResponseWriter, req *http.Request) {
	if p.OIDCAuth == nil {
		http.Error(res, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	// Handle OAuth delegation example endpoints
	switch {
	case strings.HasPrefix(req.URL.Path, "/api/user-profile") || strings.HasPrefix(req.URL.Path, "/api/user-data"):
		// Use the OAuth example handler
		p.OIDCAuth.OAuthDelegationHandler().ServeHTTP(res, req)
		return
	case strings.HasPrefix(req.URL.Path, "/api/proxy/"):
		// Generic API proxy with user's credentials
		p.handleGenericAPIProxy(res, req)
		return
	default:
		// Forward to backend with user's access token
		p.enhanceRequestWithUserContext(req)
		p.Proxy.ServeHTTP(res, req)
	}
}

// handleGenericAPIProxy forwards requests to external APIs using user's access token
func (p *proxyServerConfig) handleGenericAPIProxy(res http.ResponseWriter, req *http.Request) {
	// Extract the target URL from the path: /api/proxy/https://api.example.com/endpoint
	targetPath := strings.TrimPrefix(req.URL.Path, "/api/proxy/")
	if targetPath == "" {
		http.Error(res, "Missing target URL in path", http.StatusBadRequest)
		return
	}

	// Make the API call on behalf of the user
	body := make([]byte, 0)
	if req.Body != nil {
		defer req.Body.Close()
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			zap.S().Errorw("failed to read request body", "error", err)
		}
	}

	resp, err := p.OIDCAuth.MakeAPICall(req, req.Method, targetPath, body)
	if err != nil {
		zap.S().Errorw("failed to make delegated API call", "error", err, "target", targetPath)
		http.Error(res, "Failed to call external API", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			res.Header().Add(key, value)
		}
	}

	// Copy status code and body
	res.WriteHeader(resp.StatusCode)
	io.Copy(res, resp.Body)

	zap.S().Debugw("successfully proxied delegated API call", "target", targetPath, "status", resp.StatusCode)
}

// enhanceRequestWithUserContext adds user context headers to the request
func (p *proxyServerConfig) enhanceRequestWithUserContext(req *http.Request) {
	if p.OIDCAuth == nil {
		return
	}

	userInfo, err := p.OIDCAuth.GetUserInfo(req)
	if err != nil {
		// User not authenticated, continue without enhancement
		return
	}

	// Add user context headers for the backend
	req.Header.Set("X-User-ID", userInfo.Sub)
	req.Header.Set("X-User-Email", userInfo.Email)
	req.Header.Set("X-User-Name", userInfo.Name)
	req.Header.Set("X-User-Username", userInfo.PreferredUsername)

	// Add roles and scopes if available
	if len(userInfo.Roles) > 0 {
		req.Header.Set("X-User-Roles", strings.Join(userInfo.Roles, ","))
	}
	if len(userInfo.Scopes) > 0 {
		req.Header.Set("X-User-Scopes", strings.Join(userInfo.Scopes, ","))
	}

	// Add user's access token as Authorization header (optional - depends on backend needs)
	if accessToken, err := p.OIDCAuth.GetUserAccessToken(req); err == nil {
		req.Header.Set("Authorization", "Bearer "+accessToken)
		zap.S().Debugw("enhanced request with user context", "user", userInfo.Sub, "has_token", true)
	} else {
		zap.S().Debugw("enhanced request with user context", "user", userInfo.Sub, "has_token", false, "token_error", err)
	}
}
